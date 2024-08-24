package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"

	"myproject/internal"
	"myproject/internal/models"
	"myproject/internal/services"
)

type MyApp struct {
	app internal.App
}

func NewRepository(pool *pgxpool.Pool) *internal.Repository {
	return &internal.Repository{Pool: pool}
}

func NewApp(Ctx context.Context, dbpool *pgxpool.Pool) *MyApp {
	return &MyApp{internal.App{Ctx: Ctx, Repo: NewRepository(dbpool), Cache: make(map[string]models.User)}}
}

func StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}

func (application *MyApp) Routes(r *httprouter.Router, Ctx context.Context, dbpool *pgxpool.Pool) {
	a := services.NewApp(Ctx, dbpool)

	r.ServeFiles("/public/*filepath", http.Dir("public"))

	r.POST("/signupLegal", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupLegalPOST(rw, r)
	}) //передача данных Юридического лица 		(регистрация)

	r.POST("/signupNatur", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupNaturPOST(rw, r)
	}) //передача данных Физического лица 		(регистрация)

	r.POST("/login", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPOST(rw, r)
	}) //логин

	r.GET("/productList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ProductListGET(rw, r)
	}) //вывод продукта

	r.POST("/printAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.PrintAdsPOST(rw, r)
	}) //вывод продукта

	r.POST("/sortProductList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListPOST(rw, r)
	}) //вывод продукта с учётом сортировки по категориям

	r.POST("/sigAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupAdsPOST(rw, r)
	}) //размещение(добавление) объявления
}

func PageMenuNavigation(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}
