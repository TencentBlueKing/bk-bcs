package web

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates static
var FS embed.FS

// WebTemplate html 摸版
func WebTemplate() *template.Template {
	tpl := template.Must(template.New("").ParseFS(FS, "templates/*.html"))
	return tpl
}

// WebStatic 静态资源
func WebStatic() fs.FS {
	static, _ := fs.Sub(FS, "static")
	return static
}
