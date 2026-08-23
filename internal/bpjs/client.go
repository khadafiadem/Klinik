package bpjs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Mode operasi integrasi.
const (
	ModeSandbox    = "SANDBOX"
	ModeProduction = "PRODUCTION"
)

type Config struct {
	Enabled    bool   `json:"enabled"`
	Mode       string `json:"mode"`
	BaseURL    string `json:"base_url"`
	ConsID     string `json:"cons_id"`
	SecretKey  string `json:"secret_key"`
	UserKey    string `json:"user_key"`
	KodePPK    string `json:"kode_ppk"`
	NamaPPK    string `json:"nama_ppk"`
	KodePoli   string `json:"kode_poli"`
	NamaPoli   string `json:"nama_poli"`
	JamPraktek string `json:"jam_praktek"`
}

// Metadata adalah bagian status dari envelope respons BPJS.
type Metadata struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Envelope adalah struktur bungkus respons seluruh WS BPJS.
type Envelope struct {
	Metadata Metadata          `json:"metadata"`
	Response json.RawMessage   `json:"response"`
	raw      map[string]string `json:"-"`
}

// Client HTTP untuk WS BPJS.
type Client struct {
	httpClient *http.Client
	cfg        Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		cfg:        cfg,
	}
}

func (c *Client) Config() Config { return c.cfg }

// SandboxResponse membuat metadata sukses tiruan untuk mode SANDBOX.
func SandboxResponse(action string) Metadata {
	return Metadata{Code: 200, Message: "OK (sandbox): " + action}
}

// Call mengirim request ke WS BPJS, memvalidasi header, lalu mendekripsi
// dan mendekompresi respons. out boleh nil bila payload respons diabaikan.
func (c *Client) Call(method, path string, body interface{}, out interface{}) (*Metadata, error) {
	if c.cfg.ConsID == "" || c.cfg.SecretKey == "" {
		return nil, fmt.Errorf("konfigurasi BPJS belum lengkap (cons_id/secret_key kosong)")
	}

	ts := CurrentTimestamp()
	sign := Signature(c.cfg.ConsID, c.cfg.SecretKey, ts)

	url := fmt.Sprintf("%s/%s", c.cfg.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = nil
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("buat request: %w", err)
	}
	req.Header.Set("X-cons-id", c.cfg.ConsID)
	req.Header.Set("X-timestamp", fmt.Sprintf("%d", ts))
	req.Header.Set("X-signature", sign)
	req.Header.Set("user_key", c.cfg.UserKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request gagal: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("baca respons: %w", err)
	}

	var env struct {
		Metadata Metadata        `json:"metadata"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(respBytes, &env); err != nil {
		return nil, fmt.Errorf("parse envelope (HTTP %d): %w", resp.StatusCode, err)
	}

	if env.Metadata.Code < 200 || env.Metadata.Code > 299 {
		return &env.Metadata, fmt.Errorf("BPJS %d: %s", env.Metadata.Code, env.Metadata.Message)
	}

	// Respons tanpa field terenkripsi (beberapa endpoint referensi lama).
	if len(env.Response) == 0 || string(env.Response) == "null" {
		return &env.Metadata, nil
	}

	var encrypted string
	if err := json.Unmarshal(env.Response, &encrypted); err != nil {
		return nil, fmt.Errorf("field response bukan string terenkripsi: %w", err)
	}

	decrypted, err := Decrypt(c.cfg.ConsID, c.cfg.SecretKey, ts, encrypted)
	if err != nil {
		return nil, fmt.Errorf("dekripsi: %w", err)
	}

	plain, err := DecompressFromEncodedURIComponent(string(decrypted))
	if err != nil {
		return nil, fmt.Errorf("dekompresi: %w", err)
	}

	if out != nil {
		if err := json.Unmarshal([]byte(plain), out); err != nil {
			return nil, fmt.Errorf("parse data respons: %w (data: %.200s)", err, plain)
		}
	}

	return &env.Metadata, nil
}
