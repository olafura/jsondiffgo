package jsondiffgo

import (
	"encoding/json"
	"os"
	"testing"
)

// benchSink prevents compiler eliminating results during benchmarks.
var benchSink map[string]any

func mustParseForBench(b *testing.B, data []byte) any {
	b.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		b.Fatalf("failed to parse JSON: %v", err)
	}
	return v
}

func loadJSONFileOrSkip(b *testing.B, path string) ([]byte, any) {
	b.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("skipping: cannot read %s: %v", path, err)
	}
	return data, mustParseForBench(b, data)
}

func BenchmarkDiff_Big(b *testing.B) {
	// Mirrors Scala ProfileBig
	data1, j1 := loadJSONFileOrSkip(b, "profile-data/ModernAtomic.json")
	data2, j2 := loadJSONFileOrSkip(b, "profile-data/LegacyAtomic.json")

	b.ReportAllocs()
	b.SetBytes(int64(len(data1) + len(data2)))
	b.ResetTimer()

	var res map[string]any
	for i := 0; i < b.N; i++ {
		res = Diff(j1, j2)
	}
	benchSink = res
}

func BenchmarkDiff_Medium(b *testing.B) {
	// Mirrors Scala ProfileMedium (without fixed iteration count)
	data1, j1 := loadJSONFileOrSkip(b, "profile-data/cdc.json")
	data2, j2 := loadJSONFileOrSkip(b, "profile-data/edg.json")

	b.ReportAllocs()
	b.SetBytes(int64(len(data1) + len(data2)))
	b.ResetTimer()

	var res map[string]any
	for i := 0; i < b.N; i++ {
		res = Diff(j1, j2)
	}
	benchSink = res
}
