package application

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func (a app) authorized(next httprouter.Handle) httprouter.Handle {
	//функция authorized. Возвращает функцию, которая перенаправляет на страницу /login в случае ошибки, или в случае
	//нарушения целостности токена, либо, выдаёт сообщение о успешной авторизации

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

		// log.Println("Authorization successful")
		next(rw, r, ps)
	}
}

func (a app) LoginGET(rw http.ResponseWriter, message string) { //страница входа в систему
	lp := filepath.Join("public", "html", "login.html") //lp содержит путь "public/html/login.html"

	tmpl, err := template.ParseFiles(lp) //передаём код из файла lp в переменную tmpl
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	type answer struct {
		Message string
	}
	data := answer{message}

	err = tmpl.ExecuteTemplate(rw, "login", data) //проверка на наличии ошибок в коде файла lp
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a app) PageMenuGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "html", "menu.html")

	tmpl, err := template.ParseFiles(sp)
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

func (a app) Logout(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, v := range r.Cookies() {
		c := http.Cookie{
			Name:   v.Name,
			MaxAge: -1}
		http.SetCookie(rw, &c)
	}
	http.Redirect(rw, r, "/login", http.StatusSeeOther)
}
