# jsondiffgo

Go implementation of JSON diff and patch compatible with the jsondiffpatch format. Ported from the Scala reference and validated against jsondiffpatch via optional JS-based comparison tests.

## Installation

```bash
go get github.com/jsondiffgo
```

Then import in your code:

```go
import "github.com/jsondiffgo"
```

## Usage

The package works on parsed JSON values (`any`), producing diffs in the jsondiffpatch format and applying them back.

### Diff

Simple change:

```go
// a: {"test": 1}
// b: {"test": 2}
diff := jsondiffgo.Diff(map[string]any{"test": 1}, map[string]any{"test": 2})
// diff => map[string]any{"test": []any{1, 2}}
```

Arrays (insertions/deletions use the `_t: "a"` array marker):

```go
// a: {"test": [1,2,3]}
// b: {"test": [2,3]}
diff := jsondiffgo.Diff(
    map[string]any{"test": []any{1, 2, 3}},
    map[string]any{"test": []any{2, 3}},
)
// diff => {"test": {"_0": [1, 0, 0], "_t": "a"}}
```

Nested objects inside arrays:

```go
// a: {"test": [{"x":1}]}
// b: {"test": [{"x":2}]}
diff := jsondiffgo.Diff(
    map[string]any{"test": []any{map[string]any{"x": 1}}},
    map[string]any{"test": []any{map[string]any{"x": 2}}},
)
// diff => {"test": {"0": {"x": [1, 2]}, "_t": "a"}}
```

### Patch

Apply a jsondiffpatch-style diff back to an object root:

```go
obj := map[string]any{"test": 1}
diff := map[string]any{"test": []any{1, 2}}
patched, err := jsondiffgo.Patch(obj, diff)
if err != nil {
    // handle error
}
// patched => map[string]any{"test": 2}
```

Array edits and nested object patches are supported:

```go
obj := map[string]any{"test": []any{1, 2, 3}}
diff := map[string]any{"test": map[string]any{"2": []any{4}, "_2": []any{3, 0, 0}, "_t": "a"}}
patched, err := jsondiffgo.Patch(obj, diff)
if err != nil {
    // handle error
}
// patched => map[string]any{"test": []any{1, 2, 4}}
```

## Recent Improvements

- **Error Handling:** The `Patch` function now returns an error, making the library more robust against invalid patches.
- **Refactoring:** The complex `applyArrayPatch` function has been refactored into smaller, more manageable functions, improving readability and maintainability.

## API

- `func Diff(a, b any) map[string]any`
  - Compute a jsondiffpatch-style diff between two parsed JSON values. Returns an object at the root (empty when values are equal).
- `func Patch(obj map[string]any, diff map[string]any) (map[string]any, error)`
  - Apply a jsondiffpatch-style diff to an object and return the patched object. Returns an error if the patch is invalid.

Note: The intended usage is with JSON object roots. Non-object roots are handled, but object roots match jsondiffpatch behavior and the included tests.

## Diff Format Compatibility

The diff format matches jsondiffpatch:

- Simple change: `[old, new]`
- Addition: `[new]`
- Deletion: `[old, 0, 0]`
- Arrays: object with `_t: "a"`, with insertions by index keys (e.g. `"2": [value]`) and deletions as underscore keys (e.g. `"_2": [old, 0, 0]`).

The repository includes tests that compare against jsondiffpatch using Node (optional).

## Development

Make targets assume a local Go toolchain:

- Test: `make test`
- Benchmarks: `make bench`
- Fuzz (Go fuzzing): `make fuzz` (set `FUZZTIME` to control duration)
- Lint: `make lint` (requires `golangci-lint`)
- Compare with jsondiffpatch (optional): `make js-compare`
  - Requires Node.js and `npm install jsondiffpatch`
  - Set `JSONDIFFGO_COMPARE_JS=1` to enable the JS comparison tests

Benchmarks expect JSON files under `profile-data/` (already included in this repo). They mirror the medium/big profiles from the reference implementation.

## License

See `LICENSE.md`.