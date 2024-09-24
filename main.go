package main

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"

	"myproject/internal/app"
	"myproject/internal/database"
)

const apiUrl = "https://api.t-bank.com/v1/nominal-accounts"

type Account struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}

// func getAccount(accountID string, apiKey string) (*Account, error) {
// client := &http.Client{}
// req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", apiUrl, accountID), nil)
// if err != nil {
// return nil, err
// }

// req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

// resp, err := client.Do(req)
// if err != nil {
// return nil, err
// }
// defer resp.Body.Close()

// var account Account
// if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
// return nil, err
// }

// return &account, nil
// }

// Инициализация клиента Redis
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "176.124.192.39:6379", // Адрес Redis сервера
		Password: "",               // Пароль Redis, если не используется, оставляем пустым
		DB:       0,                // Используем базу данных 0
	})
}

var redisClient = NewRedisClient() // Глобальный клиент Redis

func main() {
	ctx := context.Background()

	dbpool, err := database.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("%w failed to init DB connection", err)
	}
	defer dbpool.Close()

	rdb := NewRedisClient()

	a := app.NewApp(ctx, dbpool)
	r := httprouter.New()
	a.Routes(r, ctx, dbpool, rdb)

	srv := &http.Server{Addr: "176.124.192.39:8080", Handler: r}
	fmt.Println("It is alive! Try http://176.124.192.39:8080")
	// fmt.Printf("Account ID: %s, Balance: %.2f\n", account.ID, account.Balance)
	srv.ListenAndServe()
}
