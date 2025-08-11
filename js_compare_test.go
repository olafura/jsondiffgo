package jsondiffgo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

// jsDiff invokes Node + jsondiffpatch via js/test_helper.js.
// Returns (diff, true, nil) on success; returns (_, false, nil) if helper not available; error otherwise.
func jsDiff(s1, s2 string) (any, bool, error) {
	// Only run when explicitly enabled to avoid CI env issues
	if os.Getenv("JSONDIFFGO_COMPARE_JS") == "" {
		return nil, false, nil
	}
	helper := filepath.Join("js", "test_helper.js")
	if _, err := os.Stat(helper); err != nil {
		return nil, false, fmt.Errorf("No js helper: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "node", helper, s1, s2)
	out, err := cmd.Output()
	if err != nil {
		// Likely missing module; skip
		return nil, false, fmt.Errorf("Node error: %w\ns1: %s\ns2: %s", err, s1, s2)
	}
	var v any
	if len(out) == 0 || string(out) == "null" {
		return nil, true, nil
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func TestCompareWithJsondiffpatch(t *testing.T) {
	cases := []struct{ a, b string }{
		{`{"1":1}`, `{"1":2}`},
		{`{"1":[1,2,3]}`, `{"1":[1,2,4]}`},
		{`{"1":[1,2,3]}`, `{"1":[2,3]}`},
		{`{"1":[1]}`, `{"1":[{"1":2}]}`},
		{`{"1":[1,{"1":1}]}`, `{"1":[{"1":2}]}`},
		{`{"a":{"x":1},"b":2}`, `{"a":{"x":2},"b":2}`},
		{`{"1":[{"1":1}]}`, `{"1":[{"1":2}]}`},
	}
	for _, tc := range cases {
		jsd, ok, err := jsDiff(tc.a, tc.b)
		if err != nil {
			t.Fatalf("js helper error: %v", err)
		}
		if !ok {
			t.Skip("JSONDIFFGO_COMPARE_JS not set or node helper unavailable; skipping")
		}
		// Our diff
		var j1, j2 any
		if err := json.Unmarshal([]byte(tc.a), &j1); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(tc.b), &j2); err != nil {
			t.Fatal(err)
		}
		got := Diff(j1, j2)

		// Normalize: jsondiffpatch returns null for no diff
		var want any
		if jsd == nil {
			want = map[string]any{}
		} else {
			want = jsd
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("mismatch with jsondiffpatch\na=%s\nb=%s\ngot=%v\nwant=%v", tc.a, tc.b, got, want)
		}
	}
}

func TestProperty_CompareWithJsondiffpatch_Quick(t *testing.T) {
	cfg := &quick.Config{MaxCount: 50, Rand: newPseudoCryptoRand()}
	prop := func(o1, o2 jsonObject) bool {
		// Marshal generator output to JSON strings to normalize numeric types
		b1, _ := json.Marshal(o1.M)
		b2, _ := json.Marshal(o2.M)

		jsd, ok, err := jsDiff(string(b1), string(b2))
		if err != nil {
			t.Fatalf("js helper error: %v", err)
		}
		if !ok {
			t.Skip("JSONDIFFGO_COMPARE_JS not set or node helper unavailable; skipping")
		}

		var j1, j2 any
		if err := json.Unmarshal(b1, &j1); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b2, &j2); err != nil {
			t.Fatal(err)
		}
		got := Diff(j1, j2)

		var want any
		if jsd == nil {
			want = map[string]any{}
		} else {
			want = jsd
		}
		if !reflect.DeepEqual(got, want) {
			t.Logf("a=%s\nb=%s\ngot=%v\nwant=%v", string(b1), string(b2), got, want)
			return false
		}
		return true
	}
	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("property failed: %v", err)
	}
}
