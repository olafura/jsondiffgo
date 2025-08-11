package jsondiffgo

import (
    "reflect"
    "slices"

    "github.com/kranfix/go_matchable"
)

type _MyerDiff struct{}
type MyerDiff matchable.Matcher[_MyerDiff]

type Equal matchable.ValuedMatchable[_MyerDiff, []any]
type Insert matchable.ValuedMatchable[_MyerDiff, []any]
type Delete matchable.ValuedMatchable[_MyerDiff, []any]

type Path struct {
	Index  int
	Oldseq []any
	Newseq []any
	Edits  []MyerDiff
}

type _Process struct{}
type Process matchable.Matcher[_Process]

type Done matchable.ValuedMatchable[_Process, []MyerDiff]
type Next matchable.ValuedMatchable[_Process, []Path]
type Continue matchable.ValuedMatchable[_Process, Path]

func Myers(oldseq, newseq []any) []MyerDiff {
	var edits []MyerDiff
	path := Path{Index: 0, Oldseq: oldseq, Newseq: newseq, Edits: edits}
	return find(0, len(oldseq)+len(newseq), []Path{path})
}

/*
  def diff(): DiffList =
    val path = Path(0, List.from(oldSeq), List.from(newSeq), List())
    find(0, oldSeq.length + newSeq.length, List(path))
*/

func find(envelope, max int, paths []Path) []MyerDiff {
    // avoid unused parameter warning while keeping signature aligned
    _ = max
    switch diag := eachDiagonal(-envelope, envelope, paths, []Path{}); v := diag.(type) {
    case Done:
        return compactReverse(v.Val, []MyerDiff{})
    case Next:
        return find(envelope+1, max, v.Val)
    }
    return []MyerDiff{}
}

/*
  @tailrec
  private def find(envelope: Int, max: Int, paths: PathList): DiffList =
    eachDiagonal(-envelope, envelope, paths, List()) match
      case Done(edits) =>
        compactReverse(edits, List())
      case Next(paths) =>
        find(envelope + 1, max, paths)
*/

func compactReverse(edits []MyerDiff, acc []MyerDiff) []MyerDiff {
    // Special-case: rearrange Equals, Insert, Equals for smaller diffs
    if len(acc) >= 3 {
        if e1, ok1 := acc[0].(Equal); ok1 {
            if ins, ok2 := acc[1].(Insert); ok2 {
                if e2, ok3 := acc[2].(Equal); ok3 {
                    // Case A: Equals(a), Insert(a), Equals(b) => Insert(a), Equals(a ++ b)
                    if reflect.DeepEqual(e1.Val, ins.Val) {
                        // Transform: Equals(a) :: Insert(a) :: Equals(b)  =>  Insert(a) :: Equals(a ++ b)
                        merged := append([]any{}, e1.Val...)
                        merged = slices.Concat(merged, e2.Val)
                        newAcc := append([]MyerDiff{Insert{Val: e1.Val}, Equal{Val: merged}}, acc[3:]...)
                        return compactReverse(edits, newAcc)
                    }
                    // Case B: Equals(x), Insert(y), Equals(z) with z starting with y
                    // => Equals(x ++ y), Insert(y), Equals(z.dropPrefix(y))
                    if hasPrefix(e2.Val, ins.Val) {
                        left := slices.Concat(append([]any{}, e1.Val...), ins.Val)
                        right := append([]any{}, e2.Val[len(ins.Val):]...)
                        newHead := []MyerDiff{Equal{Val: left}, Insert{Val: ins.Val}, Equal{Val: right}}
                        newAcc := append(newHead, acc[3:]...)
                        return compactReverse(edits, newAcc)
                    }
                }
            }
        }
    }
    if len(edits) == 0 {
        return acc
    }
    first := edits[0]
    rest := edits[1:]
    if len(acc) > 0 {
        firstAcc := acc[0]
        accRest := acc[1:]
        switch fV := first.(type) {
        case Equal:
            switch fAV := firstAcc.(type) {
            case Equal:
                return compactReverse(rest, append([]MyerDiff{Equal{Val: slices.Concat(fV.Val, fAV.Val)}}, accRest...))
            default:
                // Keep existing accumulator; just prepend current edit
                return compactReverse(rest, append([]MyerDiff{first}, acc...))
            }
        case Insert:
            switch fAV := firstAcc.(type) {
            case Insert:
                return compactReverse(rest, append([]MyerDiff{Insert{Val: slices.Concat(fV.Val, fAV.Val)}}, accRest...))
            default:
                // Keep existing accumulator; just prepend current edit
                return compactReverse(rest, append([]MyerDiff{first}, acc...))
            }
        case Delete:
            switch fAV := firstAcc.(type) {
            case Delete:
                return compactReverse(rest, append([]MyerDiff{Delete{Val: slices.Concat(fV.Val, fAV.Val)}}, accRest...))
            default:
                // Keep existing accumulator; just prepend current edit
                return compactReverse(rest, append([]MyerDiff{first}, acc...))
            }
        }
    } else {
        // When accumulator is empty, just push the first edit
        return compactReverse(rest, append([]MyerDiff{first}, acc...))
    }
    return acc
}

// hasPrefix reports whether seq starts with the full contents of prefix,
// comparing elements with deep equality.
func hasPrefix(seq, prefix []any) bool {
    if len(prefix) == 0 {
        return false
    }
    if len(seq) < len(prefix) {
        return false
    }
    for i := range prefix {
        if !reflect.DeepEqual(seq[i], prefix[i]) {
            return false
        }
    }
    return true
}

/*
  @tailrec
  private def compactReverse(edits: DiffList, acc: DiffList): DiffList =
    (edits, acc) match
      case (Nil, _) =>
        acc
      case (Equals(elem) :: rest, Equals(result) :: accRest) =>
        compactReverse(rest, Equals(elem ::: result) :: accRest)
      case (Insert(elem) :: rest, Insert(result) :: accRest) =>
        compactReverse(rest, Insert(elem ::: result) :: accRest)
      case (Delete(elem) :: rest, Delete(result) :: accRest) =>
        compactReverse(rest, Delete(elem ::: result) :: accRest)
      case (rest, Equals(elem1) :: Insert(elem2) :: Equals(other) :: accRest)
          if elem1 == elem2 =>
        compactReverse(
          rest,
          Insert(elem1) :: Equals(elem1 ::: other) :: accRest
        )
      case (Equals(elem) :: rest, accRest) =>
        compactReverse(rest, Equals(elem) :: accRest)
      case (Insert(elem) :: rest, accRest) =>
        compactReverse(rest, Insert(elem) :: accRest)
      case (Delete(elem) :: rest, accRest) =>
        compactReverse(rest, Delete(elem) :: accRest)
*/

// processedPath is an internal helper used by the diagonal sweep
type processedPath struct {
    path *Path
    rest []Path
}

func eachDiagonal(diagonal, limit int, paths, nextPaths []Path) Process {
    if diagonal > limit {
        // return next paths in reverse order
        out := append([]Path(nil), nextPaths...)
        slices.Reverse(out)
        return Next{Val: out}
    }

    pp := processPath(diagonal, limit, paths)

    // Match reference behavior: if there is no path, treat as Done([])
    if pp.path == nil {
        return Done{Val: []MyerDiff{}}
    }

    switch res := followSnake(*pp.path).(type) {
    case Continue:
        // proceed to next diagonal with the advanced path
        return eachDiagonal(diagonal+2, limit, pp.rest, append([]Path{res.Val}, nextPaths...))
    case Done:
        return Done{Val: res.Val}
    default:
        return Done{Val: []MyerDiff{}}
    }
}

func processPath(diagonal, limit int, paths []Path) processedPath {
    if diagonal == 0 && limit == 0 {
        if len(paths) == 1 {
            p := paths[0]
            return processedPath{path: &p, rest: []Path{}}
        }
        return processedPath{path: nil, rest: []Path{}}
    }

    if len(paths) == 0 {
        return processedPath{path: nil, rest: []Path{}}
    }

    // At the bottom-most diagonal: must move down
    if diagonal == -limit {
        p := moveDown(paths[0])
        return processedPath{path: &p, rest: paths}
    }

    // At the top-most diagonal: must move right
    if diagonal == limit && len(paths) == 1 {
        p := moveRight(paths[0])
        return processedPath{path: &p, rest: []Path{}}
    }

    if len(paths) >= 2 {
        p1 := paths[0]
        p2 := paths[1]
        rest := paths[2:]
        if p1.Index > p2.Index {
            p := moveRight(p1)
            return processedPath{path: &p, rest: append([]Path{p2}, rest...)}
        }
        p := moveDown(p2)
        return processedPath{path: &p, rest: append([]Path{p2}, rest...)}
    }

    return processedPath{path: nil, rest: []Path{}}
}

func moveRight(path Path) Path {
    if len(path.Newseq) > 0 {
        elem := path.Newseq[0]
        rest := path.Newseq[1:]
        edits := append([]MyerDiff{Insert{Val: []any{elem}}}, path.Edits...)
        return Path{Index: path.Index, Oldseq: path.Oldseq, Newseq: rest, Edits: edits}
    }
    return Path{Index: path.Index, Oldseq: path.Oldseq, Newseq: []any{}, Edits: path.Edits}
}

func moveDown(path Path) Path {
    if len(path.Oldseq) > 0 {
        elem := path.Oldseq[0]
        rest := path.Oldseq[1:]
        edits := append([]MyerDiff{Delete{Val: []any{elem}}}, path.Edits...)
        return Path{Index: path.Index + 1, Oldseq: rest, Newseq: path.Newseq, Edits: edits}
    }
    return Path{Index: path.Index + 1, Oldseq: []any{}, Newseq: path.Newseq, Edits: path.Edits}
}

func followSnake(path Path) Process {
    p := path
    for len(p.Oldseq) > 0 && len(p.Newseq) > 0 && reflect.DeepEqual(p.Oldseq[0], p.Newseq[0]) {
        elem := p.Oldseq[0]
        p = Path{
            Index:  p.Index + 1,
            Oldseq: p.Oldseq[1:],
            Newseq: p.Newseq[1:],
            Edits:  append([]MyerDiff{Equal{Val: []any{elem}}}, p.Edits...),
        }
    }

    if len(p.Oldseq) == 0 && len(p.Newseq) == 0 {
        return Done{Val: p.Edits}
    }
    return Continue{Val: p}
}

/*
  @tailrec
  private def eachDiagonal(
      diagonal: Int,
      limit: Int,
      paths: PathList,
      nextPaths: PathList
  ): Done | Next = (diagonal, limit) match
    case (diagonal, limit) if diagonal > limit =>
      Next(nextPaths.reverse)
    case _ =>
      val processedPath = processPath(diagonal, limit, paths)

      followSnake(processedPath.path) match
        case Continue(path) =>
          eachDiagonal(
            diagonal + 2,
            limit,
            processedPath.rest,
            path :: nextPaths
          )
        case Done(edits) =>
          Done(edits)

  private def processPath(
      diagonal: Int,
      limit: Int,
      paths: PathList
  ): ProcessedPath = (diagonal, limit, paths) match
    case (0, 0, path :: Nil) =>
      ProcessedPath(Some(path), List())
    case (0, 0, Nil) =>
      ProcessedPath(None, List())
    case (diagonal, limit, path :: _) if diagonal == -limit =>
      ProcessedPath(Some(moveDown(path)), paths)
    case (diagonal, limit, path :: Nil) if diagonal == limit =>
      ProcessedPath(Some(moveRight(path)), List())
    case (_, _, path1 :: path2 :: rest) if path1.index > path2.index =>
      ProcessedPath(Some(moveRight(path1)), path2 :: rest)
    case (_, _, path1 :: path2 :: rest) if path1.index <= path2.index =>
      ProcessedPath(Some(moveDown(path2)), path2 :: rest)
    case (_, _, _) =>
      ProcessedPath(None, List())

  private def moveRight(path: Path): Path = path match
    case Path(index, list1, elem :: Nil, edits) =>
      Path(index, list1, List(), Insert(List(elem)) :: edits)
    case Path(index, list1, elem :: rest, edits) =>
      Path(index, list1, rest, Insert(List(elem)) :: edits)
    case Path(index, list1, Nil, edits) =>
      Path(index, list1, List(), edits)

  private def moveDown(path: Path): Path = path match
    case Path(index, elem :: rest, list2, edits) =>
      Path(index + 1, rest, list2, Delete(List(elem)) :: edits)
    case Path(index, Nil, list2, edits) =>
      Path(index + 1, List(), list2, edits)

  @tailrec
  private def followSnake(path: Option[Path]): Done | Continue = path match
    case Some(Path(index, elem1 :: rest1, elem2 :: rest2, edits))
        if elem1 == elem2 =>
      followSnake(
        Some(Path(index + 1, rest1, rest2, Equals(List(elem1)) :: edits))
      )
    case Some(Path(_, Nil, Nil, edits)) =>
      Done(edits)
    case Some(other) =>
      Continue(other)
    case None =>
      Done(List())
*/
