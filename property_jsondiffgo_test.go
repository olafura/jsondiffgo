package jsondiffgo

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

// -------------------------
// Property-based with testing/quick
// -------------------------

type jsonObject struct{ M map[string]any }

// Generate implements quick.Generator to produce random JSON-like objects.
func (jsonObject) Generate(r *rand.Rand, size int) reflect.Value {
	// limit size/complexity
	if size <= 0 {
		size = 1
	}
	m := genMap(r, 0, 3+size%3) // depth 0, max depth small
	return reflect.ValueOf(jsonObject{M: m})
}

func genMap(r *rand.Rand, depth, maxDepth int) map[string]any {
	n := r.Intn(4) // up to 3 keys
	if depth >= maxDepth {
		n = r.Intn(2) // keep shallow at max depth
	}
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		key := randKey(r)
		m[key] = genValue(r, depth+1, maxDepth)
	}
	return m
}

func genArray(r *rand.Rand, depth, maxDepth int) []any {
	n := r.Intn(4)
	if depth >= maxDepth {
		n = r.Intn(2)
	}
	a := make([]any, n)
	for i := 0; i < n; i++ {
		a[i] = genValue(r, depth+1, maxDepth)
	}
	return a
}

func genValue(r *rand.Rand, depth, maxDepth int) any {
	if depth > maxDepth {
		// primitives only
		switch r.Intn(4) {
		case 0:
			return r.Intn(1000)
		case 1:
			return r.Float64()
		case 2:
			return r.Intn(2) == 0
		default:
			return randString(r)
		}
	}
	switch r.Intn(6) {
	case 0:
		return genMap(r, depth, maxDepth)
	case 1:
		return genArray(r, depth, maxDepth)
	case 2:
		return r.Intn(1000)
	case 3:
		// avoid NaN/Inf
		return float64(r.Int63n(1_000_000)) / 10.0
	case 4:
		return r.Intn(2) == 0
	default:
		if r.Intn(10) == 0 {
			return nil
		}
		return randString(r)
	}
}

func randKey(r *rand.Rand) string {
	l := 1 + r.Intn(6)
	b := make([]byte, l)
	for i := range b {
		b[i] = byte('a' + r.Intn(26))
	}
	return string(b)
}

func randString(r *rand.Rand) string {
	l := r.Intn(8)
	b := make([]byte, l)
	for i := range b {
		b[i] = byte('a' + r.Intn(26))
	}
	return string(b)
}

func TestProperty_RoundTrip_Quick(t *testing.T) {
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     newPseudoCryptoRand(),
	}
	prop := func(o1, o2 jsonObject) bool {
		// compute diff and patch, must equal
		d := DiffRoot(o1.M, o2.M)
		p := PatchRoot(o1.M, d)
		if !reflect.DeepEqual(p, o2.M) {
			// serialize for stable diagnostics
			b1, _ := json.Marshal(o1.M)
			b2, _ := json.Marshal(o2.M)
			dp, _ := json.Marshal(d)
			pp, _ := json.Marshal(p)
			t.Logf("o1=%s\no2=%s\ndiff=%s\npatched=%s", b1, b2, dp, pp)
			return false
		}
		return true
	}
	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("property failed: %v", err)
	}
}

// newPseudoCryptoRand provides a seed from crypto/rand to reduce flakiness.
func newPseudoCryptoRand() *rand.Rand {
	var seed int64
	var b [8]byte
	if _, err := crand.Read(b[:]); err == nil {
		seed = int64(binary.LittleEndian.Uint64(b[:]))
	} else {
		seed = time.Now().UnixNano()
	}
	return rand.New(rand.NewSource(seed))
}

// -------------------------
// Native fuzzing: go test -fuzz=Fuzz -run=^$
// -------------------------

func FuzzRoundTripJSON(f *testing.F) {
	// Seed with some known cases
	seeds := []string{
		`{"1":1}`,
		`{"1":[1,2,3]}`,
		`{"1":{"a":1,"b":2}}`,
		`{"a":[{"x":1}]}`,
	}
	for _, s := range seeds {
		f.Add(s, s)
	}
	f.Add(`{"1":1}`, `{"1":2}`)
	f.Add(`{"1":[1,2,3]}`, `{"1":[1,2,4]}`)

	f.Fuzz(func(t *testing.T, s1 string, s2 string) {
		// Limit overly large inputs
		if len(s1)+len(s2) > 1<<20 { // ~1MB
			t.Skip()
		}

		var v1, v2 any
		if err := json.Unmarshal([]byte(s1), &v1); err != nil {
			t.Skip()
		}
		if err := json.Unmarshal([]byte(s2), &v2); err != nil {
			t.Skip()
		}

		// Ensure object roots; wrap otherwise
		m1, ok1 := v1.(map[string]any)
		if !ok1 {
			m1 = map[string]any{"_": v1}
		}
		m2, ok2 := v2.(map[string]any)
		if !ok2 {
			m2 = map[string]any{"_": v2}
		}

		d := DiffRoot(m1, m2)
		p := PatchRoot(m1, d)
		if !reflect.DeepEqual(p, m2) {
			// Not expected to fail; crash to file interesting cases
			t.Fatalf("round-trip mismatch\ns1=%s\ns2=%s\ndiff=%v\npatched=%v", s1, s2, d, p)
		}
	})
}
