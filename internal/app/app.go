package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
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

func (application *MyApp) Routes(r *httprouter.Router, Ctx context.Context, dbpool *pgxpool.Pool, rdb *redis.Client) {
	a := services.NewApp(Ctx, dbpool)

	r.ServeFiles("/public/*filepath", http.Dir("public"))

	r.POST("/signupLegal", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupLegalPOST(rw, r)
	}) //передача данных Юридического лица (регистрация)

	r.POST("/signupNatur", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupNaturPOST(rw, r, rdb)
	}) //передача данных Физического лица (регистрация)

	r.POST("/confEmailForReg", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ConfEmailForRegPOST(rw, r, rdb)
	}) //передача данных Физического лица (подтверждение)

	r.POST("/sendCodForEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodForEmail(rw, r, rdb)
	}) //отправка сообщения на почту для подтверждения

	r.POST("/enterCodFromEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodFromEmail(rw, r, rdb)
	}) //отправка сообщения на почту для подтверждения

	r.POST("/login", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPOST(rw, r)
	}) //логин отправка

	r.POST("/productList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ProductListPOST(rw, r)
	})

	r.POST("/printAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.PrintAdsPOST(rw, r)
	}) //вывод продукта

	r.POST("/sortProductListAll", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListAllPOST(rw, r)
	}) //вывод продукта с учётом сортировки всем категориям

	r.POST("/sigAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupAdsPOST(rw, r)
	}) //размещение(добавление) объявления

	r.POST("/updAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdAdsPOST(rw, r)
	}) //редактирование(изменение) объявления

	r.POST("/delAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DelAdsPOST(rw, r)
	}) //удаление объявления

	r.POST("/sigFavAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigFavAdsPOST(rw, r)
	}) //добавление объявления в избранное

	r.POST("/delFavAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DelFavAdsPOST(rw, r)
	}) //удаление объявления из избранного

	r.POST("/searchForTech", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SearchForTechPOST(rw, r)
	}) //поиск объявления

	r.POST("/sortProductListCategoriez", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListCategoriezPOST(rw, r)
	}) //вывод продукта с учётом сортировки категории

	r.POST("/sigChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigChatPOST(rw, r)
	}) //начало переписки

	r.POST("/openChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.OpenChatPOST(rw, r)
	}) //открытие чата

	r.POST("/sendMessage", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendMessagePOST(rw, r)
	}) //отправить сообщение

	r.POST("/sigDisputInChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigDisputInChatPOST(rw, r)
	}) //начать спор

	r.POST("/sigReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigReviewPOST(rw, r)
	}) //оставить отзыв

	r.POST("/updReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdReviewPOST(rw, r)
	}) //обновить сообщение

	r.GET("/disputeChatPanel", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DisputeChatPanelGET(rw, r)
	}) //показать лист спорных чатов

	r.POST("/mediatorEnterInChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.MediatorEnterInChatPOST(rw, r)
	}) //принять спор на себя(работа медиатора)

	r.POST("/mediatorFinishJob", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.MediatorFinishJobInChatPOST(rw, r)
	}) //медиатор выносит решение

	r.GET("/groupAdsByHourlyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByHourlyRateGET(rw, r)
	}) //группировка объявлений, сначала дороже(почасовая цена)

	r.GET("/groupFavByRecent", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByRecentGET(rw, r)
	}) //группировка избранных объявлений, сначала новые

	r.GET("/groupFavByCheaper", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByCheaperGET(rw, r)
	}) //группировка избранных объявлений, сначала дороже

	r.GET("/groupFavByDearly", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByDearlyGET(rw, r)
	}) //группировка избранных объявлений, сначала дешевле

	r.GET("/groupReviewNewOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewNewOnesFirstGET(rw, r)
	}) //вывод отзывов по порядку, сначала новые

	r.GET("/groupReviewOldOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewOldOnesFirstGET(rw, r)
	}) //вывод отзывов не по порядку, сначала старые

	r.GET("/groupReviewLowRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewLowRatOnesFirstGET(rw, r)
	}) //вывод отзывов, сначала с высокой оценкой

	r.GET("/groupReviewHigRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewHigRatOnesFirstGET(rw, r)
	}) //вывод отзывов, сначала с низкой оценкой

	r.GET("/groupAdsByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByRentedGET(rw, r)
	}) //вывод объявлений по хозяину(активных)

	r.GET("/groupAdsByArchived", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByArchivedGET(rw, r)
	}) //вывод объявлений по хозяину(неактивный)

	r.POST("/transactionToAnother", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToAnotherPOST(rw, r)
	}) //получаем от юзера деньги(или отправляем ему их)

	r.POST("/transactionToSomething", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToSomethingPOST(rw, r)
	}) //получаем возврат денег от системы(возврат по ошибке или что-то похожее)

	r.POST("/transactionToReturnAmount", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToReturnAmountPOST(rw, r)
	}) //платим деньги за что-то системе(не конкретному юзеру)

	r.POST("/registerOrder", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegisterOrderPOST(rw, r)
	}) //регистрация заказа

	r.POST("/regBooking", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegBookingPOST(rw, r)
	}) //переброинрование

	r.POST("/rebookBooking", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RebookBookingPOST(rw, r)
	}) //переброинрование

	r.GET("/groupOrdersByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByRentedGET(rw, r)
	}) //группировка заказов по активным

	r.GET("/groupOrdersByUnRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByUnRentedGET(rw, r)
	}) //группировка заказов по неактивным

	r.POST("/regReport", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegReportPOST(rw, r)
	}) //регистрация репорта

	// r.POST("/printReport", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// a.PrintReportPOST(rw, r)
	// }) //выво

	r.POST("/sendCodeForRecoveryPassWithEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodeForRecoveryPassWithEmailPOST(rw, r, rdb)
	}) //восстановление пароля через почту

	r.POST("/enterCodeForRecoveryPassWithEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodeForRecoveryPassWithEmailPOST(rw, r, rdb)
	}) //восстановление пароля через почту(отправление на почту)

	r.POST("/editingNaturUserData", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EditingNaturUserDataPOST(rw, r)
	}) //изменение данных для физика

	r.POST("/editingLegalUserData", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EditingLegalUserDataPOST(rw, r)
	}) //изменение данных для юрика

	r.POST("/autorizLoginEmailSend", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AutorizLoginEmailSendPOST(rw, r)
	}) //логин отправка

	r.POST("/autorizLoginEmailEnter", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AutorizLoginEmailEnterPOST(rw, r, rdb)
	}) //логин ввод
}

func PageMenuNavigation(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}
