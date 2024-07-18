package application

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/julienschmidt/httprouter"

	"myproject/internal/jwt"
	"myproject/internal/repository"
)

type app struct {
	ctx   context.Context
	repo  *repository.Repository
	cache map[string]repository.User
}

func (a app) Routes(r *httprouter.Router) {
	r.ServeFiles("/public/*filepath", http.Dir("public"))
	r.GET("/", a.authorized(a.StartPage))
	r.GET("/login", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.LoginPage(rw, "")
	})
	r.POST("/login", a.Login)  //Лист чтобы залогиниться
	r.GET("/logout", a.Logout) //
	r.GET("/signupNatur", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupNaturPage(rw, "")
	})
	r.GET("/signupLegal", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupLegalPage(rw, "")
	})
	r.GET("/signupAds", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.SignupAdsPage(rw, "")
	})
	r.POST("/signupLegal", a.SignupLegal) //создание Юридического лица
	r.POST("/signupNatur", a.SignupNatur) //создание Физического лица
	r.POST("/signupAds", a.Add_ads)       //создание Объявления
}

func (a app) authorized(next httprouter.Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token, err := readCookie("token", r)
		if err != nil {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		if _, ok := a.cache[token]; !ok {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		log.Println("Authorization successful")
		next(rw, r, ps)
	}
}

func (a app) LoginPage(rw http.ResponseWriter, message string) {
	lp := filepath.Join("public", "html", "login.html")

	tmpl, err := template.ParseFiles(lp)
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

func (a app) Login(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := r.FormValue("email")
	password := r.FormValue("password")

	if login == "" || password == "" {
		a.LoginPage(rw, "Необходимо указать логин и пароль!")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	user, err := a.repo.Login(a.ctx, login, hashedPass)
	if err != nil {
		a.LoginPage(rw, "Вы ввели неверный логин или пароль!")
		return
	}

	//логин и пароль совпадают, поэтому генерируем токен, пишем его в кеш и в куки
	validToken, err := jwt.GenerateJWT("имя") //получаем токен в строковом типе
	if err != nil {
		fmt.Fprintf(rw, err.Error()) //выводим ошибку, при её наличии,
	}
	fmt.Fprintf(rw, validToken)

	a.cache[validToken] = user

	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)

	//кука будет жить 1 час
	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(validToken), Expires: expiration}
	http.SetCookie(rw, &cookie)
	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func (a app) Logout(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1}
		http.SetCookie(rw, &c)
	}
	http.Redirect(rw, r, "/login", http.StatusSeeOther)
}

func (a app) SignupNaturPage(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "html", "signupNatur.html")

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

func (a app) SignupLegalPage(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "html", "signupLegal.html")

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

func (a app) SignupAdsPage(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "html", "signupAds.html")

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

func (a app) SignupNatur(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := strings.TrimSpace(r.FormValue("name"))
	surname := strings.TrimSpace(r.FormValue("surname"))
	patronymic := strings.TrimSpace(r.FormValue("patronymic"))
	email := strings.TrimSpace(r.FormValue("email"))
	phoneNum := strings.TrimSpace(r.FormValue("phoneNum"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))

	if name == "" || surname == "" || patronymic == "" || email == "" || phoneNum == "" || password == "" {
		a.SignupNaturPage(rw, "Все поля должны быть заполнены!")
		return
	}

	if password != password2 {
		a.SignupNaturPage(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	err := a.repo.AddNewNaturUser(a.ctx, name, surname, patronymic, email, phoneNum, hashedPass)
	if err != nil {
		a.SignupNaturPage(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	a.LoginPage(rw, fmt.Sprintf("%s, вы успешно зарегистрированы! Теперь вам доступен вход через страницу авторизации.", name))
}

func (a app) SignupLegal(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ind_num_taxp := strings.TrimSpace(r.FormValue("ind_num_taxp"))
	name_of_company := strings.TrimSpace(r.FormValue("name_of_company"))
	address_name := strings.TrimSpace(r.FormValue("address_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	phoneNum := strings.TrimSpace(r.FormValue("phoneNum"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))

	if ind_num_taxp == "" || name_of_company == "" || address_name == "" || email == "" || phoneNum == "" || password == "" {
		a.SignupLegalPage(rw, "Все поля должны быть заполнены!")
		return
	}

	if password != password2 {
		a.SignupLegalPage(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	err := a.repo.AddNewLegalUser(a.ctx, ind_num_taxp, name_of_company, address_name, email, phoneNum, hashedPass)
	if err != nil {
		a.SignupLegalPage(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	a.LoginPage(rw, fmt.Sprintf("%s, вы успешно зарегистрированы! Теперь вам доступен вход через страницу авторизации.", name_of_company))
}

func (a app) Add_ads(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	hourly_rate := strings.TrimSpace(r.FormValue("hourly_rate"))
	daily_rate := strings.TrimSpace(r.FormValue("daily_rate"))
	owner_id := 29
	category_id := strings.TrimSpace(r.FormValue("category_id"))
	location := strings.TrimSpace(r.FormValue("location"))
	created_at := time.Now()
	updated_at := time.Now()
	status := 1

	if title == "" || description == "" ||
		hourly_rate == "" || daily_rate == "" ||
		owner_id == 0 || category_id == "" ||
		location == "" || status == 0 {
		a.SignupAdsPage(rw, "Все поля должны быть заполнены!")
		return
	}

	int_hourly_rate, _ := strconv.Atoi(hourly_rate)
	int_daily_rate, _ := strconv.Atoi(daily_rate)
	int_category_id, _ := strconv.Atoi(category_id)

	err := a.repo.AddNewAds(a.ctx, title, description, int_hourly_rate, int_daily_rate, owner_id, int_category_id, location, created_at, updated_at, status)
	if err != nil {
		a.SignupAdsPage(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	token, err := readCookie("token", r)
	a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	a.LoginPage(rw, fmt.Sprintf(err.Error()))
}

func (a app) StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(rw, "")
}

func readCookie(name string, r *http.Request) (value string, err error) {
	if name == "" {
		log.Println("Trying to read an empty cookie name")
		return value, errors.New("you are trying to read empty cookie")
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		log.Printf("Cookie %s not found: %v\n", name, err)
		return value, err
	}
	str := cookie.Value
	value, _ = url.QueryUnescape(str)
	log.Printf("Cookie %s found with value: %s\n", name, value)
	return value, err
}

func NewApp(ctx context.Context, dbpool *pgxpool.Pool) *app {
	return &app{ctx, repository.NewRepository(dbpool), make(map[string]repository.User)}
}
