package application

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func (a app) RegisterRentedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ad_id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("ad_id")))
	renter_id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("renter_id")))
	total_price, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("total_price")))
	starts_at, _ := time.Parse("2006-01-02", r.FormValue("starts_at"))
	ends_at, _ := time.Parse("2006-01-02", r.FormValue("ends_at"))

	err := a.repo.RegisterRentedSQL(a.ctx, rw, ad_id, renter_id, total_price, starts_at, ends_at)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) RebookRentedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id_old_book, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("id_old_book")))
	id_users, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("id_user")))
	starts_at, _ := time.Parse("2006-01-02", r.FormValue("starts_at"))
	ends_at, _ := time.Parse("2006-01-02", r.FormValue("ends_at"))
	amount, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("amount")))

	err := a.repo.RebookRentedGET(a.ctx, rw, id_old_book, id_users, starts_at, ends_at, amount)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupOrdersByRentedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupOrdersByRentedSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupOrdersByUnRentedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupOrdersByUnRentedSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
