package application

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (a app) SignupNaturGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "users", "signupNatur.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "signup", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) SignupLegalGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "users", "signupLegal.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "signup", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
