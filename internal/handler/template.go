package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
)

//go:embed templates
var templatesFS embed.FS

type TemplateData struct {
	User       *auth.User
	ActivePage string
	Data       interface{}
	Data2      interface{}
	Error      string
	Success    string
	Search     string
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, data TemplateData) {
	if data.ActivePage == "" {
		data.ActivePage = extractPageName(r.URL.Path)
	}

	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"slice": func(s string, start, end int) string {
			if len(s) <= start {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"now": func() string {
			return time.Now().Format("2006-01-02")
		},
		"hasRole": func(u *auth.User, names ...string) bool {
			if u == nil {
				return false
			}
			for _, role := range u.Roles {
				for _, n := range names {
					if strings.EqualFold(role.Name, n) {
						return true
					}
				}
			}
			return false
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/layouts/base.html", "templates/"+tmplName+".html")
	if err != nil {
		logger.Error.Printf("Gagal parse template %s: %v", tmplName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		logger.Error.Printf("Gagal render template %s: %v", tmplName, err)
	}
}

func RenderPrint(w http.ResponseWriter, tmplName string, data interface{}) {
	funcs := template.FuncMap{
		"now": func() string {
			return time.Now().Format("02 January 2006")
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "templates/"+tmplName+".html")
	if err != nil {
		logger.Error.Printf("Gagal parse template cetak %s: %v", tmplName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "print", data); err != nil {
		logger.Error.Printf("Gagal render template cetak %s: %v", tmplName, err)
	}
}

func extractPageName(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "dashboard"
	}
	parts := strings.Split(path, "/")
	return parts[0]
}

// RenderForbidden menampilkan halaman 403 untuk user tanpa izin.
func RenderForbidden(w http.ResponseWriter, r *http.Request, user *auth.User) {
	logger.Error.Printf("Akses ditolak: user %s mencoba mengakses %s", user.Username, r.URL.Path)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, `<!DOCTYPE html><html lang="id"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>403 - Akses Ditolak - KlinikApp</title>
<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<link rel="stylesheet" href="/static/css/style.css">
</head><body style="display:flex;align-items:center;justify-content:center;min-height:100vh;background:var(--bg-primary);font-family:'Inter',sans-serif;">
<div style="text-align:center;padding:40px;">
<div style="font-size:6rem;font-weight:800;color:var(--danger);line-height:1;">403</div>
<h1 style="margin:16px 0 8px;font-size:1.5rem;color:var(--text-primary);">Akses Ditolak</h1>
<p style="color:var(--text-muted);margin-bottom:24px;">Anda tidak memiliki kewenangan untuk mengakses halaman ini.</p>
<a href="/dashboard" class="btn btn-primary">Kembali ke Dashboard</a>
</div></body></html>`)
}
