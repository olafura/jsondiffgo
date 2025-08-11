package jsondiffgo

import (
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "reflect"
    "testing"
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
        return nil, false, nil
    }
    cmd := exec.Command("node", helper, s1, s2)
    out, err := cmd.Output()
    if err != nil {
        // Likely missing module; skip
        return nil, false, nil
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
        if err := json.Unmarshal([]byte(tc.a), &j1); err != nil { t.Fatal(err) }
        if err := json.Unmarshal([]byte(tc.b), &j2); err != nil { t.Fatal(err) }
        got := DiffRoot(j1, j2)

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

