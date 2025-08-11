package jsondiffgo

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func mustReadJSON(t *testing.T, path string) any {
	t.Helper()
	// #nosec G304 -- test helper reads local testdata paths
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return v
}

func TestBigDiff_MatchExpected(t *testing.T) {
	a := mustReadJSON(t, "testdata/big_json1.json")
	b := mustReadJSON(t, "testdata/big_json2.json")
	got := Diff(a, b)
	// expected diff object from the Elixir suite
	want := map[string]any{
		"_id":           []any{"56353d1bca16dd7354045f7f", "56353d1bec3821c78ad14479"},
		"about":         []any{"Laborum cupidatat proident deserunt fugiat aliquip deserunt. Mollit deserunt amet ut tempor veniam qui. Nulla ipsum non nostrud ut magna excepteur nulla non cupidatat magna ipsum.\r\n", "Consequat ullamco proident anim sunt ipsum esse Lorem tempor pariatur. Nostrud officia mollit aliqua sit consectetur sint minim veniam proident labore anim incididunt ex. Est amet laboris pariatur ut id qui et.\r\n"},
		"address":       []any{"265 Sutton Street, Tioga, Hawaii, 9975", "919 Lefferts Avenue, Winchester, Colorado, 2905"},
		"age":           []any{float64(21), float64(29)},
		"balance":       []any{"$1,343.75", "$3,273.15"},
		"company":       []any{"RAMJOB", "ANDRYX"},
		"email":         []any{"eleanorbaxter@ramjob.com", "talleyreyes@andryx.com"},
		"eyeColor":      []any{"brown", "blue"},
		"favoriteFruit": []any{"apple", "banana"},
		"gender":        []any{"female", "male"},
		"friends": map[string]any{
			"0":  map[string]any{"name": []any{"Larsen Sawyer", "Shelby Barrett"}},
			"1":  map[string]any{"name": []any{"Frost Carey", "Gloria Mccray"}},
			"2":  map[string]any{"name": []any{"Irene Lee", "Hopper Luna"}},
			"_t": "a",
		},
		"greeting":   []any{"Hello, Eleanor Baxter! You have 8 unread messages.", "Hello, Talley Reyes! You have 2 unread messages."},
		"guid":       []any{"809e01c1-b8c4-4d49-a9e7-204091cd6ae8", "b2b50dae-5d30-4514-82b1-26714d91e264"},
		"index":      []any{float64(0), float64(1)},
		"isActive":   []any{true, false},
		"latitude":   []any{-44.600585, 39.655822},
		"longitude":  []any{-9.257008, -70.899696},
		"name":       []any{"Eleanor Baxter", "Talley Reyes"},
		"phone":      []any{"+1 (876) 456-3989", "+1 (895) 435-3714"},
		"registered": []any{"2014-07-20T11:36:42 +04:00", "2015-03-11T11:45:43 +04:00"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("big diff mismatch\n got=%v\nwant=%v", got, want)
	}
}

func TestBigPatch_RoundTrip(t *testing.T) {
	a := mustReadJSON(t, "testdata/big_json1.json").(map[string]any)
	b := mustReadJSON(t, "testdata/big_json2.json").(map[string]any)
	d := Diff(a, b)
	p, err := Patch(a, d)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(p, b) {
		t.Fatalf("big patch mismatch")
	}
}

func TestArrayPatch_ShiftInside(t *testing.T) {
	a := map[string]any{"1": []any{float64(1), float64(2), float64(3)}}
	want := map[string]any{"1": []any{float64(1), float64(2), float64(0), float64(3)}}
	// Insertion before the last element
	diff := map[string]any{"1": map[string]any{"2": []any{float64(0)}, "_t": "a"}}
	got, err := Patch(a, diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("patch shift inside mismatch: got=%v want=%v", got, want)
	}
}

func TestArrayPatch_Reorder(t *testing.T) {
	a := map[string]any{"1": []any{float64(1), float64(2), float64(3)}}
	want := map[string]any{"1": []any{float64(3), float64(2), float64(1)}}
	diff := map[string]any{"1": map[string]any{"_0": []any{"", float64(2), float64(3)}, "_2": []any{"", float64(0), float64(3)}, "_t": "a"}}
	got, err := Patch(a, diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("patch reorder mismatch: got=%v want=%v", got, want)
	}
}

func TestIndexWithTwoDigits_NoChange(t *testing.T) {
	// Based on Elixir test: ensure multi-digit array keys are handled
	cards := []any{}
	for i := 1; i <= 12; i++ {
		key := "foo" + toString(i)
		cards = append(cards, map[string]any{key: true})
	}
	a := map[string]any{"cards": cards}
	b := map[string]any{"cards": cards}
	// Diff that refers to index 12; should be ignored (out of range)
	diff := map[string]any{"cards": map[string]any{"12": map[string]any{"foo11": []any{true, false}}, "_t": "a"}}
	got, err := Patch(a, diff)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(got, b) {
		t.Fatalf("two-digit index patch changed value: got=%v want=%v", got, b)
	}
}

func toString(i int) string { return strconv.Itoa(i) }

func TestChangeEleventhItemInList(t *testing.T) {
	list := make([]any, 100)
	for i := 0; i < 100; i++ {
		list[i] = i + 1
	}
	a := map[string]any{"primitives": list}
	list2 := append([]any{}, list...)
	list2[10] = -(list2[10].(int) + 0) // change 11th item
	b := map[string]any{"primitives": list2}
	d := Diff(a, b)
	p, err := Patch(a, d)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(p, b) {
		t.Fatalf("eleventh item patch mismatch")
	}
}

func TestChangeEleventhItemInListOfMaps(t *testing.T) {
	list := make([]any, 100)
	for i := 0; i < 100; i++ {
		list[i] = map[string]any{"val": i + 1}
	}
	a := map[string]any{"maps": list}
	list2 := append([]any{}, list...)
	orig := list2[10].(map[string]any)
	changed := map[string]any{}
	for k, v := range orig {
		changed[k] = v
	}
	changed["value"] = "changed"
	list2[10] = changed
	b := map[string]any{"maps": list2}
	d := Diff(a, b)
	p, err := Patch(a, d)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(p, b) {
		t.Fatalf("eleventh map patch mismatch")
	}
}

func TestNullFieldsPreservedAfterPatch(t *testing.T) {
	a := map[string]any{"name": "original", "should_be_nil": nil}
	b := map[string]any{"name": "changed", "should_be_nil": nil}
	d := Diff(a, b)
	p, err := Patch(a, d)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(p, b) {
		t.Fatalf("nil preservation mismatch: got=%v want=%v", p, b)
	}
}

func TestChangeAndAddItemsInLargeList(t *testing.T) {
	list := make([]any, 1000)
	for i := 0; i < 1000; i++ {
		list[i] = map[string]any{"val": i + 1}
	}
	changed := map[string]any{"val": "changed", "a_new_field": true}
	// replace at fixed positions
	idxs1 := []int{1, 33, 127, 68, 374, 782, 683, 237, 912}
	newList := append([]any{}, list...)
	for _, idx := range idxs1 {
		if idx < len(newList) {
			newList[idx] = changed
		}
	}
	// insert items at positions
	idxs2 := []int{17, 112, 678, 234, 922, 63, 876, 5}
	for _, idx := range idxs2 {
		if idx < 0 {
			idx = 0
		}
		if idx >= len(newList) {
			newList = append(newList, changed)
		} else {
			newList = append(newList[:idx+1], newList[idx:]...)
			newList[idx] = changed
		}
	}
	// append 20 items
	for i := 0; i < 20; i++ {
		newList = append(newList, changed)
	}
	a := map[string]any{"maps": list}
	b := map[string]any{"maps": newList}
	d := Diff(a, b)
	p, err := Patch(a, d)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	if !reflect.DeepEqual(p, b) {
		t.Fatalf("large list patch mismatch")
	}
}

func TestDifferentNumericTypes_ProduceDiff(t *testing.T) {
	a := map[string]any{"1": 4, "2": 2}
	b := map[string]any{"1": 4.0, "2": 2}
	want := map[string]any{"1": []any{4, 4.0}}
	got := Diff(a, b)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("numeric diff mismatch: got=%v want=%v", got, want)
	}
}
