package mailer

import (
	"bytes"
	"html/template"
	"path/filepath"
)

func RenderTemplate(filePath string, data interface{}) (string, error) {
	tpl, err := template.New("").Funcs(template.FuncMap{
		"safe":     func(s string) template.HTML { return template.HTML(s) },
		"notEmpty": func(s string) bool { return s != "" },
	}).ParseFiles(filePath)

	if err != nil {
		return "", err
	}

	// ParseFiles registers templates by their base filename, not the full path
	templateName := filepath.Base(filePath)
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
