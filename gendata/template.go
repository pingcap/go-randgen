package gendata

import (
	"log"
	"text/template"
)

func mustParse(name string, templateText string) *template.Template {
	tmpl, err := template.New(name).Parse(templateText)
	if err != nil  {
		log.Fatalf("template %s parse fail, %v \n", templateText, err)
	}

	return tmpl
}
