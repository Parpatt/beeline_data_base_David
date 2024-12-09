package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"

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

func (application *MyApp) Routes(r *httprouter.Router, Ctx context.Context, dbpool *pgxpool.Pool, rdb *redis.Client, logger zerolog.Logger) {
	a := services.NewApp(Ctx, dbpool)

	r.ServeFiles("/public/*filepath", http.Dir("public"))

	r.POST("/signupUserByEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupUserByEmailPOST(rw, r, rdb, logger)
	}) //пользователь укзывает почту(регистрация)

	r.POST("/signupUserByPhone", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupUserByPhonePOST(rw, r, rdb, logger)
	}) //пользователь укзывает телефон(регистрация)

	r.POST("/enterCodeFromEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodeFromEmailPOST(rw, r, rdb, logger)
	}) //пользователь укзывает код почта

	r.POST("/enterCodeFromPhone", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodeFromPhonePOST(rw, r, rdb, logger)
	}) //пользователь укзывает код телефон

	r.POST("/signupLegalEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupLegalEmailPOST(rw, r, rdb, logger)
	}) //передача данных Юридического лица (регистрация) Email

	r.POST("/signupLegalPhone", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupLegalPhonePOST(rw, r, rdb, logger)
	}) //передача данных Юридического лица (регистрация) Email

	r.POST("/signupNaturEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupNaturEmailPOST(rw, r, rdb, logger)
	}) //передача данных Физического лица (регистрация) Email

	r.POST("/signupNaturPhone", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupNaturPhonePOST(rw, r, rdb, logger)
	}) //передача данных Физического лица (регистрация) Email

	r.POST("/editingLegalUserData", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EditingLegalUserDataPOST(rw, r, logger)
	}) //изменение данных для юрика

	r.POST("/editingNaturUserData", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EditingNaturUserDataPOST(rw, r, logger)
	}) //изменение данных для физика

	r.POST("/sendCodForEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodForEmailPOST(rw, r, rdb, logger)
	}) //отправка сообщения на почту для подтверждения

	r.POST("/enterCodFromEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodFromEmailPOST(rw, r, rdb, logger)
	}) //отправка сообщения на почту для подтверждения

	r.POST("/sendCodForPhoneNum", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodForPhoneNumPOST(rw, r, rdb, logger)
	}) //отправка сообщения на телефон для подтверждения

	r.POST("/enterCodFromPhoneNum", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodFromPhoneNumPOST(rw, r, rdb, logger)
	}) //отправка сообщения на телефон для подтверждения

	r.POST("/login", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPOST(rw, r, logger)
	}) //логин отправка

	r.POST("/productList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ProductListPOST(rw, r, logger)
	})

	r.POST("/printAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.PrintAdsPOST(rw, r, logger)
	}) //вывод продукта

	r.POST("/sortProductListDailyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListDailyRatePOST(rw, r, logger)
	}) //вывод продукта с учётом сортировки всем категориям

	r.POST("/sortProductListHourlyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListHourlyRatePOST(rw, r, logger)
	}) //вывод продукта с учётом сортировки всем категориям

	r.POST("/sigAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupAdsPOST(rw, r, logger)
	}) //размещение(добавление) объявления

	r.POST("/editAdsList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EditAdsListPOST(rw, r, logger)
	}) //редактирование(изменение) объявления

	r.POST("/updAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdAdsPOST(rw, r, logger)
	}) //редактирование(изменение) объявления

	r.POST("/delAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DelAdsPOST(rw, r, logger)
	}) //удаление объявления

	r.POST("/sigFavAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigFavAdsPOST(rw, r, logger)
	}) //добавление объявления в избранное

	r.POST("/delFavAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DelFavAdsPOST(rw, r, logger)
	}) //удаление объявления из избранного

	r.POST("/searchForTech", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SearchForTechPOST(rw, r, logger)
	}) //поиск объявления

	r.POST("/sortProductListCategoriez", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortProductListCategoriezPOST(rw, r, logger)
	}) //вывод продукта с учётом сортировки категории

	r.POST("/chatButtonInAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ChatButtonInAdsPOST(rw, r, logger)
	}) //кнопка "написать" в листе объявления

	r.POST("/sigChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigChatPOST(rw, r, logger)
	}) //начало переписки

	r.POST("/openChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.OpenChatPOST(rw, r, logger)
	}) //открытие чата

	r.POST("/sendMessageAndImage", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendMessageAndImagePOST(rw, r, logger)
	}) //отправить сообщение и медиа

	r.POST("/sendImage", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendImagePOST(rw, r, logger)
	}) //отправить медиа

	r.POST("/sendMessage", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendMessagePOST(rw, r, logger)
	}) //отправить сообщение

	r.POST("/sendMessageAndVideo", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendMessageAndVideoPOST(rw, r, logger)
	}) //отправить сообщение и медиа

	r.POST("/sendVideo", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendVideoPOST(rw, r, logger)
	}) //отправить сообщение

	r.POST("/sigDisputInChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigDisputInChatPOST(rw, r, logger)
	}) //начать спор
	//чат заканчивается тут

	r.POST("/sigReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SigReviewPOST(rw, r, logger)
	}) //оставить отзыв

	r.POST("/updReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdReviewPOST(rw, r, logger)
	}) //обновить сообщение

	r.GET("/disputeChatPanel", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DisputeChatPanelGET(rw, r, logger)
	}) //показать лист спорных чатов

	r.POST("/mediatorEnterInChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.MediatorEnterInChatPOST(rw, r, logger)
	}) //принять спор на себя(работа медиатора)

	r.POST("/mediatorFinishJob", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.MediatorFinishJobInChatPOST(rw, r, logger)
	}) //медиатор выносит решение

	r.GET("/groupAdsByHourlyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByHourlyRateGET(rw, r, logger)
	}) //группировка объявлений, сначала дороже(почасовая цена)

	r.GET("/groupFavByRecent", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByRecentGET(rw, r, logger)
	}) //группировка избранных объявлений, сначала новые

	r.GET("/groupFavByCheaper", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByCheaperGET(rw, r, logger)
	}) //группировка избранных объявлений, сначала дороже

	r.GET("/groupFavByDearly", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupFavByDearlyGET(rw, r, logger)
	}) //группировка избранных объявлений, сначала дешевле

	r.POST("/groupReviewNewOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewNewOnesFirstPOST(rw, r, logger)
	}) //вывод отзывов по порядку, сначала новые

	r.POST("/groupReviewOldOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewOldOnesFirstPOST(rw, r, logger)
	}) //вывод отзывов не по порядку, сначала старые

	r.POST("/groupReviewLowRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewLowRatOnesFirstPOST(rw, r, logger)
	}) //вывод отзывов, сначала с высокой оценкой

	r.POST("/groupReviewHigRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupReviewHigRatOnesFirstPOST(rw, r, logger)
	}) //вывод отзывов, сначала с низкой оценкой

	r.GET("/groupAdsByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByRentedGET(rw, r, logger)
	}) //вывод объявлений по хозяину(активных)

	r.GET("/groupAdsByArchived", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByArchivedGET(rw, r, logger)
	}) //вывод объявлений по хозяину(неактивный)

	r.POST("/transactionToAnother", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToAnotherPOST(rw, r, logger)
	}) //получаем от юзера деньги(или отправляем ему их)

	r.POST("/transactionToSomething", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToSomethingPOST(rw, r, logger)
	}) //получаем возврат денег от системы(возврат по ошибке или что-то похожее)

	r.POST("/transactionToReturnAmount", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.TransactionToReturnAmountPOST(rw, r, logger)
	}) //платим деньги за что-то системе(не конкретному юзеру)

	r.POST("/regOrderHourly", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegOrderHourlyPOST(rw, r, logger)
	}) //броинрование и оформление заказа

	r.POST("/regOrderDaily", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegOrderDailyPOST(rw, r, logger)
	}) //броинрование и оформление заказа

	r.POST("/bidding", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.BiddingPOST(rw, r, logger)
	}) //пользователи торгуются

	r.POST("/regOrderWithBidding", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegOrderWithBiddingPOST(rw, r, logger)
	}) //броинрование на основе торгов

	r.POST("/rebookOrderHourly", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RebookOrderHourlyPOST(rw, r, logger)
	}) //переброинрование TYT

	r.POST("/rebookOrderDaily", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RebookOrderDailyPOST(rw, r, logger)
	}) //переброинрование TYT

	r.POST("/complBooking", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.ComplBookingPOST(rw, r, logger)
	}) //бронирование прошло успешно и мы начисляем бабки юзеру

	r.GET("/bookingList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.BookingListGET(rw, r, logger)
	}) //переброинрование

	r.GET("/groupOrdersByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByRentedGET(rw, r, logger)
	}) //группировка заказов по активным

	r.GET("/groupOrdersByUnRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByUnRentedGET(rw, r, logger)
	}) //группировка заказов по неактивным

	r.POST("/regReport", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegReportPOST(rw, r, logger)
	}) //регистрация репорта

	// r.POST("/printReport", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// a.PrintReportPOST(rw, r, logger)
	// }) //выво

	r.POST("/sendCodeForRecoveryPassWithEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodeForRecoveryPassWithEmailPOST(rw, r, rdb, logger)
	}) //восстановление пароля через почту

	r.POST("/enterCodeForRecoveryPassWithEmail", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodeForRecoveryPassWithEmailPOST(rw, r, rdb, logger)
	}) //восстановление пароля через почту(отправление на почту)

	r.POST("/sendCodeForRecoveryPassWithPhoneNum", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SendCodeForRecoveryPassWithPhoneNumPOST(rw, r, rdb, logger)
	}) //восстановление пароля через телефон

	r.POST("/enterCodeForRecoveryPassWithPhoneNum", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.EnterCodeForRecoveryPassWithPhoneNumPOST(rw, r, rdb, logger)
	}) //восстановление пароля через телефон

	r.POST("/autorizLoginEmailSend", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AutorizLoginEmailSendPOST(rw, r, rdb, logger)
	}) //логин отправка

	r.POST("/autorizLoginEmailEnter", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AutorizLoginEmailEnterPOST(rw, r, rdb, logger)
	}) //логин ввод

	r.GET("/refreshToken", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RefreshTokenGET(rw, r, logger)
	}) //рефреш токены

	r.GET("/printChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.PrintChatGET(rw, r, logger)
	}) //вывод всех чатов

	r.POST("/allUserAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AllUserAdsPOST(rw, r, logger)
	}) //Кнопка 11 объявлений пользователя

	r.POST("/allAdsOfThisUser", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AllAdsOfThisUserPOST(rw, r, logger)
	}) //все объявления этого юзера

	r.POST("/walletHistory", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.WalletHistoryPOST(rw, r, logger)
	}) //история кошелька

	r.GET("/walletList", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.WalletListGET(rw, r, logger)
	}) //лист кошелька
}
