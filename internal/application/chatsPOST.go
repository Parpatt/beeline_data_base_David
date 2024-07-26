package application

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (a app) RegNewChatPOST(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id_user, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("id_user")))
	id_buddy, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("id_buddy")))

	if id_user != id_buddy {
		err := a.repo.RegNewChatSQL(a.ctx, rw, id_user, id_buddy)
		if err != nil {
			a.SignupAdsGET(rw, fmt.Sprintf("Ошибка создания пользователя: %v", err))
			return
		}
	}

	// token, err := readCookie("token", r)
	// a.LoginPage(rw, fmt.Sprintf(a.cache[token].Name))
	// a.LoginPage(rw, fmt.Sprintf(err.Error()))
}
