package render

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/fdanis/ygtrack/cmd/server/config"
	"github.com/fdanis/ygtrack/cmd/server/models"
)

var funcMap = template.FuncMap{}
var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultDate(td *models.TemplateDate) *models.TemplateDate {
	return td
}

func Render(w http.ResponseWriter, templateName string, data *models.TemplateDate) {
	var ts map[string]*template.Template
	if app.UseTemplateCache {
		ts = app.TemplateCache
	} else {
		ts, _ = CreateTemplateCache()
	}
	t, ok := ts[templateName]
	if !ok {
		log.Fatalf("could not get template %d", len(ts))
	}
	data = AddDefaultDate(data)
	_ = t.Execute(w, data)
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	result := map[string]*template.Template{}
	tmps, err := filepath.Glob("./cmd/server/templates/*.html")
	if err != nil {
		return result, err
	}
	for _, tmp := range tmps {
		name := filepath.Base(tmp)

		ts, err := template.New(name).Funcs(funcMap).ParseFiles(tmp)
		if err != nil {
			return result, err
		}

		result[name] = ts
	}
	return result, nil
}
