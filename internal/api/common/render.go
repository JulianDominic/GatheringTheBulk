package common

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/JulianDominic/GatheringTheBulk/internal/store"
)

type Renderer struct {
	Store store.Store
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, tmplName string, data interface{}) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	files := []string{
		filepath.Join("web/templates", tmplName),
	}

	// Only include layout if NOT an HTMX request
	renderTmpl := tmplName
	if req.Header.Get("HX-Request") != "true" {
		files = append(files, "web/templates/layout.html")
		renderTmpl = "layout.html"
	}

	tmpl, err := template.New(renderTmpl).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		http.Error(w, "Template Parsing Error: "+err.Error(), http.StatusInternalServerError)
		log.Println("Template parse error:", err)
		return
	}

	if req.Header.Get("HX-Request") == "true" {
		err = tmpl.ExecuteTemplate(w, "content", data)
	} else {
		err = tmpl.Execute(w, data)
	}

	if err != nil {
		log.Println("Render error:", err)
		http.Error(w, "Template Execution Error", http.StatusInternalServerError)
	}
}

func (r *Renderer) RenderPartial(w http.ResponseWriter, tmplName string, data interface{}) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}

	tmpl, err := template.New(filepath.Base(tmplName)).Funcs(funcMap).ParseFiles(filepath.Join("web/templates", tmplName))
	if err != nil {
		http.Error(w, "Template Parsing Error: "+err.Error(), http.StatusInternalServerError)
		log.Println("Partial parse error:", err)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Partial render error:", err)
	}
}
