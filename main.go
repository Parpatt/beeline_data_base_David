package main

import (
	"context"
	"fmt"
	"log"
	"myproject/internal/app"
	"myproject/internal/database"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

const apiUrl = "https://api.t-bank.com/v1/nominal-accounts" // Инициализация клиента Redis
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "185.112.83.36:6379", // Проверьте, что адрес и порт верные
		Password: "",                   // Если Redis настроен без пароля
		DB:       0,                    // Используйте базу данных 0
	})
}

var redisClient = NewRedisClient() // Глобальный клиент Redis
func main() {
	ctx := context.Background()
	// Инициализация подключения к базе данных
	dbpool, err := database.InitDBConn(ctx)
	if err != nil {
		log.Fatalf("%v failed to init DB connection", err)
	}
	defer dbpool.Close()

	logFile := &lumberjack.Logger{
		Filename:   "/root/home/beeline_project/beeline.log", //Имя лог-файла
		MaxSize:    10,                                       // Максимальный размер файла в МБ
		MaxBackups: 5,                                        // Максимальное количество копий
		MaxAge:     30,                                       // Хранить логи 30 дней
		Compress:   true,
	}

	logger := zerolog.New(logFile).With().Timestamp().Logger()
	logger.Info().Msg("Start server")

	// Инициализация приложения
	a := app.NewApp(ctx, dbpool)
	r := httprouter.New()
	// Определяем маршруты для приложения
	a.Routes(r, ctx, dbpool, redisClient, logger)
	// Применяем CORS middleware ко всем маршрутам
	handlerWithCORS := corsMiddleware(r)
	// Настройка сервера
	srv := &http.Server{Addr: "185.112.83.36:8090",
		Handler: handlerWithCORS, // Используем обработчик с поддержкой CORS
	}
	// Запуск сервера
	fmt.Println("Сервер запущен на http://185.112.83.36:8090")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %s\n", err)
	}
}

// Middleware для добавления CORS-заголовков
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Устанавливаем CORS-заголовки
		// w.Header().Set("Access-Control-Allow-Origin", origin)
		// w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Устанавливаем CORS-заголовки авторская версия от Давида !-)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// Если это preflight-запрос OPTIONS, возвращаем 200 OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}
