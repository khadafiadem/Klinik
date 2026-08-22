package handler

import (
	"net/http"

	"klinik-app/app"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	h, err := app.Handler()
	if err != nil {
		http.Error(w, "Kesalahan konfigurasi server", http.StatusInternalServerError)
		return
	}
	if h == nil {
		http.Error(w, "Server tidak tersedia", http.StatusInternalServerError)
		return
	}
	h.ServeHTTP(w, r)
}
