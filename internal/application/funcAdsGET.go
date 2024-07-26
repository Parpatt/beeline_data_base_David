package application

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (a app) SignupAdsGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "sigAds.html")

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

func (a app) DeleteAdsGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "delAds.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "delete", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) AddFavoriteGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "addFavorite.html")

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

func (a app) DelFavoriteGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "delFavorite.html")

	tmpl, err := template.ParseFiles(sp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "delete", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) GroupByHourlyRateGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupByHourlyRate.html")

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

func (a app) GroupByDailyRateGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupByDailyRate.html")

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

func (a app) GroupByCategoryGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupByCategory.html")

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

func (a app) GroupByLocatGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupByLocat.html")

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

func (a app) SortFavByRecentGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "sortFavByRecent.html")

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

func (a app) SortFavByCheaperGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "sortFavByCheaper.html")

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

func (a app) SortFavByDearlyGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "sortFavByDearly.html")

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

func (a app) GroupAdsByRentedGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupAdsByRented.html")

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

func (a app) GroupAdsByArchivedGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "ads", "groupAdsByArchived.html")

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
