package config

import "html/template"

type AppConfig struct {
	UseTemplateCache bool
	TemplateCache    map[string]*template.Template
}
