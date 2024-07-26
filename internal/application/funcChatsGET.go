package application

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (a app) RegNewChatGET(rw http.ResponseWriter, message string) {
	sp := filepath.Join("public", "chats", "regNewChat.html")

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
