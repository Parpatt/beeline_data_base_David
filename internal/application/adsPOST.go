package application

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func (a app) AddAdsPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	hourly_rate := strings.TrimSpace(r.FormValue("hourly_rate"))
	daily_rate := strings.TrimSpace(r.FormValue("daily_rate"))
	owner_id := 29
	category_id := strings.TrimSpace(r.FormValue("category_id"))
	location := strings.TrimSpace(r.FormValue("location"))
	created_at := time.Now()
	updated_at := time.Now()

	if title == "" || description == "" ||
		hourly_rate == "" || daily_rate == "" ||
		owner_id == 0 || category_id == "" ||
		location == "" {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	int_hourly_rate, _ := strconv.Atoi(hourly_rate)
	int_daily_rate, _ := strconv.Atoi(daily_rate)
	int_category_id, _ := strconv.Atoi(category_id)

	fmt.Println("А тут есть проблема ?")

	err := a.repo.AddAdsSQL(a.ctx, title, description, int_hourly_rate, int_daily_rate, owner_id, int_category_id, location, created_at, updated_at)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) DeleteAdsPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := strings.TrimSpace(r.FormValue("id"))

	if id == "" {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	int_id, _ := strconv.Atoi(id)

	err := a.repo.DeleteAdsSQL(a.ctx, int_id)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) AddFavoritePOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user_id := strings.TrimSpace(r.FormValue("user_id"))
	ads_id := strings.TrimSpace(r.FormValue("ads_id"))

	if user_id == "" || ads_id == "" {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	err := a.repo.AddFavoriteSQL(a.ctx, user_id, ads_id)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) DelFavoritePOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	user_id := strings.TrimSpace(r.FormValue("user_id"))
	ads_id := strings.TrimSpace(r.FormValue("ads_id"))

	if user_id == "" || ads_id == "" {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	err := a.repo.DelFavoriteSQL(a.ctx, user_id, ads_id)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) AddReviewsPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ads_id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("ads_id")))
	user_id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("user_id")))
	rating, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("rating")))
	comment := strings.TrimSpace(r.FormValue("comment"))

	if ads_id == 0 || user_id == 0 || rating == 0 || comment == "" {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	err := a.repo.AddReviewsSQL(a.ctx, ads_id, user_id, rating, comment)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) UpdReviewsPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	review_id, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("review_id")))
	rating, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("rating")))
	comment := strings.TrimSpace(r.FormValue("comment"))

	if rating == 0 || comment == "" || review_id == 0 {
		a.SignupAdsGET(rw, "Все поля должны быть заполнены!")
		return
	}

	err := a.repo.UpdReviewsSQL(a.ctx, review_id, rating, comment)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupByHourlyRatePOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupByHourlyRateSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupByDailyRatePOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupByDailyRateSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupByCategoryPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupByCategorySQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupByLocatPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	locat := strings.TrimSpace(r.FormValue("location"))

	err := a.repo.GroupByLocatSQL(a.ctx, rw, locat)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) SortFavByRecentPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortFavByRecentSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) SortFavByCheaperPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortFavByCheaperSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) SortFavByDearlyPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortFavByDearlySQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) SortReviewNewOnesFirstPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortReviewNewOnesFirstSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
func (a app) SortReviewOldOnesFirstPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortReviewOldOnesFirstSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
func (a app) SortReviewLowRatOnesFirstPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortReviewLowRatOnesFirstSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
func (a app) SortReviewHigRatOnesFirstPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.SortReviewHigRatOnesFirstSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupAdsByRentedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupAdsByRentedSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) GroupAdsByArchivedPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := a.repo.GroupAdsByArchivedSQL(a.ctx, rw)
	if err != nil {
		a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
