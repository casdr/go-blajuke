package controllers

import (
	"html/template"
	"net/http"
)

type HomeController struct {
	template * template.Template
}

func (this *HomeController) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content Type", "text/html")
	this.template.Execute(w, new(interface{}))
}