package services

import (
	"context"
	"myproject/internal/database"
	"myproject/internal/models"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct { //структура приложеия
	Ctx   context.Context        //
	Repo  *Repository            //
	Cache map[string]models.User //карта, хранящая User сткуртуру
}

type Repository struct {
	Pool *pgxpool.Pool
}

func (a *MyApp) DisputeChatPanelGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err := repo.DisputeChatPanelSQL(a.app.Ctx, a.app.Repo.Pool, rw)

	errorr(err)
}

func (a *MyApp) GroupAdsByHourlyRateGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupAdsByHourlyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupAdsByDailyRateGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupAdsByDailyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupFavByRecentGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupFavByRecentSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupFavByCheaperGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupFavByCheaperSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupFavByDearlyGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupFavByDearlySQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupReviewNewOnesFirstGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupReviewNewOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupReviewOldOnesFirstGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupReviewOldOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupReviewLowRatOnesFirstGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupReviewLowRatOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupReviewHigRatOnesFirstGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupReviewHigRatOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupAdsByRentedGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	errorr(err)
	// } else {
	// flag, user_id := jwt.IsAuthorized(rw, token)
	// if flag {
	err := repo.GroupAdsByRentedSQL(a.app.Ctx, rw, 29, a.app.Repo.Pool)

	errorr(err)
	// }
	// }
}

func (a *MyApp) GroupAdsByArchivedGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	errorr(err)
	// } else {
	// 	flag, user_id := jwt.IsAuthorized(rw, token)
	// 	if flag {
	err := repo.GroupAdsByArchivedSQL(a.app.Ctx, rw, 29, a.app.Repo.Pool)

	errorr(err)
	// 	}
	// }
}

func (a *MyApp) GroupOrdersByRentedGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	errorr(err)
	// } else {
	// 	flag, user_id := jwt.IsAuthorized(rw, token)
	// 	if flag {
	err := repo.GroupAdsByArchivedSQL(a.app.Ctx, rw, 29, a.app.Repo.Pool)

	errorr(err)
	// 	}
	// }
}

func (a *MyApp) GroupOrdersByUnRentedGET(rw http.ResponseWriter, r *http.Request) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	errorr(err)
	// } else {
	// 	flag, user_id := jwt.IsAuthorized(rw, token)
	// 	if flag {
	err := repo.GroupAdsByArchivedSQL(a.app.Ctx, rw, 29, a.app.Repo.Pool)

	errorr(err)
	// 	}
	// }
}
