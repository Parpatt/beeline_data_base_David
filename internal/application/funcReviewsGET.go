package application

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (a app) AddReviewsGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "addReview.html")

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

func (a app) UpdReviewsGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "updReview.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "update", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) SortReviewNewOnesFirstGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "sortReviewNewOnesFirst.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) SortReviewOldOnesFirstGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "sortReviewOldOnesFirst.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) SortReviewLowRatOnesFirstGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "sortReviewLowRatOnesFirst.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) SortReviewHigRatOnesFirstGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "reviews", "sortReviewHigRatOnesFirst.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
