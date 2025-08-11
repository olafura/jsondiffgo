package bdd

import (
    "encoding/json"
    "fmt"
    "reflect"
    "testing"

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

func (s *suite) whenIComputeTheDiff() error {
    sdiff := jsondiffgo.Diff(s.a, s.b)
    s.diff = sdiff
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
    return nil
}

// InitializeScenario wires the step definitions.
func InitializeScenario(ctx *godog.ScenarioContext) {
    s := &suite{}
    ctx.Step(`^JSON A:$`, s.givenJSONA)
    ctx.Step(`^JSON B:$`, s.givenJSONB)
    ctx.Step(`^I compute the diff$`, s.whenIComputeTheDiff)
    ctx.Step(`^the diff equals:$`, s.thenTheDiffEquals)

    ctx.Step(`^Original JSON:$`, s.givenOriginalJSON)
    ctx.Step(`^Diff:$`, s.givenDiff)
    ctx.Step(`^I apply the patch$`, s.whenIApplyThePatch)
    ctx.Step(`^the result equals:$`, s.thenTheResultEquals)
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
