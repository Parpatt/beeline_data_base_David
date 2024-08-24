package services

import (
	"context"
	"encoding/json"
	"fmt"
	"myproject/internal"
	"myproject/internal/database"
	"myproject/internal/jwt"
	"myproject/internal/models"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var deleteMe = true
var deleteMeToo = 27

type MyApp struct {
	app internal.App
}

type LegalUser struct {
	Id            int       `json:"id"`
	Password_hash string    `json:"password_hash"`
	Email         string    `json:"email" db:"email"`
	Phone_number  string    `json:"phone_number"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
	Avatar_path   string    `json:"avatar_path"`
	User_type     int       `json:"user_type"`
	User_role     int       `json:"user_role"`

	Ind_num_taxp    int    `json:"ind_num_taxp"`
	Name_of_company string `json:"name_of_company"`
	Address_name    string `json:"address_name"`
}

func errorr(err error) {
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
}

func NewRepository(pool *pgxpool.Pool) *internal.Repository {
	return &internal.Repository{Pool: pool}
}

func NewApp(Ctx context.Context, dbpool *pgxpool.Pool) *MyApp {
	return &MyApp{internal.App{Ctx: Ctx, Repo: NewRepository(dbpool), Cache: make(map[string]models.User)}}
}

func (a *MyApp) SignupLegalPOST(rw http.ResponseWriter, r *http.Request) {
	var legal LegalUser

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&legal)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.AddNewLegalUserSQL(a.app.Ctx, a.app.Repo.Pool, legal.Ind_num_taxp, legal.Name_of_company, legal.Address_name, legal.Email, legal.Phone_number, legal.Password_hash, rw)

	errorr(err)
}

type NaturUser struct {
	Id            int       `json:"id"`
	Password_hash string    `json:"password_hash"`
	Email         string    `json:"email" db:"email"`
	Phone_number  string    `json:"phone_number"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
	Avatar_path   string    `json:"avatar_path"`
	User_type     int       `json:"user_type"`
	User_role     int       `json:"user_role"`

	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
}

func (a *MyApp) SignupNaturPOST(rw http.ResponseWriter, r *http.Request) {
	var natur NaturUser

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&natur)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.AddNewNaturUserSQL(a.app.Ctx, natur.Surname, natur.Name, natur.Patronymic, natur.Email, natur.Phone_number, natur.Password_hash, a.app.Repo.Pool, rw)

	errorr(err)
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *MyApp) LoginPOST(rw http.ResponseWriter, r *http.Request) {
	var login Login

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	user, err, user_id := repo.LoginSQL(a.app.Ctx, login.Login, login.Password, a.app.Repo.Pool, rw)

	errorr(err)
	if user_id == 0 {
		fmt.Errorf("Incorrect login")
		return
	}

	// логин и пароль совпадают, поэтому генерируем токен, пишем его в кеш и в куки
	validToken, err := jwt.GenerateJWT("имя", user_id)     //получаем токен в строковом типе
	fmt.Println("токен в строковом формате: ", validToken) // токен

	errorr(err)

	a.app.Cache[validToken] = user

	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)

	// кука будет жить 1 час
	cookie := http.Cookie{
		Name:    "token",
		Value:   url.QueryEscape(validToken),
		Expires: expiration,
	}

	http.SetCookie(rw, &cookie)

	fmt.Println(internal.ReadCookie("token", r)) // токен
}

func (a *MyApp) ProductListGET(rw http.ResponseWriter, r *http.Request) {

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.ProductListSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

type PrintAds struct {
	Ads_id int `json:"Ads_id"`
}

func (a *MyApp) PrintAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var printAds PrintAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&printAds)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.PrintAdsSQL(a.app.Ctx, rw, printAds.Ads_id, a.app.Repo.Pool)

	errorr(err)
}

type SortProductList struct {
	Category int `json:"Category"`
}

func (a *MyApp) SortProductListPOST(rw http.ResponseWriter, r *http.Request) {
	var sortProductList SortProductList

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sortProductList)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListSQL(a.app.Ctx, rw, sortProductList.Category, a.app.Repo.Pool)

	errorr(err)
}

type Ads struct {
	// Id          int       `json:"id"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Hourly_rate int    `json:"Hourly_rate"`
	Daily_rate  int    `json:"Daily_rate"`
	// Owner_id    int       `json:"owner_id"`
	Category_id int    `json:"Category_id"`
	Location    string `json:"Location"`
	// Created_at  time.Time `json:"created_at"`

	// Photo_id     int    `json:"photo_id"`
	// Ad_id        int    `json:"ad_id"`
	// File_path    string `json:"file_path"`
	// Status_photo bool   `json:"status_photo"`
}

func (a *MyApp) SignupAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var ads Ads

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)
	if deleteMe {
		err := repo.SignupAdsSQL(a.app.Ctx, ads.Title, ads.Description, ads.Hourly_rate, ads.Daily_rate, deleteMeToo, ads.Category_id, ads.Location, time.Now(), rw, a.app.Repo.Pool)
		if err != nil {
			fmt.Errorf("Ошибка создания объявления: %v", err)
			return
		}
	}
	//}
}
