package application

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/jwt"
)

func (a app) LoginPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	login := r.FormValue("email")
	password := r.FormValue("password")

	if login == "" || password == "" {
		a.LoginGET(rw, "Необходимо указать логин и пароль!")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	user, err := a.repo.LoginSQL(a.ctx, login, hashedPass)
	if err != nil {
		a.LoginGET(rw, "Вы ввели неверный логин или пароль!")
		return
	}

	//логин и пароль совпадают, поэтому генерируем токен, пишем его в кеш и в куки
	validToken, err := jwt.GenerateJWT("имя") //получаем токен в строковом типе
	if err != nil {
		fmt.Fprintf(rw, err.Error()) //выводим ошибку, при её наличии,
	}

	// fmt.Fprintf(rw, validToken)	// токен

	a.cache[validToken] = user

	livingTime := 60 * time.Minute
	expiration := time.Now().Add(livingTime)

	//кука будет жить 1 час
	cookie := http.Cookie{Name: "token", Value: url.QueryEscape(validToken), Expires: expiration}
	http.SetCookie(rw, &cookie)
	http.Redirect(rw, r, "/pageMenu", http.StatusSeeOther)
}

func (a app) SignupNaturPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := strings.TrimSpace(r.FormValue("name"))
	surname := strings.TrimSpace(r.FormValue("surname"))
	patronymic := strings.TrimSpace(r.FormValue("patronymic"))
	email := strings.TrimSpace(r.FormValue("email"))
	phoneNum := strings.TrimSpace(r.FormValue("phoneNum"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))

	if name == "" || surname == "" || patronymic == "" || email == "" || phoneNum == "" || password == "" {
		a.SignupNaturGET(rw, "Все поля должны быть заполнены!")
		return
	}

	if password != password2 {
		a.SignupNaturGET(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	err := a.repo.AddNewNaturUserSQL(a.ctx, name, surname, patronymic, email, phoneNum, hashedPass)
	if err != nil {
		a.SignupNaturGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	a.LoginGET(rw, fmt.Sprintf("%s, вы успешно зарегистрированы! Теперь вам доступен вход через страницу авторизации.", name))
}

func (a app) SignupLegalPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ind_num_taxp := strings.TrimSpace(r.FormValue("ind_num_taxp"))
	name_of_company := strings.TrimSpace(r.FormValue("name_of_company"))
	address_name := strings.TrimSpace(r.FormValue("address_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	phoneNum := strings.TrimSpace(r.FormValue("phoneNum"))
	password := strings.TrimSpace(r.FormValue("password"))
	password2 := strings.TrimSpace(r.FormValue("password2"))

	if ind_num_taxp == "" || name_of_company == "" || address_name == "" || email == "" || phoneNum == "" || password == "" {
		a.SignupNaturGET(rw, "Все поля должны быть заполнены!")
		return
	}

	if password != password2 {
		a.SignupNaturGET(rw, "Пароли не совпадают! Попробуйте еще")
		return
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	err := a.repo.AddNewLegalUserSQL(a.ctx, ind_num_taxp, name_of_company, address_name, email, phoneNum, hashedPass)
	if err != nil {
		a.SignupNaturGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
		return
	}

	a.LoginGET(rw, fmt.Sprintf("%s, вы успешно зарегистрированы! Теперь вам доступен вход через страницу авторизации.", name_of_company))
}
