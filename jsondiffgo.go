package jsondiffgo

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
)

// Json diff implementation ported from the Scala reference.

// fastEqual performs optimized equality comparison for common JSON types
func fastEqual(a, b any) bool {
	// Try common cases first without reflection
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case nil:
		return b == nil
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	case int64:
		if bv, ok := b.(int64); ok {
			return av == bv
		}
	}

	// Fall back to reflection for complex types and type mismatches
	return reflect.DeepEqual(a, b) && reflect.TypeOf(a) == reflect.TypeOf(b)
}

// Diff computes the JSON diff between two parsed JSON values and returns
// an object (map) at the root. If there is no difference, an empty object is returned.
func Diff(a, b any) map[string]any {
	d := diff(a, b)
	if d == nil {
		return map[string]any{}
	}
	// d must be an object at the root per reference implementation
	if m, ok := d.(map[string]any); ok {
		return m
	}
	// Fallback: wrap non-object differences into an object
	return map[string]any{"_root": d}
}

// diff mirrors the behavior of JsonDiff#doDiff in Scala.
// Returns one of:
// - nil for no difference (JsNull)
// - map[string]any for object differences
// - []any for scalar differences or array/object markers
func diff(a, b any) any {
	switch aTyped := a.(type) {
	case []any:
		if bTyped, ok := b.([]any); ok {
			return diffArray(aTyped, bTyped)
		}
	case map[string]any:
		if bTyped, ok := b.(map[string]any); ok {
			return diffObject(aTyped, bTyped)
		}
	}

	// Scalars or type mismatch
	if fastEqual(a, b) {
		return nil
	}
	return []any{a, b}
}

func diffObject(o1, o2 map[string]any) any {
	diffMap := map[string]any{}

	// Union of keys
	keys := map[string]struct{}{}
	for k := range o1 {
		keys[k] = struct{}{}
	}
	for k := range o2 {
		keys[k] = struct{}{}
	}

	for k := range keys {
		v1, ok1 := o1[k]
		v2, ok2 := o2[k]
		var d any
		switch {
		case ok1 && ok2:
			d = diff(v1, v2)
		case ok1 && !ok2:
			d = []any{v1, float64(0), float64(0)}
		case !ok1 && ok2:
			d = []any{v2}
		}
		if d != nil {
			diffMap[k] = d
		}
	}

	if len(diffMap) == 0 {
		return nil
	}
	return diffMap
}

type arrayAcc struct {
	count        int
	deletedCount int
	acc          map[string]any
}

func diffArray(l1, l2 []any) any {
	// Use Myers to diff arrays
	edits := Myers(l1, l2)

	acc := arrayAcc{count: 0, deletedCount: 0, acc: map[string]any{}}
	for _, e := range edits {
		switch v := e.(type) {
		case Equal:
			n := len(v.Val)
			acc.count += n
			acc.deletedCount += n
		case Delete:
			for _, it := range v.Val {
				key := "_" + strconv.Itoa(acc.deletedCount)
				acc.acc[key] = []any{it, float64(0), float64(0)}
				acc.deletedCount++
			}
		case Insert:
			for _, it := range v.Val {
				key := strconv.Itoa(acc.count)
				acc.acc[key] = []any{it}
				acc.count++
			}
		}
	}

	// Partition underscore deletions that are objects
	deleted := map[string]any{}
	checked := map[string]any{}
	for k, v := range acc.acc {
		if splitUnderscoreMap(k, v) {
			deleted[k] = v
		} else {
			checked[k] = v
		}
	}

	var out map[string]any
	if len(deleted) == 0 {
		out = acc.acc
	} else {
		out = allChecked(checked, deleted)
		// filter nils
		for k, v := range out {
			if v == nil {
				delete(out, k)
			}
		}
	}

	if len(out) == 0 {
		return nil
	}
	out["_t"] = "a"
	return out
}

// splitUnderscoreMap implements the Scala splitUnderscoreMap predicate.
func splitUnderscoreMap(key string, value any) bool {
	if len(key) > 0 && key[0] == '_' {
		if arr, ok := value.([]any); ok && len(arr) == 3 {
			if _, isObj := arr[0].(map[string]any); isObj {
				return isZero(arr[1]) && isZero(arr[2])
			}
		}
	}
	return false
}

func isZero(v any) bool {
	switch t := v.(type) {
	case float64:
		return t == 0
	case float32:
		return t == 0
	case int:
		return t == 0
	case int64:
		return t == 0
	case json.Number:
		f, _ := t.Float64()
		return f == 0
	default:
		return false
	}
}

func toNumber(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		if err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// allChecked transforms insertions of objects combined with corresponding deletions
// into nested diffs, mirroring the Scala logic.
func allChecked(checked, deleted map[string]any) map[string]any {
	result := map[string]any{}

	// Work on a copy of deleted for mutation
	del := map[string]any{}
	for k, v := range deleted {
		del[k] = v
	}

	for k, v := range checked {
		// Only transform entries like i -> [ {..} ]
		if arr, ok := v.([]any); ok && len(arr) == 1 {
			if obj, ok2 := arr[0].(map[string]any); ok2 {
				negKey := "_" + k
				if dv, ok3 := del[negKey]; ok3 {
					if darr, ok4 := dv.([]any); ok4 && len(darr) == 3 {
						if dobj, ok5 := darr[0].(map[string]any); ok5 && isZero(darr[1]) && isZero(darr[2]) {
							nested := diff(dobj, obj)
							if nested != nil {
								result[k] = nested
							}
							delete(del, negKey)
							continue
						}
					}
				}
			}
		}
		// Otherwise keep as-is
		result[k] = v
	}

	// Append remaining deleted entries
	for k, v := range del {
		result[k] = v
	}
	return result
}

// Patch applies a jsondiffpatch-style diff to the provided object and returns the patched object.
// Both inputs must be JSON objects (map[string]any).
func Patch(obj map[string]any, diff map[string]any) map[string]any {
	return doPatch(obj, diff)
}

func doPatch(m1 map[string]any, d1 map[string]any) map[string]any {
	// Preprocess: turn [new_value] entries into new_value directly
	pre := map[string]any{}
	for k, v := range d1 {
		if arr, ok := v.([]any); ok && len(arr) == 1 {
			pre[k] = arr[0]
		} else {
			pre[k] = v
		}
	}

	// Merge
	out := map[string]any{}
	// start with original
	for k, v := range m1 {
		out[k] = v
	}
	for k, v := range pre {
		existing, has := out[k]
		if has {
			merged, ok, remove := doPatchMerge(existing, v)
			if remove {
				delete(out, k)
				continue
			}
			if ok {
				out[k] = merged
			} else {
				// replace with provided when not specially handled
				out[k] = v
			}
		} else {
			// new key: assign value as-is (it's a concrete value, not a diff)
			out[k] = v
		}
	}
	return out
}

// doPatchMerge applies one diff value vDiff to an existing value vMap.
// Returns (newValue, replaced?, removeKey?)
func doPatchMerge(vMap, vDiff any) (any, bool, bool) {
	// Case: [old, new]
	if arr, ok := vDiff.([]any); ok {
		if len(arr) == 2 {
			if fastEqual(arr[0], vMap) {
				return arr[1], true, false
			}
			// if old doesn't match, still replace with new
			return arr[1], true, false
		}
		if len(arr) == 3 && isZero(arr[1]) && isZero(arr[2]) {
			// deletion marker for object key
			return nil, false, true
		}
		// otherwise treat as replacement value
		return vDiff, true, false
	}

	// Case: object
	if m, ok := vDiff.(map[string]any); ok {
		if t, hasT := m["_t"]; hasT && t == "a" {
			// array diff
			// remove marker before applying
			return applyArrayPatch(asArray(vMap), m), true, false
		}
		// nested object diff
		return doPatch(asMap(vMap), m), true, false
	}

	// Default: replace
	return vDiff, true, false
}

func asArray(v any) []any {
	if a, ok := v.([]any); ok {
		return a
	}
	// Return empty slice for non-array types to maintain compatibility
	return []any{}
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	// Return empty map for non-object types to maintain compatibility
	return map[string]any{}
}

// applyArrayPatch implements the array patching logic for jsondiffpatch-style diffs.
func applyArrayPatch(list []any, diff map[string]any) []any {
	// Make a shallow copy of diff and remove _t
	d2 := map[string]any{}
	for k, v := range diff {
		if k == "_t" {
			continue
		}
		d2[k] = v
	}

	// Split deleted, moves and rest
	deletedIdx := map[int]struct{}{}
	type moveOp struct{ src, dest int }
	moves := make([]moveOp, 0)
	remaining := map[string]any{}
	for k, v := range d2 {
		if splitUnderscore(k, v) {
			if idx, err := strconv.Atoi(k[1:]); err == nil {
				deletedIdx[idx] = struct{}{}
			}
		} else if len(k) > 0 && k[0] == '_' {
			if arr, ok := v.([]any); ok && len(arr) == 3 {
				// jsondiffpatch move: ["", dest, 3]
				if num, ok2 := toNumber(arr[1]); ok2 {
					dest := int(num)
					if src, err := strconv.Atoi(k[1:]); err == nil {
						moves = append(moves, moveOp{src: src, dest: dest})
						continue
					}
				}
			}
			// Unknown underscore op, keep in remaining
			remaining[k] = v
		} else {
			remaining[k] = v
		}
	}

	// Remove deleted indices
	filtered := make([]any, 0, len(list))
	for i, val := range list {
		if _, isDel := deletedIdx[i]; !isDel {
			filtered = append(filtered, val)
		}
	}

	// Apply remaining operations in index order
	type kv struct {
		idx int
		val any
	}
	ops := make([]kv, 0, len(remaining))
	for k, v := range remaining {
		if idx, err := strconv.Atoi(k); err == nil {
			ops = append(ops, kv{idx: idx, val: v})
		}
	}
	// sort by idx
	sort.Slice(ops, func(i, j int) bool { return ops[i].idx < ops[j].idx })

	res := make([]any, len(filtered))
	copy(res, filtered)

	// Apply moves: use original values to locate current positions
	if len(moves) > 0 {
		// capture original list for identity
		orig := make([]any, len(list))
		copy(orig, list)
		sort.Slice(moves, func(i, j int) bool { return moves[i].dest < moves[j].dest })
		for _, m := range moves {
			if m.src < 0 || m.src >= len(orig) {
				continue
			}
			val := orig[m.src]
			// find current index of val in res
			cur := -1
			for i := range res {
				if fastEqual(res[i], val) {
					cur = i
					break
				}
			}
			if cur == -1 {
				continue
			}
			// remove at cur
			res = append(res[:cur], res[cur+1:]...)
			// insert at dest
			if m.dest < 0 {
				m.dest = 0
			}
			if m.dest >= len(res) {
				res = append(res, val)
			} else {
				res = append(res[:m.dest+1], res[m.dest:]...)
				res[m.dest] = val
			}
		}
	}

	for _, op := range ops {
		switch v := op.val.(type) {
		case map[string]any:
			// nested diff at index
			if op.idx >= 0 && op.idx < len(res) {
				res[op.idx] = doPatch(asMap(res[op.idx]), v)
			}
		case []any:
			if len(v) == 1 {
				// insert at index
				val := v[0]
				if op.idx < 0 {
					op.idx = 0
				}
				if op.idx >= len(res) {
					res = append(res, val)
				} else {
					res = append(res[:op.idx+1], res[op.idx:]...)
					res[op.idx] = val
				}
			} else if len(v) == 2 {
				// replace at index with new value
				if op.idx >= 0 && op.idx < len(res) {
					res[op.idx] = v[1]
				}
			}
		}
	}
	return res
}

// splitUnderscore identifies deletion markers like _i: [x,0,0]
func splitUnderscore(key string, value any) bool {
	if len(key) > 0 && key[0] == '_' {
		if arr, ok := value.([]any); ok && len(arr) == 3 {
			return isZero(arr[1]) && isZero(arr[2])
		}
	}
	return false
}
