package central

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGobEncoding(t *testing.T) {
	cc := Config{
		Networks: map[string]*Network{
			"testnet": &Network{
				Name: "sasara",
			},
		},
	}
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(cc)
	if err != nil {
		t.Fatalf("Encode: %s", err)
	}
	var cc2 Config
	err = gob.NewDecoder(buf).Decode(&cc2)
	if err != nil {
		t.Fatalf("Decode: %s", err)
	}
	if !cmp.Equal(cc, cc2) {
		t.Log(cmp.Diff(cc, cc2))
		t.Fatal("!equal")
	}
}

func TestJSONEncoding(t *testing.T) {
	cc := Config{
		Networks: map[string]*Network{
			"testnet": &Network{
				Name: "sasara",
			},
		},
	}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(cc)
	if err != nil {
		t.Fatalf("Encode: %s", err)
	}
	var cc2 Config
	err = json.NewDecoder(buf).Decode(&cc2)
	if err != nil {
		t.Fatalf("Decode: %s", err)
	}
	if !cmp.Equal(cc, cc2) {
		t.Log(cmp.Diff(cc, cc2))
		t.Fatal("!equal")
	}
}

func TestIPNetSubsetOf(t *testing.T) {
	type testCase struct {
		a, b   string
		subset bool
	}

	cases := []testCase{
		{"1.1.3.0/24", "1.1.3.97/32", true},
		{"1.1.3.0/24", "1.1.4.0/32", false},
		{"1.1.3.97/24", "1.1.3.0/32", false},
		{"0.0.0.0/24", "0.0.0.0/8", false},
	}
	for i, tc := range cases {
		_, net1, err := net.ParseCIDR(tc.a)
		if err != nil {
			panic(err)
		}
		_, net2, err := net.ParseCIDR(tc.b)
		if err != nil {
			panic(err)
		}
		if got := IPNetSubsetOf(*net1, *net2); tc.subset != got {
			t.Logf("test case %d", i)
			t.Logf("    a = %s", tc.a)
			t.Logf("    b = %s", tc.b)
			t.Logf("    want = %t ; got = %t", tc.subset, got)
		}
	}
}
