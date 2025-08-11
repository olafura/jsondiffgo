package jsondiffgo_test

import (
    "reflect"
    "testing"
    jsondiffgo "github.com/jsondiffgo"
)

// helper to convert a slice of values to []any
func anySlice[T any](in []T) []any {
    out := make([]any, len(in))
    for i := range in {
        out[i] = any(in[i])
    }
    return out
}

// normalize converts diffs into a comparable representation.
type simpleDiff struct {
    kind string
    vals []any
}

func simplify(diffs []jsondiffgo.MyerDiff) []simpleDiff {
    out := make([]simpleDiff, 0, len(diffs))
    for _, d := range diffs {
        switch v := d.(type) {
        case jsondiffgo.Equal:
            out = append(out, simpleDiff{kind: "Equal", vals: v.Val})
        case jsondiffgo.Insert:
            out = append(out, simpleDiff{kind: "Insert", vals: v.Val})
        case jsondiffgo.Delete:
            out = append(out, simpleDiff{kind: "Delete", vals: v.Val})
        default:
            t := reflect.TypeOf(d)
            out = append(out, simpleDiff{kind: t.String(), vals: nil})
        }
    }
    return out
}

func TestMyers_EmptySequences(t *testing.T) {
    got := jsondiffgo.Myers([]any{}, []any{})
    if len(got) != 0 {
        t.Fatalf("expected no edits, got: %#v", got)
    }
}

func TestMyers_AllEqual(t *testing.T) {
    old := anySlice([]int{1, 2, 3})
    got := jsondiffgo.Myers(old, anySlice([]int{1, 2, 3}))

    want := []simpleDiff{
        {kind: "Equal", vals: anySlice([]int{1, 2, 3})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_InsertsOnly(t *testing.T) {
    got := jsondiffgo.Myers([]any{}, anySlice([]string{"a", "b", "c"}))
    want := []simpleDiff{
        {kind: "Insert", vals: anySlice([]string{"a", "b", "c"})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_DeletesOnly(t *testing.T) {
    got := jsondiffgo.Myers(anySlice([]int{9, 8, 7}), []any{})
    want := []simpleDiff{
        {kind: "Delete", vals: anySlice([]int{9, 8, 7})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_MixedChanges(t *testing.T) {
    old := anySlice([]int{1, 2, 3})
    neu := anySlice([]int{0, 1, 3, 4})
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Insert", vals: anySlice([]int{0})},
        {kind: "Equal", vals: anySlice([]int{1})},
        {kind: "Delete", vals: anySlice([]int{2})},
        {kind: "Equal", vals: anySlice([]int{3})},
        {kind: "Insert", vals: anySlice([]int{4})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_Grouping(t *testing.T) {
    old := anySlice([]string{"a", "b", "c"})
    neu := anySlice([]string{"a", "b", "d", "c"})
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Equal", vals: anySlice([]string{"a", "b"})},
        {kind: "Insert", vals: anySlice([]string{"d"})},
        {kind: "Equal", vals: anySlice([]string{"c"})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_Paper_MiddleInsert(t *testing.T) {
    old := anySlice([]int{1, 2, 3})
    neu := anySlice([]int{1, 4, 2, 3})
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Equal", vals: anySlice([]int{1})},
        {kind: "Insert", vals: anySlice([]int{4})},
        {kind: "Equal", vals: anySlice([]int{2, 3})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_Paper_MiddleDelete(t *testing.T) {
    old := anySlice([]int{1, 4, 2, 3})
    neu := anySlice([]int{1, 2, 3})
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Equal", vals: anySlice([]int{1})},
        {kind: "Delete", vals: anySlice([]int{4})},
        {kind: "Equal", vals: anySlice([]int{2, 3})},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_Paper_NestedInsertFromScalar(t *testing.T) {
    old := []any{1}
    neu := []any{anySlice([]int{1})}
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Delete", vals: []any{1}},
        {kind: "Insert", vals: []any{anySlice([]int{1})}},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_Paper_NestedDeleteToScalar(t *testing.T) {
    old := []any{anySlice([]int{1})}
    neu := []any{1}
    got := jsondiffgo.Myers(old, neu)
    want := []simpleDiff{
        {kind: "Delete", vals: []any{anySlice([]int{1})}},
        {kind: "Insert", vals: []any{1}},
    }
    if !reflect.DeepEqual(simplify(got), want) {
        t.Fatalf("unexpected diff. got=%#v want=%#v", simplify(got), want)
    }
}

func TestMyers_RearrangesForSmallerDiffs(t *testing.T) {
    // Case 1
    {
        old := anySlice([]int{3, 2, 0, 2})
        neu := anySlice([]int{2, 2, 0, 2})
        got := jsondiffgo.Myers(old, neu)
        want := []simpleDiff{
            {kind: "Delete", vals: anySlice([]int{3})},
            {kind: "Insert", vals: anySlice([]int{2})},
            {kind: "Equal", vals: anySlice([]int{2, 0, 2})},
        }
        if !reflect.DeepEqual(simplify(got), want) {
            t.Fatalf("unexpected diff case1. got=%#v want=%#v", simplify(got), want)
        }
    }

    // Case 2
    {
        old := anySlice([]int{3, 2, 1, 0, 2})
        neu := anySlice([]int{2, 1, 2, 1, 0, 2})
        got := jsondiffgo.Myers(old, neu)
        want := []simpleDiff{
            {kind: "Delete", vals: anySlice([]int{3})},
            {kind: "Insert", vals: anySlice([]int{2, 1})},
            {kind: "Equal", vals: anySlice([]int{2, 1, 0, 2})},
        }
        if !reflect.DeepEqual(simplify(got), want) {
            t.Fatalf("unexpected diff case2. got=%#v want=%#v", simplify(got), want)
        }
    }

    // Case 3
    {
        old := anySlice([]int{3, 2, 2, 1, 0, 2})
        neu := anySlice([]int{2, 2, 1, 2, 1, 0, 2})
        got := jsondiffgo.Myers(old, neu)
        want := []simpleDiff{
            {kind: "Delete", vals: anySlice([]int{3})},
            {kind: "Equal", vals: anySlice([]int{2, 2, 1})},
            {kind: "Insert", vals: anySlice([]int{2, 1})},
            {kind: "Equal", vals: anySlice([]int{0, 2})},
        }
        if !reflect.DeepEqual(simplify(got), want) {
            t.Fatalf("unexpected diff case3. got=%#v want=%#v", simplify(got), want)
        }
    }

    // Case 4
    {
        old := anySlice([]int{3, 2, 0, 2})
        neu := anySlice([]int{2, 2, 1, 0, 2})
        got := jsondiffgo.Myers(old, neu)
        want := []simpleDiff{
            {kind: "Delete", vals: anySlice([]int{3})},
            {kind: "Equal", vals: anySlice([]int{2})},
            {kind: "Insert", vals: anySlice([]int{2, 1})},
            {kind: "Equal", vals: anySlice([]int{0, 2})},
        }
        if !reflect.DeepEqual(simplify(got), want) {
            t.Fatalf("unexpected diff case4. got=%#v want=%#v", simplify(got), want)
        }
    }
}
