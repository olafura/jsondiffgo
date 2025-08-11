Contributing

Prerequisites
- Go (stable)
- Optional: Node.js (for jsondiffpatch comparison tests)

Run tests
- go test ./...

Benchmarks
- Files expected in profile-data/ matching the Scala profiles.
- Run: go test -run '^$' -bench . -benchmem

Profiling CLI
- Build: go build ./cmd/jsondiffgo-profile
- Big: ./jsondiffgo-profile -mode=big -file1=profile-data/ModernAtomic.json -file2=profile-data/LegacyAtomic.json
- Medium: ./jsondiffgo-profile -mode=medium -file1=profile-data/cdc.json -file2=profile-data/edg.json -iters=50
- Profiles: add -cpuprofile=cpu.out -memprofile=mem.out

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

