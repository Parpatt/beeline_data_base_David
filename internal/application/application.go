package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"

	"myproject/internal/repository"
)

// "SELECT id, hourly_rate FROM Ads.ads GROUP BY id ORDER BY hourly_rate DESC;",
type app struct { //структура приложеия
	ctx   context.Context            //
	repo  *repository.Repository     //
	cache map[string]repository.User //карта, хранящая User сткуртуру
}

func (a app) Routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))

	r.GET("/", a.authorized(a.StartPage)) //вызывает функцию authorized(a.StartPage) при адресе "/"
	r.GET("/login",                       //вызывает функцию LoginPage() при адресе "/login"
		func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			a.LoginGET(rw, "")
		},
	)
	r.GET("/logout", a.Logout) //вызывает функцию Logout при адресе "/logout"
	r.GET("/signupNatur",      //вызывает функцию SignupNaturPage при адресе "/signupNatur"
		func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			a.SignupNaturGET(rw, "")
		},
	)
	r.GET("/signupLegal", //вызывает функцию SignupLegalPage при адресе "/signupLegal"
		func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			a.SignupLegalGET(rw, "")
		},
	)
	r.GET("/pageMenu", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.PageMenuGET(rw, "")
	})
	r.GET("/sigAds", //вызывает функцию SignupAdsPage при адресе "/sigAds"
		func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			a.SignupAdsGET(rw, "")
		},
	)
	r.GET("/delAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DeleteAdsGET(rw, "")
	})
	r.GET("/addFavorite", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AddFavoriteGET(rw, "")
	})
	r.GET("/delFavorite", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.DelFavoriteGET(rw, "")
	})
	r.GET("/addReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.AddReviewsGET(rw, "")
	})
	r.GET("/updReview", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.UpdReviewsGET(rw, "")
	})
	r.GET("/groupByHourlyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupByHourlyRateGET(rw, "")
	})
	r.GET("/groupByDailyRate", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupByDailyRateGET(rw, "")
	})
	r.GET("/groupByCategory", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupByCategoryGET(rw, "")
	})
	r.GET("/groupByLocation", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupByLocatGET(rw, "")
	})
	r.GET("/sortFavByRecent", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortFavByRecentGET(rw, "")
	})
	r.GET("/sortFavByCheaper", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortFavByCheaperGET(rw, "")
	})
	r.GET("/sortFavByDearly", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortFavByDearlyGET(rw, "")
	})
	r.GET("/sortReviewNewOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortReviewNewOnesFirstGET(rw, "")
	})
	r.GET("/sortReviewOldOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortReviewOldOnesFirstGET(rw, "")
	})
	r.GET("/sortReviewLowRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortReviewLowRatOnesFirstGET(rw, "")
	})
	r.GET("/sortReviewHigRatOnesFirst", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SortReviewHigRatOnesFirstGET(rw, "")
	})
	r.GET("/groupAdsByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByRentedGET(rw, "")
	})
	r.GET("/groupAdsByArchived", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupAdsByArchivedGET(rw, "")
	})
	r.GET("/registerRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegisterRentedGET(rw, "")
	})
	r.GET("/rebookRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RebookRentedGET(rw, "")
	})
	r.GET("/groupOrdersByRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByRentedGET(rw, "")
	})
	r.GET("/groupOrdersByUnRented", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.GroupOrdersByUnRentedGET(rw, "")
	})

	//Chats
	r.GET("/regNewChat", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.RegNewChatGET(rw, "")
	})

	r.POST("/signupLegal", a.SignupLegalPOST) //передача данных Юридического лица 		(регистрация)
	r.POST("/signupNatur", a.SignupNaturPOST) //передача данных Физического лица		(регистрация)
	r.POST("/sigAds", a.AddAdsPOST)           //передача данных Объявления				(регистрация)
	r.POST("/login", a.LoginPOST)             //передача данных Для входа в систему		(вход)
	r.POST("/delAds", a.DeleteAdsPOST)
	r.POST("/addFavorite", a.AddFavoritePOST)
	r.POST("/delFavorite", a.DelFavoritePOST)
	r.POST("/addReview", a.AddReviewsPOST)
	r.POST("/updReview", a.UpdReviewsPOST)
	r.POST("/groupByHourlyRate", a.GroupByHourlyRatePOST)
	r.POST("/groupByDailyRate", a.GroupByDailyRatePOST)
	r.POST("/groupByCategory", a.GroupByCategoryPOST)
	r.POST("/groupByLocat", a.GroupByLocatPOST)
	r.POST("/sortFavByRecent", a.SortFavByRecentPOST)
	r.POST("/sortFavByCheaper", a.SortFavByCheaperPOST)
	r.POST("/sortFavByDearly", a.SortFavByDearlyPOST)
	r.POST("/sortReviewNewOnesFirst", a.SortReviewNewOnesFirstPOST)
	r.POST("/sortReviewOldOnesFirst", a.SortReviewOldOnesFirstPOST)
	r.POST("/sortReviewLowRatOnesFirst", a.SortReviewLowRatOnesFirstPOST)
	r.POST("/sortReviewHigRatOnesFirst", a.SortReviewHigRatOnesFirstPOST)
	r.POST("/groupAdsByRented", a.GroupAdsByRentedPOST)
	r.POST("/groupAdsByArchived", a.GroupAdsByArchivedPOST)
	r.POST("/registerRented", a.RegisterRentedPOST)
	r.POST("/rebookRented", a.RebookRentedPOST)
	r.POST("/groupOrdersByRented", a.GroupOrdersByRentedPOST)
	r.POST("/groupOrdersByUnRented", a.GroupOrdersByUnRentedPOST)

	//Chats
	r.POST("/regNewChat", a.RegNewChatPOST)
}

func (a app) PageMenuNavigation(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}

func (a app) StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}

func readCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		// log.Println("Trying to read an empty cookie name")
		return value, errors.New("you are trying to read empty cookie")
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		// log.Printf("Cookie %s not found: %v\n", name, err)
		return value, err
	}
	str := cookie.Value
	value, _ = url.QueryUnescape(str)
	// log.Printf("Cookie %s found with value: %s\n", name, value)
	return value, err
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *app {
	return &app{ctx, repository.NewRepository(dbpool), make(map[string]repository.User)}
}
