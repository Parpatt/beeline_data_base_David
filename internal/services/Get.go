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
	"github.com/rs/zerolog"
)

type App struct { //структура приложеия
	Ctx   context.Context        //
	Repo  *Repository            //
	Cache map[string]models.User //карта, хранящая User сткуртуру
}

type Repository struct {
	Pool *pgxpool.Pool
}

func (a *MyApp) DisputeChatPanelGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) GroupAdsByHourlyRateGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupAdsByHourlyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupAdsByDailyRateGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err := repo.GroupAdsByDailyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) GroupFavByRecentGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	_, user_id := jwt.IsAuthorized(token)

	if true {
		err := repo.GroupFavByRecentSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id)

		errorr(err)
	}
}

func (a *MyApp) GroupFavByCheaperGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	_, user_id := jwt.IsAuthorized(token)

	if true {
		err := repo.GroupFavByCheaperSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id)

		errorr(err)
	}
}

func (a *MyApp) GroupFavByDearlyGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	_, user_id := jwt.IsAuthorized(token)

	if true {
		err := repo.GroupFavByDearlySQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id)

		errorr(err)
	}
}

func (a *MyApp) GroupAdsByRentedGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)
	if flag {
		err := repo.GroupAdsByRentedSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id)

		errorr(err)
	}
}

func (a *MyApp) GroupAdsByArchivedGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)
	if flag {
		err := repo.GroupAdsByArchivedSQL(a.app.Ctx, rw, user_id, a.app.Repo.Pool)

		errorr(err)
	}
}

func (a *MyApp) GroupOrdersByRentedGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	errorr(err)
	// } else {
	// 	flag, user_id := jwt.IsAuthorized(rw, token)
	// 	if flag {
	err := repo.GroupAdsByRentedSQL(a.app.Ctx, rw, a.app.Repo.Pool, 29)

	errorr(err)
	// 	}
	// }
}

func (a *MyApp) GroupOrdersByUnRentedGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) PrintChatGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	err = repo.PrintChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id)
	errorr(err)
}

func (a *MyApp) BookingListGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Errorf("Ошибка создания объявления: %v", err)
		return
	} else {
		flag, user_id := jwt.IsAuthorized(token)
		if flag {
			err = repo.BookingListSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id)

			errorr(err)
		}
	}
}

func (a *MyApp) RefreshTokenGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	token, err := r.Cookie("Refresh_token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		// Генерация JWT токена
		validToken_jwt, err := jwt.GenerateJWT("jwt", user_id)
		errorr(err)

		// Генерация refresh токена
		refresh_token, err := jwt.GenerateJWT("refresh", user_id)
		errorr(err)

		type User struct {
			Id            int    `json:"id"`
			JWT           string `json:"JWT"`
			Refresh_token string `json:"Refresh_token"`
		}

		type Response struct {
			Status  string `json:"status"`
			Data    User   `json:"data,omitempty"`
			Message string `json:"message"`
		}

		user := User{
			Id:            user_id,
			JWT:           validToken_jwt,
			Refresh_token: refresh_token,
		}

		// Установка куки
		livingTime := 60 * time.Minute
		expiration := time.Now().Add(livingTime)
		cookie := http.Cookie{
			Name:     "token",
			Value:    url.QueryEscape(validToken_jwt),
			Expires:  expiration,
			Path:     "/",             // Убедитесь, что путь корректен
			Domain:   "185.112.83.36", // IP-адрес вашего сервера
			HttpOnly: true,
			Secure:   false, // Для HTTP можно оставить false
			SameSite: http.SameSiteLaxMode,
		}

		// Установка куки
		livingTime = 30 * 24 * time.Hour //не смог найти чего-то получше
		expiration = time.Now().Add(livingTime)
		cookie = http.Cookie{
			Name:     "token",
			Value:    url.QueryEscape(refresh_token),
			Expires:  expiration,
			Path:     "/",             // Убедитесь, что путь корректен
			Domain:   "185.112.83.36", // IP-адрес вашего сервера
			HttpOnly: true,
			Secure:   false, // Для HTTP можно оставить false
			SameSite: http.SameSiteLaxMode,
		}

		fmt.Printf("Кука установлена: %v\n", cookie)

		response := Response{
			Status:  "success",
			Data:    user,
			Message: "You have successfully logged in",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (a *MyApp) WalletListGET(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		err = repo.WalletListSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id)

		errorr(err)
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Что-то не так с токеном",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
}
