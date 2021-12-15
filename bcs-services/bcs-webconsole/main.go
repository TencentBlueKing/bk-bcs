package main

import (
	"embed"
	"fmt"
)

//go:embed web
var f embed.FS

func main() {
	data, _ := f.ReadFile("web/templates/index.html")
	fmt.Println(string(data))
}
