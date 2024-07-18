package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/application"
	"myproject/internal/repository"
)

func main() {
	ctx := context.Background()

	dbpool, err := repository.InitDBConn(ctx)
	if err != nil {
	}
	defer dbpool.Close()

	a := application.NewApp(ctx, dbpool)
	r := httprouter.New()
	a.Routes(r)

	srv := &http.Server{Addr: "0.0.0.0:8080", Handler: r}
	fmt.Println("It is alive! Try http://localhost:8080")
	srv.ListenAndServe()
}
