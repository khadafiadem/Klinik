package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"klinik-app/internal/auth"
)

var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// parseReportRange reads from/to query params, defaulting to the current month.
func parseReportRange(r *http.Request) (from, to string) {
	from = r.URL.Query().Get("from")
	to = r.URL.Query().Get("to")
	today := time.Now().Format("2006-01-02")
	if !datePattern.MatchString(from) {
		from = time.Now().Format("2006-01") + "-01"
	}
	if !datePattern.MatchString(to) {
		to = today
	}
	if from > to {
		from, to = to, from
	}
	return from, to
}

func respondCSV(w http.ResponseWriter, filename string, header []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	// UTF-8 BOM so Excel renders Indonesian text correctly
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	cw := csv.NewWriter(w)
	cw.Write(header)
	cw.WriteAll(rows)
	cw.Flush()
}

func exportFilename(name string) string {
	return fmt.Sprintf("%s_%s.csv", name, url.PathEscape(time.Now().Format("2006-01-02")))
}

func (h *WebHandler) ExportPatientsCSV(w http.ResponseWriter, r *http.Request, user *auth.User) {
	rows, err := h.rptSvc.GetPatientRows(parseReportRange(r))
	if err != nil {
		http.Error(w, "Gagal membuat laporan", http.StatusInternalServerError)
		return
	}
	data := make([][]string, 0, len(rows))
	for _, v := range rows {
		data = append(data, []string{v.MRN, v.FullName, v.Gender, v.BirthDate, v.Phone, v.Insurance, v.CreatedAt})
	}
	header := []string{"No. RM", "Nama Pasien", "Jenis Kelamin", "Tanggal Lahir", "Telepon", "Asuransi", "Terdaftar"}
	respondCSV(w, exportFilename("laporan_pasien"), header, data)
}

func (h *WebHandler) ExportRegistrationsCSV(w http.ResponseWriter, r *http.Request, user *auth.User) {
	rows, err := h.rptSvc.GetVisitRows(parseReportRange(r))
	if err != nil {
		http.Error(w, "Gagal membuat laporan", http.StatusInternalServerError)
		return
	}
	data := make([][]string, 0, len(rows))
	for _, v := range rows {
		data = append(data, []string{v.RegNumber, v.RegDate, v.PatientMRN, v.Patient, v.Doctor, v.RegType, v.Status})
	}
	header := []string{"No. Pendaftaran", "Tanggal", "No. RM", "Pasien", "Dokter", "Tipe", "Status"}
	respondCSV(w, exportFilename("laporan_pendaftaran"), header, data)
}

func (h *WebHandler) ExportRevenueCSV(w http.ResponseWriter, r *http.Request, user *auth.User) {
	detail, err := h.rptSvc.GetRevenueDetail(parseReportRange(r))
	if err != nil || detail == nil {
		http.Error(w, "Gagal membuat laporan", http.StatusInternalServerError)
		return
	}
	data := make([][]string, 0, len(detail.Rows)+1)
	for _, v := range detail.Rows {
		data = append(data, []string{
			v.PaymentNumber, v.PaymentDate, v.InvoiceNumber, v.Patient,
			v.Method, v.Status, strconv.FormatFloat(v.Amount, 'f', 2, 64),
		})
	}
	data = append(data, []string{"", "", "", "", "", "TOTAL SELESAI", strconv.FormatFloat(detail.Total, 'f', 2, 64)})
	header := []string{"No. Pembayaran", "Tanggal", "No. Invoice", "Pasien", "Metode", "Status", "Jumlah (Rp)"}
	respondCSV(w, exportFilename("laporan_pendapatan"), header, data)
}
