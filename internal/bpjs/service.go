package bpjs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// AntreanRequest adalah payload endpoint /antrean/add.
type AntreanRequest struct {
	NomorKartu      string `json:"nomorkartu"`
	NIK             string `json:"nik"`
	NoHP            string `json:"nohp"`
	KodePoli        string `json:"kodepoli"`
	NamaPoli        string `json:"namapoli"`
	Norm            string `json:"norm"`
	TanggalPeriksa  string `json:"tanggalperiksa"`
	KodeDokter      int64  `json:"kodedokter"`
	NamaDokter      string `json:"namadokter"`
	JamPraktek      string `json:"jampraktek"`
	NomorAntrean    string `json:"nomorantrean"`
	AngkaAntrean    int    `json:"angkaantrean"`
	Keterangan      string `json:"keterangan,omitempty"`
}

// UpdateStatusRequest adalah payload endpoint /antrean/updatestatus.
// Status: 1 = mulai waktu tunggu (dipanggil), 2 = mulai layanan,
// 3 = selesai layanan. Waktu dalam epoch milidetik.
type UpdateStatusRequest struct {
	TanggalPeriksa string `json:"tanggalperiksa"`
	KodePoli       string `json:"kodepoli"`
	NomorKartu     string `json:"nomorkartu"`
	Status         int    `json:"status"`
	Waktu          int64  `json:"waktu"`
}

// BatalRequest adalah payload endpoint /antrean/batal.
type BatalRequest struct {
	TanggalPeriksa string `json:"tanggalperiksa"`
	KodePoli       string `json:"kodepoli"`
	NomorKartu     string `json:"nomorkartu"`
	Alasan         string `json:"alasan"`
}

// Service mengelola konfigurasi, sinkronisasi antrean, dan logging.
type Service struct {
	db     *sql.DB
	logger *log.Logger
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// LoadConfig membaca bpjs_config; environment variable menimpa nilai DB.
func (s *Service) LoadConfig() (*Config, error) {
	cfg := &Config{}
	err := s.db.QueryRow(`SELECT enabled, mode, base_url, cons_id, secret_key, user_key,
		kode_ppk, nama_ppk, kode_poli, nama_poli, jam_praktek FROM bpjs_config WHERE id = 1`).
		Scan(&cfg.Enabled, &cfg.Mode, &cfg.BaseURL, &cfg.ConsID, &cfg.SecretKey,
			&cfg.UserKey, &cfg.KodePPK, &cfg.NamaPPK, &cfg.KodePoli, &cfg.NamaPoli, &cfg.JamPraktek)
	if err != nil {
		return nil, fmt.Errorf("muat konfigurasi BPJS: %w", err)
	}

	if v := os.Getenv("BPJS_CONS_ID"); v != "" {
		cfg.ConsID = v
	}
	if v := os.Getenv("BPJS_SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	}
	if v := os.Getenv("BPJS_USER_KEY"); v != "" {
		cfg.UserKey = v
	}
	if v := os.Getenv("BPJS_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}

	return cfg, nil
}

// SaveConfig menyimpan konfigurasi ke tabel bpjs_config (single row).
func (s *Service) SaveConfig(cfg *Config) error {
	_, err := s.db.Exec(`UPDATE bpjs_config SET enabled=$1, mode=$2, base_url=$3,
		cons_id=$4, secret_key=$5, user_key=$6, kode_ppk=$7, nama_ppk=$8,
		kode_poli=$9, nama_poli=$10, jam_praktek=$11, updated_at=now()
		WHERE id = 1`,
		cfg.Enabled, cfg.Mode, cfg.BaseURL, cfg.ConsID, cfg.SecretKey,
		cfg.UserKey, cfg.KodePPK, cfg.NamaPPK, cfg.KodePoli, cfg.NamaPoli, cfg.JamPraktek)
	if err != nil {
		return fmt.Errorf("simpan konfigurasi BPJS: %w", err)
	}
	return nil
}

func (s *Service) logSync(queueID int, action string, payload interface{}, meta *Metadata, success, sandbox bool) {
	var reqRaw []byte
	if payload != nil {
		reqRaw, _ = json.Marshal(payload)
	}
	var code, msg string
	if meta != nil {
		code = strconv.Itoa(meta.Code)
		msg = meta.Message
	} else if !success {
		msg = "tidak ada respons"
	}
	_, _ = s.db.Exec(`INSERT INTO bpjs_log (queue_id, action, request_payload, response_code, response_message, success, is_sandbox)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, queueID, action, reqRaw, code, msg, success, sandbox)
}

// queueInfo memuat data antrean + pasien + mapping dokter untuk sync.
type queueInfo struct {
	QueueID     int
	QueueNumber string
	QueueDate   string
	PatientID   int
	PatientName string
	MRN         string
	NIK         string
	Phone       string
	Insurance   string
	CardNumber  string
	DoctorID    int
	DoctorName  string
	BPJSCode    sql.NullString
}

func (s *Service) loadQueue(queueID int) (*queueInfo, error) {
	qi := &queueInfo{QueueID: queueID}
	err := s.db.QueryRow(`SELECT q.queue_number, q.queue_date::text, q.patient_id,
			p.full_name, p.medical_record_number, COALESCE(p.nik,''), COALESCE(p.phone,''),
			COALESCE(p.insurance_name,''), COALESCE(p.insurance_number,''),
			q.doctor_id, d.full_name, m.bpjs_code
		FROM queues q
		JOIN patients p ON q.patient_id = p.id
		JOIN doctors d ON q.doctor_id = d.id
		LEFT JOIN bpjs_doctor_map m ON m.doctor_id = d.id
		WHERE q.id = $1`, queueID).
		Scan(&qi.QueueNumber, &qi.QueueDate, &qi.PatientID,
			&qi.PatientName, &qi.MRN, &qi.NIK, &qi.Phone,
			&qi.Insurance, &qi.CardNumber,
			&qi.DoctorID, &qi.DoctorName, &qi.BPJSCode)
	if err != nil {
		return nil, fmt.Errorf("muat antrean %d: %w", queueID, err)
	}
	return qi, nil
}

func (qi *queueInfo) isBPJSPatient() bool {
	return strings.Contains(strings.ToLower(qi.InsureName()), "bpjs")
}

func (qi *queueInfo) InsureName() string { return qi.Insurance }

func (s *Service) shouldSync(cfg *Config, qi *queueInfo) (bool, string) {
	if !cfg.Enabled {
		return false, "integrasi nonaktif"
	}
	if !qi.isBPJSPatient() {
		return false, "pasien bukan peserta BPJS"
	}
	if qi.CardNumber == "" {
		return false, "nomor kartu BPJS pasien kosong"
	}
	return true, ""
}

func epochMillis() int64 { return time.Now().UnixMilli() }

func parseAngkaAntrean(num string) int {
	digits := ""
	for _, r := range num {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	n, _ := strconv.Atoi(digits)
	return n
}

// OnQueueCreated mengirim antrean baru ke BPJS (best-effort).
func (s *Service) OnQueueCreated(queueID int) {
	cfg, err := s.LoadConfig()
	if err != nil {
		s.logSync(queueID, "add", nil, nil, false, false)
		return
	}

	qi, err := s.loadQueue(queueID)
	if err != nil {
		s.logSync(queueID, "add", nil, nil, false, cfg.Mode == ModeSandbox)
		return
	}

	if ok, reason := s.shouldSync(cfg, qi); !ok {
		s.logSync(queueID, "skip:"+reason, nil, &Metadata{Code: 0, Message: reason}, true, cfg.Mode == ModeSandbox)
		return
	}

	payload := &AntreanRequest{
		NomorKartu:     qi.CardNumber,
		NIK:            qi.NIK,
		NoHP:           qi.Phone,
		KodePoli:       cfg.KodePoli,
		NamaPoli:       cfg.NamaPoli,
		Norm:           qi.MRN,
		TanggalPeriksa: qi.QueueDate,
		NamaDokter:     qi.DoctorName,
		JamPraktek:     cfg.JamPraktek,
		NomorAntrean:   qi.QueueNumber,
		AngkaAntrean:   parseAngkaAntrean(qi.QueueNumber),
	}
	if qi.BPJSCode.Valid {
		if code, err := strconv.ParseInt(strings.TrimSpace(qi.BPJSCode.String), 10, 64); err == nil {
			payload.KodeDokter = code
		}
	}

	meta, err := s.send(cfg, "POST", "antrean/add", payload)
	success := meta != nil && meta.Code == 200 && err == nil
	s.logSync(queueID, "add", payload, meta, success, cfg.Mode == ModeSandbox)
}

// OnQueueStatusChanged mengirim perubahan status antrean ke BPJS.
func (s *Service) OnQueueStatusChanged(queueID int, status int) {
	cfg, err := s.LoadConfig()
	if err != nil {
		s.logSync(queueID, fmt.Sprintf("updatestatus:%d", status), nil, nil, false, false)
		return
	}

	qi, err := s.loadQueue(queueID)
	if err != nil {
		s.logSync(queueID, fmt.Sprintf("updatestatus:%d", status), nil, nil, false, cfg.Mode == ModeSandbox)
		return
	}

	if ok, reason := s.shouldSync(cfg, qi); !ok {
		s.logSync(queueID, "skip:"+reason, nil, &Metadata{Code: 0, Message: reason}, true, cfg.Mode == ModeSandbox)
		return
	}

	payload := &UpdateStatusRequest{
		TanggalPeriksa: qi.QueueDate,
		KodePoli:       cfg.KodePoli,
		NomorKartu:     qi.CardNumber,
		Status:         status,
		Waktu:          epochMillis(),
	}

	meta, err := s.send(cfg, "POST", "antrean/updatestatus", payload)
	action := fmt.Sprintf("updatestatus:%d", status)
	success := meta != nil && meta.Code == 200 && err == nil
	s.logSync(queueID, action, payload, meta, success, cfg.Mode == ModeSandbox)
}

// OnQueueCancelled memberitahu BPJS antrean dibatalkan.
func (s *Service) OnQueueCancelled(queueID int, reason string) {
	cfg, err := s.LoadConfig()
	if err != nil {
		s.logSync(queueID, "batal", nil, nil, false, false)
		return
	}

	qi, err := s.loadQueue(queueID)
	if err != nil {
		s.logSync(queueID, "batal", nil, nil, false, cfg.Mode == ModeSandbox)
		return
	}

	if ok, skipReason := s.shouldSync(cfg, qi); !ok {
		s.logSync(queueID, "skip:"+skipReason, nil, &Metadata{Code: 0, Message: skipReason}, true, cfg.Mode == ModeSandbox)
		return
	}

	payload := &BatalRequest{
		TanggalPeriksa: qi.QueueDate,
		KodePoli:       cfg.KodePoli,
		NomorKartu:     qi.CardNumber,
		Alasan:         reason,
	}

	meta, err := s.send(cfg, "POST", "antrean/batal", payload)
	success := meta != nil && meta.Code == 200 && err == nil
	s.logSync(queueID, "batal", payload, meta, success, cfg.Mode == ModeSandbox)
}

// send mengirim request sesuai mode; SANDBOX tidak menyentuh jaringan.
func (s *Service) send(cfg *Config, method, path string, body interface{}) (*Metadata, error) {
	if cfg.Mode != ModeProduction {
		meta := SandboxResponse(path)
		return &meta, nil
	}
	client := NewClient(*cfg)
	return client.Call(method, path, body, nil)
}

// RecentLogs mengambil N log terakhir untuk halaman admin.
type SyncLog struct {
	ID        int64     `json:"id"`
	QueueID   *int      `json:"queue_id"`
	Action    string    `json:"action"`
	ResponseCode *string `json:"response_code"`
	Message   *string   `json:"response_message"`
	Success   bool      `json:"success"`
	Sandbox   bool      `json:"is_sandbox"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Service) RecentLogs(limit int) ([]SyncLog, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(`SELECT id, queue_id, action, response_code, response_message, success, is_sandbox, created_at
		FROM bpjs_log ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var l SyncLog
		if err := rows.Scan(&l.ID, &l.QueueID, &l.Action, &l.ResponseCode, &l.Message, &l.Success, &l.Sandbox, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// DoctorMap merepresentasikan mapping dokter internal → kode BPJS.
type DoctorMap struct {
	DoctorID   int    `json:"doctor_id"`
	DoctorName string `json:"doctor_name"`
	BPJSCode   string `json:"bpjs_code"`
}

// ListDoctorMaps mengembalikan semua mapping beserta daftar dokter
// yang belum dimapping (BPJSCode kosong).
func (s *Service) ListDoctorMaps() ([]DoctorMap, error) {
	rows, err := s.db.Query(`SELECT d.id, d.full_name, COALESCE(m.bpjs_code, '')
		FROM doctors d
		LEFT JOIN bpjs_doctor_map m ON m.doctor_id = d.id
		ORDER BY d.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var maps []DoctorMap
	for rows.Next() {
		var dm DoctorMap
		if err := rows.Scan(&dm.DoctorID, &dm.DoctorName, &dm.BPJSCode); err != nil {
			return nil, err
		}
		maps = append(maps, dm)
	}
	return maps, nil
}

// UpsertDoctorMap menyimpan atau memperbarui kode BPJS seorang dokter.
func (s *Service) UpsertDoctorMap(doctorID int, code string) error {
	if doctorID <= 0 || code == "" {
		return fmt.Errorf("dokter dan kode BPJS wajib diisi")
	}
	_, err := s.db.Exec(`INSERT INTO bpjs_doctor_map (doctor_id, bpjs_code)
		VALUES ($1, $2)
		ON CONFLICT (doctor_id) DO UPDATE SET bpjs_code = EXCLUDED.bpjs_code, updated_at = now()`,
		doctorID, code)
	if err != nil {
		return fmt.Errorf("simpan mapping dokter: %w", err)
	}
	return nil
}
