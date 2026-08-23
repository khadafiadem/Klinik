package bpjs

import (
	"encoding/json"
	"os"
	"testing"
)

// Vektor referensi dibuat dengan library resmi pieroxy/lz-string v1.5.0 (Node.js).
type lzVector struct {
	Input      string `json:"input"`
	Compressed string `json:"compressed"`
}

func loadVectors(t *testing.T) map[string]lzVector {
	t.Helper()
	data, err := os.ReadFile("testdata/lzstring_vectors.json")
	if err != nil {
		t.Fatalf("gagal baca vektor: %v", err)
	}
	var vectors map[string]lzVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("gagal parse vektor: %v", err)
	}
	return vectors
}

func TestCompressMatchesReferenceVectors(t *testing.T) {
	for name, vec := range loadVectors(t) {
		if vec.Input == "" {
			continue
		}
		got := CompressToEncodedURIComponent(vec.Input)
		if got != vec.Compressed {
			t.Errorf("[%s] kompresi berbeda dari referensi:\n got: %s\nwant: %s", name, got, vec.Compressed)
		}
	}
}

func TestDecompressMatchesReferenceVectors(t *testing.T) {
	for name, vec := range loadVectors(t) {
		if vec.Input == "" {
			continue
		}
		got, err := DecompressFromEncodedURIComponent(vec.Compressed)
		if err != nil {
			t.Errorf("[%s] dekompresi error: %v", name, err)
			continue
		}
		if got != vec.Input {
			t.Errorf("[%s] dekompresi tidak cocok:\n got: %q\nwant: %q", name, got, vec.Input)
		}
	}
}

func TestRoundTripRandomPayloads(t *testing.T) {
	payloads := []string{
		`{"metadata":{"code":200,"message":"OK"}}`,
		"A",
		"AA",
		"ABABABABABABABAB",
		strings_repeat("klinik-app bpjs antrean sync payload; ", 50),
		"\u00e9\u00fc\u00f1 emoji \U0001F600 test",
	}
	for i, p := range payloads {
		compressed := CompressToEncodedURIComponent(p)
		decompressed, err := DecompressFromEncodedURIComponent(compressed)
		if err != nil {
			t.Errorf("payload %d: %v", i, err)
			continue
		}
		if decompressed != p {
			t.Errorf("payload %d roundtrip mismatch:\n got %q\nwant %q", i, decompressed, p)
		}
	}
}

func TestDecompressEmpty(t *testing.T) {
	out, err := DecompressFromEncodedURIComponent("")
	if err != nil || out != "" {
		t.Errorf("empty input harus menghasilkan string kosong, got %q, err %v", out, err)
	}
}

// strings_repeat pengganti strings.Repeat agar import tetap minimal.
func strings_repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
