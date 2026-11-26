package idgen_test

import (
	"regexp"
	"sort"
	"testing"

	"go-arch/library/idgen"
)

// hex32 is exactly what the seam promises: 32 hard hex chars, no hyphens.
var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestRealIsDashlessHex32(t *testing.T) {
	ids := idgen.Real{}
	for i := 0; i < 1000; i++ {
		got := ids.New()
		if !hex32.MatchString(got) {
			t.Fatalf("id %q does not match %s", got, hex32)
		}
		if len(got) != 32 {
			t.Fatalf("id %q has length %d, want 32", got, len(got))
		}
		if regexp.MustCompile(`-`).MatchString(got) {
			t.Fatalf("id %q contains a hyphen", got)
		}
	}
}

func TestRealIsUnique(t *testing.T) {
	ids := idgen.Real{}
	seen := make(map[string]struct{}, 10_000)
	for i := 0; i < 10_000; i++ {
		got := ids.New()
		if _, dup := seen[got]; dup {
			t.Fatalf("duplicate id generated: %q", got)
		}
		seen[got] = struct{}{}
	}
}

// UUID v7's 48-bit timestamp prefix means ids generated later must sort after
// earlier ones lexicographically. This is the property that makes them a good
// log/index key, so it is worth pinning.
func TestRealIsTimeOrdered(t *testing.T) {
	ids := idgen.Real{}

	const n = 500
	generated := make([]string, n)
	for i := range generated {
		generated[i] = ids.New()
	}

	sorted := make([]string, n)
	copy(sorted, generated)
	sort.Strings(sorted)

	for i := range generated {
		if generated[i] != sorted[i] {
			t.Fatalf("ids not time-ordered at index %d: generated %q, sorted %q",
				i, generated[i], sorted[i])
		}
	}
}

func TestFakeMatchesRealShape(t *testing.T) {
	ids := idgen.NewFake()
	for i := 1; i <= 5; i++ {
		got := ids.New()
		if !hex32.MatchString(got) {
			t.Fatalf("fake id %q does not match real shape %s", got, hex32)
		}
	}
}

func TestFakeIsDeterministicAndSequential(t *testing.T) {
	a := idgen.NewFake()
	b := idgen.NewFake()

	for i := 0; i < 3; i++ {
		if ga, gb := a.New(), b.New(); ga != gb {
			t.Fatalf("fake not deterministic: %q != %q", ga, gb)
		}
	}

	first := idgen.NewFake().New()
	if first != "00000000000000000000000000000001" {
		t.Fatalf("first fake id = %q, want ...0001", first)
	}
}
