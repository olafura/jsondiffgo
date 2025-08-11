package jsondiffgo

import (
	"encoding/json"
	"reflect"
	"testing"
)

func parseJSON(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	return v
}

func TestJsonDiff_Basic(t *testing.T) {
	s1 := "{\"1\": 1}"
	s2 := "{\"1\": 2}"
	expected := parseJSON(t, "{\"1\": [1,2]}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_Array(t *testing.T) {
	s1 := "{\"1\": [1,2,3]}"
	s2 := "{\"1\": [1,2,4]}"
	expected := parseJSON(t, "{\"1\": {\"2\": [4], \"_2\": [3,0,0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_SameObject(t *testing.T) {
	s1 := "{\"1\": [1,2,3], \"2\": 1}"
	j1 := parseJSON(t, s1)
	got := Diff(j1, j1)
	if len(got) != 0 {
		t.Fatalf("expected empty object, got=%v", got)
	}
}

func TestJsonDiff_ObjectDiffNotChanged(t *testing.T) {
	s1 := "{\"1\": 1, \"2\": 2}"
	s2 := "{\"1\": 2, \"2\": 2}"
	expected := parseJSON(t, "{\"1\": [1,2]}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ArrayDiffNotChanged(t *testing.T) {
	s1 := "{\"1\": 1, \"2\": [1]}"
	s2 := "{\"1\": 2, \"2\": [1]}"
	expected := parseJSON(t, "{\"1\": [1,2]}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ArrayAllChanged(t *testing.T) {
	s1 := "{\"1\": [1,2,3]}"
	s2 := "{\"1\": [4,5,6]}"
	expected := parseJSON(t, "{\"1\": {\"0\": [4], \"1\": [5], \"2\": [6], \"_0\": [1,0,0], \"_1\": [2,0,0], \"_2\": [3,0,0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ArrayDeleteFirst(t *testing.T) {
	s1 := "{\"1\": [1,2,3]}"
	s2 := "{\"1\": [2,3]}"
	expected := parseJSON(t, "{\"1\": {\"_0\": [1,0,0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ArrayShiftOne(t *testing.T) {
	s1 := "{\"1\": [1,2,3]}"
	s2 := "{\"1\": [0,1,2,3]}"
	expected := parseJSON(t, "{\"1\": {\"0\": [0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ArrayWithDuplicates(t *testing.T) {
	s1 := "{\"1\": [1,2,1,3,3,2]}"
	s2 := "{\"1\": [3,1,2,1,2,3,3,2,1]}"
	expected := parseJSON(t, "{\"1\": {\"_t\": \"a\", \"0\": [3], \"4\": [2], \"8\": [1]}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ObjectInArray(t *testing.T) {
	s1 := "{\"1\": [{\"1\":1}]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	expected := parseJSON(t, "{\"1\": {\"0\": {\"1\": [1,2]}, \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ObjectWithMultipleValuesInArray(t *testing.T) {
	s1 := "{\"1\": [{\"1\":1,\"2\":2}]}"
	s2 := "{\"1\": [{\"1\":2,\"2\":2}]}"
	expected := parseJSON(t, "{\"1\": {\"0\": {\"1\": [1,2]}, \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_ObjectWithMultipleValuesPlusInArray(t *testing.T) {
	s1 := "{\"1\": [{\"1\":1,\"2\":2},{\"3\":3,\"4\":4}]}"
	s2 := "{\"1\": [{\"1\":2,\"2\":2},{\"3\":5,\"4\":6}]}"
	expected := parseJSON(t, "{\"1\": {\"0\": {\"1\": [1,2]}, \"1\": {\"3\": [3,5], \"4\": [4,6]}, \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_OneObjectInArray(t *testing.T) {
	s1 := "{\"1\": [1]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	expected := parseJSON(t, "{\"1\": {\"0\": [{\"1\":2}], \"_0\": [1,0,0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_DeletedValueWithObjectChangeInArray(t *testing.T) {
	s1 := "{\"1\": [1,{\"1\":1}]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	expected := parseJSON(t, "{\"1\": {\"0\": [{\"1\":2}], \"_0\": [1,0,0], \"_1\": [{\"1\":1},0,0], \"_t\": \"a\"}}")
	got := Diff(parseJSON(t, s1), parseJSON(t, s2))
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected diff. got=%v want=%v", got, expected)
	}
}

func TestJsonDiff_SameNumericType(t *testing.T) {
	j1 := parseJSON(t, "{\"1\": 4, \"2\": 2}")
	j2 := parseJSON(t, "{\"1\": 4, \"2\": 2}")
	got := Diff(j1, j2)
	if len(got) != 0 {
		t.Fatalf("expected empty object, got=%v", got)
	}
}

// Patch tests adapted from json_diff_ex
func TestJsonPatch_Basic(t *testing.T) {
	s1 := "{\"1\": 1}"
	s2 := "{\"1\": 2}"
	diff := parseJSON(t, "{\"1\": [1,2]}").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}

func TestJsonPatch_Array(t *testing.T) {
	s1 := "{\"1\": [1,2,3]}"
	s2 := "{\"1\": [1,2,4]}"
	diff := parseJSON(t, "{\"1\": {\"2\": [4], \"_2\": [3,0,0], \"_t\": \"a\"}} ").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}

func TestJsonPatch_ObjectInArray(t *testing.T) {
	s1 := "{\"1\": [{\"1\":1}]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	diff := parseJSON(t, "{\"1\": {\"0\": {\"1\": [1,2]}, \"_t\": \"a\"}} ").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}

func TestJsonPatch_OneObjectInArray(t *testing.T) {
	s1 := "{\"1\": [1]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	diff := parseJSON(t, "{\"1\": {\"0\": [{\"1\":2}], \"_0\": [1,0,0], \"_t\": \"a\"}} ").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}

func TestJsonPatch_DeletedValueWithObjectChangeInArray(t *testing.T) {
	s1 := "{\"1\": [1,{\"1\":1}]}"
	s2 := "{\"1\": [{\"1\":2}]}"
	diff := parseJSON(t, "{\"1\": {\"0\": [{\"1\":2}], \"_0\": [1,0,0], \"_1\": [{\"1\":1},0,0], \"_t\": \"a\"}} ").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}

func TestJsonPatch_DeletedKeyWorks(t *testing.T) {
	s1 := "{\"foo\": 1}"
	s2 := "{\"bar\": 3}"
	diff := parseJSON(t, "{\"bar\": [3], \"foo\": [1,0,0]} ").(map[string]any)
	patched, err := Patch(parseJSON(t, s1).(map[string]any), diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(patched, parseJSON(t, s2)) {
		t.Fatalf("patch mismatch: got=%v want=%v", patched, s2)
	}
}
