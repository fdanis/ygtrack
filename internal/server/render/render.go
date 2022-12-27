package render

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/fdanis/ygtrack/internal/server/config"
	"github.com/fdanis/ygtrack/internal/server/models"
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

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func CreateTemplateCache() (map[string]*template.Template, error) {
	result := map[string]*template.Template{}
	tmps, err := filepath.Glob(basepath + "/../templates/*.html")
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
