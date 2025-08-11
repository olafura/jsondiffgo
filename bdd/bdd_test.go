package bdd

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "reflect"
    "testing"
    "time"

    "github.com/cucumber/godog"
    "github.com/jsondiffgo"
)

// Suite context for step state
type suite struct {
    a, b any
    diff any
    orig map[string]any
    patch map[string]any
    result map[string]any
    jsd   any
    jsok  bool
    skipJSCompare bool
}

func readFileFlexible(path string) ([]byte, error) {
    // #nosec G304 -- the features control only relative paths inside repo testdata
    b, err := os.ReadFile(path)
    if err == nil {
        return b, nil
    }
    // try parent-relative
    if len(path) > 0 && path[0] != '/' {
        // #nosec G304 -- see above; only repo-local paths are used
        if b2, err2 := os.ReadFile("../" + path); err2 == nil {
            return b2, nil
        }
    }
    return nil, err
}

func parseJSON[T any](s string) (any, error) {
    var v any
    if err := json.Unmarshal([]byte(s), &v); err != nil {
        return nil, err
    }
    return v, nil
}

func (s *suite) givenJSONA(doc *godog.DocString) error {
    v, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    s.a = v
    return nil
}

func (s *suite) givenJSONB(doc *godog.DocString) error {
    v, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    s.b = v
    return nil
}

func (s *suite) givenJSONAFromFile(path string) error {
    b, err := readFileFlexible(path)
    if err != nil {
        return err
    }
    var v any
    if err := json.Unmarshal(b, &v); err != nil {
        return err
    }
    s.a = v
    return nil
}

func (s *suite) givenJSONBFromFile(path string) error {
    b, err := readFileFlexible(path)
    if err != nil {
        return err
    }
    var v any
    if err := json.Unmarshal(b, &v); err != nil {
        return err
    }
    s.b = v
    return nil
}

func (s *suite) whenIComputeTheDiff() error {
    sdiff := jsondiffgo.Diff(s.a, s.b)
    s.diff = sdiff
    // Optionally compare with jsondiffpatch via JS helper
    if d, ok, err := runJSDiff(s.a, s.b); err == nil && ok {
        s.jsd, s.jsok = d, true
    }
    return nil
}

func (s *suite) thenTheDiffEquals(doc *godog.DocString) error {
    wantAny, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    // godog DocString JSONs are objects for root diffs
    if !reflect.DeepEqual(s.diff, wantAny) {
        bw, _ := json.Marshal(wantAny)
        bg, _ := json.Marshal(s.diff)
        return fmt.Errorf("diff mismatch\nwant=%s\ngot =%s", string(bw), string(bg))
    }
    if s.jsok && !s.skipJSCompare {
        if !reflect.DeepEqual(s.diff, s.jsd) {
            bj, _ := json.Marshal(s.jsd)
            bg, _ := json.Marshal(s.diff)
            return fmt.Errorf("diff mismatch vs js helper\njs =%s\ngot=%s", string(bj), string(bg))
        }
    }
    return nil
}

func (s *suite) givenOriginalJSON(doc *godog.DocString) error {
    v, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    m, ok := v.(map[string]any)
    if !ok {
        m = map[string]any{"_": v}
    }
    s.orig = m
    return nil
}

func (s *suite) givenOriginalJSONFromFile(path string) error {
    b, err := readFileFlexible(path)
    if err != nil {
        return err
    }
    var v any
    if err := json.Unmarshal(b, &v); err != nil {
        return err
    }
    m, ok := v.(map[string]any)
    if !ok {
        m = map[string]any{"_": v}
    }
    s.orig = m
    return nil
}

func (s *suite) givenDiff(doc *godog.DocString) error {
    v, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    m, ok := v.(map[string]any)
    if !ok {
        return fmt.Errorf("diff must be an object at root")
    }
    s.patch = m
    return nil
}

func (s *suite) whenIApplyThePatch() error {
    s.result = jsondiffgo.Patch(s.orig, s.patch)
    return nil
}

func (s *suite) thenTheResultEquals(doc *godog.DocString) error {
    v, err := parseJSON[any](doc.Content)
    if err != nil {
        return err
    }
    m, ok := v.(map[string]any)
    if !ok {
        m = map[string]any{"_": v}
    }
    if !reflect.DeepEqual(s.result, m) {
        bw, _ := json.Marshal(m)
        bg, _ := json.Marshal(s.result)
        return fmt.Errorf("patch mismatch\nwant=%s\ngot =%s", string(bw), string(bg))
    }
    // If JS helper is available and we have a given diff, verify it matches js diff
    if s.patch != nil && !s.skipJSCompare {
        if d, ok, err := runJSDiff(s.orig, m); err == nil && ok {
            if !reflect.DeepEqual(s.patch, d) {
                bj, _ := json.Marshal(d)
                bp, _ := json.Marshal(s.patch)
                return fmt.Errorf("given diff mismatch vs js helper\njs =%s\ngiven=%s", string(bj), string(bp))
            }
        }
    }
    return nil
}

func (s *suite) skipJS() error {
    s.skipJSCompare = true
    return nil
}

// runJSDiff executes the Node helper if JSONDIFFGO_COMPARE_JS is set and helper exists.
func runJSDiff(a, b any) (any, bool, error) {
    if os.Getenv("JSONDIFFGO_COMPARE_JS") == "" {
        return nil, false, nil
    }
    helper := "js/test_helper.js"
    if _, err := os.Stat(helper); err != nil {
        // try from parent
        helper = "../js/test_helper.js"
        if _, err2 := os.Stat(helper); err2 != nil {
            return nil, false, nil
        }
    }
    j1, _ := json.Marshal(a)
    j2, _ := json.Marshal(b)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    cmd := exec.CommandContext(ctx, "node", helper, string(j1), string(j2))
    out, err := cmd.Output()
    if err != nil {
        return nil, false, nil
    }
    var v any
    if len(out) == 0 {
        return map[string]any{}, true, nil
    }
    if err := json.Unmarshal(out, &v); err != nil {
        return nil, false, err
    }
    return v, true, nil
}

// InitializeScenario wires the step definitions.
func InitializeScenario(ctx *godog.ScenarioContext) {
    s := &suite{}
    ctx.Step(`^JSON A:$`, s.givenJSONA)
    ctx.Step(`^JSON A from file "([^"]+)"$`, s.givenJSONAFromFile)
    ctx.Step(`^JSON B:$`, s.givenJSONB)
    ctx.Step(`^JSON B from file "([^"]+)"$`, s.givenJSONBFromFile)
    ctx.Step(`^I compute the diff$`, s.whenIComputeTheDiff)
    ctx.Step(`^the diff equals:$`, s.thenTheDiffEquals)

    ctx.Step(`^Original JSON:$`, s.givenOriginalJSON)
    ctx.Step(`^Original JSON from file "([^"]+)"$`, s.givenOriginalJSONFromFile)
    ctx.Step(`^Diff:$`, s.givenDiff)
    ctx.Step(`^I apply the patch$`, s.whenIApplyThePatch)
    ctx.Step(`^the result equals:$`, s.thenTheResultEquals)
    ctx.Step(`^JS compare is skipped$`, s.skipJS)
}

func TestBDDFeatures(t *testing.T) {
    suite := godog.TestSuite{
        Name:                "jsondiffgo-bdd",
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"features"},
            Strict:   true,
            TestingT: t,
        },
    }
    if suite.Run() != 0 {
        t.Fatal("godog suite failed")
    }
}
