package handler

import (
	"html/template"
	"io/fs"
	"testing"
)

// TestAllTemplatesParse memastikan seluruh template dapat diparse
// bersama layout dasar sehingga kesalahan sintaks terdeteksi lebih awal.
func TestAllTemplatesParse(t *testing.T) {
	entries, err := fs.ReadDir(templatesFS, "templates")
	if err != nil {
		t.Fatalf("gagal membaca direktori templates: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		files, err := fs.ReadDir(templatesFS, "templates/"+entry.Name())
		if err != nil {
			t.Fatalf("gagal membaca direktori templates/%s: %v", entry.Name(), err)
		}
		for _, f := range files {
			name := "templates/" + entry.Name() + "/" + f.Name()
			t.Run(name, func(t *testing.T) {
				var tmpl *template.Template
				var err error
				if f.Name() == "print.html" {
					funcs := template.FuncMap{
						"now": func() string { return "" },
						"add": func(a, b int) int { return a + b },
					}
					tmpl, err = template.New("").Funcs(funcs).ParseFS(templatesFS, name)
				} else {
					funcMap := template.FuncMap{
						"upper":   func(s string) string { return s },
						"lower":   func(s string) string { return s },
						"title":   func(s string) string { return s },
						"slice":   func(s string, start, end int) string { return s },
						"now":     func() string { return "" },
						"hasRole": func(u interface{}, names ...string) bool { return false },
					}
					tmpl, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/layouts/base.html", name)
				}
				if err != nil {
					t.Errorf("template gagal diparse: %v", err)
				}
				if tmpl == nil {
					t.Error("template nil")
				}
			})
		}
	}
}
