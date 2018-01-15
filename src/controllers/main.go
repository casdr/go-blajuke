package controllers

import (
	"html/template"
	"net/http"
)

func Register(templates *template.Template) {
	hc := new(HomeController)
	hc.template = templates.Lookup("index.html")
	http.HandleFunc("/", hc.Get)

	st := http.FileServer(http.Dir("public"))
	http.Handle("/assets/", st)
}