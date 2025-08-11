Contributing

Prerequisites
- Go (stable)
- Optional: Node.js (for jsondiffpatch comparison tests)

Run tests
- go test ./...

Benchmarks
- Run: go test -run '^$' -bench . -benchmem

Fuzzing (Go 1.18+)
- go test -run '^$' -fuzz=Fuzz -fuzztime=10s

Property-based tests
- Property tests use testing/quick and run as part of go test.

Compare with jsondiffpatch (Node)
- npm install jsondiffpatch
- JSONDIFFGO_COMPARE_JS=1 go test ./...
- JS helper: js/test_helper.js

Linting
- golangci-lint is configured via .golangci.yml

