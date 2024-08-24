package internal

import (
	"context"
	"errors"
	"myproject/internal/models"
	"net/http"
	"net/url"

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

func ReadCookie(name string, r *http.Request) (value string, err error) {
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
