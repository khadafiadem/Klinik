package bpjs

import (
	"bytes"
	"testing"
)

// Vektor dari dokumentasi resmi BPJS Kesehatan:
// HMAC-SHA256(message="aaa", key="bbb") → base64
func TestSignatureBPJSReferenceVector(t *testing.T) {
	got := hmacBase64("aaa", "bbb")
	want := "20BKS3PWnD3XU4JbSSZvVlGi2WWnDa8Sv9uHJ+wsELA="
	if got != want {
		t.Errorf("signature tidak sesuai dokumen BPJS:\n got: %s\nwant: %s", got, want)
	}

	sig := Signature("10000", "secret", 1570171209)
	if sig == "" || len(sig) != 44 {
		t.Errorf("format signature tidak valid: %q", sig)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	consID, secret, ts := "12345", "rahasia", int64(1700000000)

	key, _, err := deriveKey(consID, secret, ts)
	if err != nil {
		t.Fatalf("deriveKey: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("panjang kunci AES harus 32 byte, dapat %d", len(key))
	}

	payloads := [][]byte{
		[]byte("halo"),
		bytes.Repeat([]byte("x"), 16), // tepat satu blok
		bytes.Repeat([]byte("y"), 32), // tepat dua blok
		[]byte(`{"a":1}`),
	}

	for i, p := range payloads {
		enc, err := Encrypt(consID, secret, ts, p)
		if err != nil {
			t.Fatalf("payload %d encrypt: %v", i, err)
		}
		if enc == string(p) {
			t.Errorf("payload %d: ciphertext sama dengan plaintext", i)
		}
		dec, err := Decrypt(consID, secret, ts, enc)
		if err != nil {
			t.Fatalf("payload %d decrypt: %v", i, err)
		}
		if !bytes.Equal(dec, p) {
			t.Errorf("payload %d roundtrip:\n got %q\nwant %q", i, dec, p)
		}
	}
}

func TestFullPipelineLikeBPJSResponse(t *testing.T) {
	// Simulasi respons BPJS: plaintext → LZString compress → encrypt.
	consID, secret, ts := "29841", "5tG4", int64(1724400000)
	plain := `{"metadata":{"code":1},"response":{"list":[{"nomorkartu":"0001234567890"}]}}`

	compressed := CompressToEncodedURIComponent(plain)
	encrypted, err := Encrypt(consID, secret, ts, []byte(compressed))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	decrypted, err := Decrypt(consID, secret, ts, encrypted)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	decompressed, err := DecompressFromEncodedURIComponent(string(decrypted))
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if decompressed != plain {
		t.Errorf("pipeline tidak cocok:\n got %q\nwant %q", decompressed, plain)
	}
}

func TestDecryptTamperedData(t *testing.T) {
	consID, secret, ts := "c", "s", int64(1)
	enc, _ := Encrypt(consID, secret, ts, []byte("data klinis rahasia"))

	raw := []byte(enc)
	raw[10] ^= 0x01 // rusak satu karakter ciphertext
	if _, err := Decrypt(consID, secret, ts, string(raw)); err == nil {
		t.Error("dekripsi data rusak harus menghasilkan error")
	}
}
