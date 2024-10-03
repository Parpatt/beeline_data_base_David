package services

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"myproject/internal"
	"myproject/internal/database"
	"myproject/internal/jwt"
	"myproject/internal/models"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Создаем глобальный контекст и Redis-клиент
var ctx = context.Background()

var deleteMe = true
var deleteMeToo = 184

type MyApp struct {
	app internal.App
}

type LegalUser struct {
	Avatar string `json:"avatar"`

	Password_hash string `json:"password_hash"`
	Email         string `json:"email" db:"email"`
	Phone_number  string `json:"phone_number"`

	Ind_num_taxp    int    `json:"ind_num_taxp"`
	Name_of_company string `json:"name_of_company"`
	Address_name    string `json:"address_name"`

	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	Data     string `json:"data"`
}

type NaturUser struct {
	Avatar string `json:"avatar"`

	Password_hash string `json:"password_hash"`
	Email         string `json:"email" db:"email"`
	Phone_number  string `json:"phone_number"`

	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`

	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	Data     string `json:"data"`
}

type Email_code struct {
	Email_code int `json:"email_code"`
}

func errorr(err error) {
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
}

func NewRepository(pool *pgxpool.Pool) *internal.Repository {
	return &internal.Repository{Pool: pool}
}

func NewApp(Ctx context.Context, dbpool *pgxpool.Pool) *MyApp {
	return &MyApp{internal.App{Ctx: Ctx, Repo: NewRepository(dbpool), Cache: make(map[string]models.User)}}
}

func UploadFilesMass(rw http.ResponseWriter, images map[string]string, pwd string) (error, bool) {
	for imageName, base64Data := range images {
		// Декодируем данные base64
		data, err := base64.StdEncoding.DecodeString(base64Data)
		errorr(err)

		// Создаем файл на сервере для сохранения декодированного файла
		dst, err := os.Create(pwd + imageName)
		errorr(err)

		defer dst.Close()

		// Записываем данные в файл
		if _, err := dst.Write(data); err != nil {
			http.Error(rw, fmt.Sprintf("Error writing to file %s: %v", imageName, err), http.StatusInternalServerError)
			return err, false
		}

		dst.Close() // Закрываем файл
	}

	return nil, true
}

func UploadFiles(rw http.ResponseWriter, image string, pwd string) (error, bool) {
	// Декодируем данные base64
	data, err := base64.StdEncoding.DecodeString(image)
	errorr(err)

	// Создаем файл на сервере для сохранения декодированного файла
	dst, err := os.Create(pwd + image)
	errorr(err)

	defer dst.Close()

	// Записываем данные в файл
	if _, err := dst.Write(data); err != nil {
		http.Error(rw, fmt.Sprintf("Error writing to file %s: %v", image, err), http.StatusInternalServerError)
		return err, false
	}

	dst.Close() // Закрываем файл

	return nil, true
}

func DeleteFile(rw http.ResponseWriter, pwd, imageName string) (error, bool) {
	// Формируем полный путь к файлу
	filePath := pwd + imageName

	// Удаляем файл
	err := os.Remove(filePath)
	errorr(err)

	return nil, true
}

func (a *MyApp) SignupUserByEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) error {
	type Kesh struct {
		Email string `json:"Email"`
		Code  int    `json:"Code"`
	}

	var kesh Kesh

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&kesh)
	errorr(err)

	// Настройки SMTP-сервера
	smtpHost := "smtp.mail.ru"
	smtpPort := "587"

	// Данные отправителя (ваша почта и пароль приложения)
	senderEmail := "parpatt_test@mail.ru"
	password := "X0h72ndPXchhjWZ4vbyT" // Пароль приложения

	// Получатель
	recipientEmail := kesh.Email

	// Сообщение
	subject := "Subject: Тебя беспокоит служба безопасности сбербанка.\n"
	body := "Введи этот код.\n"
	codeNum := 777 // Здесь лучше использовать случайный код
	message := []byte(subject + "\n" + body + strconv.Itoa(codeNum))

	// Авторизация для отправки email
	auth := smtp.PlainAuth("", senderEmail, password, smtpHost)

	// Устанавливаем обычное нешифрованное соединение
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		log.Fatal(err)
	}

	// Используем команду STARTTLS для начала TLS-сессии
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Это нужно убрать в продакшене
		ServerName:         smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Fatal(err)
	}

	// Старт авторизации
	if err = client.Auth(auth); err != nil {
		log.Fatal(err)
	}

	// Установка адреса отправителя
	if err = client.Mail(senderEmail); err != nil {
		log.Fatal(err)
	}

	// Установка адреса получателя
	if err = client.Rcpt(recipientEmail); err != nil {
		log.Fatal(err)
	}

	// Отправка сообщения
	w, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = w.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Завершение сеанса
	client.Quit()

	// Преобразуем структуру kesh в JSON
	keshData, err := json.Marshal(Kesh{Email: kesh.Email, Code: codeNum})
	if err != nil {
		log.Fatal("Ошибка при сериализации структуры kesh:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Сохраняем код подтверждения в Redis с TTL 10 минут
	err = redisClient.Set(ctx, "jwt", keshData, 10*time.Minute).Err()
	if err != nil {
		log.Fatal("Ошибка при сохранении кода в Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	rw.WriteHeader(http.StatusOK)
	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Объявление не найдено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}
	response := Response{
		Status:  "success",
		Data:    codeNum,
		Message: "Объявление показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (a *MyApp) EnterCodeFromEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) error {
	type Kesh struct {
		Email string `json:"Email"`
		Code  int    `json:"Code"`
	}

	var email Email_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return err
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Fatal("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	type Response struct {
		Status  string `json:"status"`
		Data    string `json:"data,omitempty"`
		Message string `json:"message"`
	}

	// Проверяем код
	if email.Email_code == storedKesh.Code {
		response := Response{
			Status:  "success",
			Data:    storedKesh.Email,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "fatal",
		Message: "Объявление не найдено",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (a *MyApp) SignupLegalPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) error {
	type Kesh struct {
		Email string
		Code  int
	}

	var user LegalUser

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return err
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Fatal("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	var pwd = "/home/beeline_project/media/user/"
	err, image_flag := UploadFiles(rw, user.Avatar, pwd)
	errorr(err)

	if image_flag {
		err = repo.AddNewLegalUserSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Ind_num_taxp,
			user.Name_of_company,
			user.Address_name,
			storedKesh.Email,
			user.Phone_number,
			user.Password_hash,

			user.Filename,
			user.Filetype,
			user.Data,
			pwd,
			user.Avatar,
		)
	}
	return err
}

func (a *MyApp) SignupNaturPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) error {
	type Kesh struct {
		Email string
		Code  int
	}

	var user NaturUser

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return err
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Fatal("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	var pwd = "/home/beeline_project/media/user/"
	err, image_flag := UploadFiles(rw, user.Avatar, pwd)
	errorr(err)

	if image_flag {
		err = repo.AddNewNaturUserSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Name,
			user.Surname,
			user.Patronymic,
			storedKesh.Email,
			user.Phone_number,
			user.Password_hash,

			user.Filename,
			user.Filetype,
			user.Data,

			pwd,
			user.Avatar,
		)
	}
	return err
}

func (a *MyApp) EditingLegalUserDataPOST(rw http.ResponseWriter, r *http.Request) {
	var legalUser LegalUser

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&legalUser)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.EditingLegalUserDataSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, legalUser.Ind_num_taxp, legalUser.Name_of_company, legalUser.Address_name, "/home/beeline_project/media/user/", legalUser.Avatar)

		errorr(err)
	}
}

func (a *MyApp) EditingNaturUserDataPOST(rw http.ResponseWriter, r *http.Request) {
	var naturUser NaturUser

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&naturUser)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.EditingNaturUserDataSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, naturUser.Surname, naturUser.Name, naturUser.Patronymic, "/home/beeline_project/media/user/", naturUser.Avatar)

		errorr(err)
	}
}

type Email_name struct {
	Email_name string `json:"email_name"`
}

func (a *MyApp) SendCodForEmail(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var email Email_name

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Настройки SMTP-сервера
	smtpHost := "smtp.mail.ru"
	smtpPort := "587"

	// Данные отправителя (ваша почта и пароль приложения)
	senderEmail := "parpatt_test@mail.ru"
	password := "X0h72ndPXchhjWZ4vbyT" // Пароль приложения

	// Получатель
	recipientEmail := email.Email_name

	// Сообщение
	subject := "Subject: Тебя беспокоит служба безопасности сбербанка.\n"
	body := "Введи этот код.\n"
	codeNum := 777 // Здесь лучше использовать случайный код
	message := []byte(subject + "\n" + body + strconv.Itoa(codeNum))

	// Авторизация для отправки email
	auth := smtp.PlainAuth("", senderEmail, password, smtpHost)

	// Устанавливаем обычное нешифрованное соединение
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		log.Fatal(err)
	}

	// Используем команду STARTTLS для начала TLS-сессии
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Это нужно убрать в продакшене
		ServerName:         smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Fatal(err)
	}

	// Старт авторизации
	if err = client.Auth(auth); err != nil {
		log.Fatal(err)
	}

	// Установка адреса отправителя
	if err = client.Mail(senderEmail); err != nil {
		log.Fatal(err)
	}

	// Установка адреса получателя
	if err = client.Rcpt(recipientEmail); err != nil {
		log.Fatal(err)
	}

	// Отправка сообщения
	w, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = w.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Завершение сеанса
	client.Quit()

	// Сохраняем код подтверждения в Redis с TTL 10 минут
	err = redisClient.Set(ctx, "jwt", 777, 10*time.Minute).Err()
	if err != nil {
		log.Fatal("Ошибка при сохранении кода в Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (a *MyApp) EnterCodFromEmail(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var email Email_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем данные из Redis
	kash_code, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	kash_code_int, err := strconv.Atoi(kash_code)
	if err != nil {
		log.Fatal("Ошибка преобразования кода:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if email.Email_code == kash_code_int {
		fmt.Println("Да")
	}
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *MyApp) LoginPOST(rw http.ResponseWriter, r *http.Request) {
	var login Login

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = repo.LoginSQL(a.app.Ctx, login.Login, login.Password, a.app.Repo.Pool, rw)

	errorr(err)
}

type ProductList struct {
	Ads_list []int `json:"Ads_list"`
}

func (a *MyApp) ProductListPOST(rw http.ResponseWriter, r *http.Request) {
	var productList ProductList

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&productList)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.ProductListSQL(a.app.Ctx, rw, productList.Ads_list, a.app.Repo.Pool)

	errorr(err)
}

type PrintAds struct {
	Ads_id int `json:"Ads_id"`
}

func (a *MyApp) PrintAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var printAds PrintAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&printAds)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.PrintAdsSQL(a.app.Ctx, rw, printAds.Ads_id, a.app.Repo.Pool)

	errorr(err)
}

type SortProductListAll struct {
	Category []int  `json:"Category"`
	LowNum   int    `json:"LowNum"`
	HigNum   int    `json:"HigNum"`
	LowDate  int64  `json:"LowDate"`
	HigDate  int64  `json:"HigDate"`
	Location string `json:"Location"`
	Rating   int    `json:"Rating"`
}

func (a *MyApp) SortProductListAllPOST(rw http.ResponseWriter, r *http.Request) {
	var sortProductList SortProductListAll

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sortProductList)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем UNIX-время в тип time.Time
	lowDate := time.Unix(sortProductList.LowDate, 0).UTC()
	higDate := time.Unix(sortProductList.HigDate, 0).UTC()

	fmt.Println(lowDate)
	fmt.Println(higDate)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListAllSQL(a.app.Ctx, rw, sortProductList.Category, sortProductList.LowNum, sortProductList.HigNum, lowDate, higDate, sortProductList.Location, sortProductList.Rating, a.app.Repo.Pool)

	errorr(err)
}

type Ads struct {
	Image []string `json:"id"`

	Id          int    `json:"Id"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Hourly_rate int    `json:"Hourly_rate"`
	Daily_rate  int    `json:"Daily_rate"`
	Category_id int    `json:"Category_id"`
	Location    string `json:"Location"`
}

func (a *MyApp) SignupAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var ads Ads

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(rw, token.Value)

	var pwd = "/home/beeline_project/media/ads/"
	var images map[string]string

	for i := 0; i < len(ads.Image); i++ {
		images[ads.Image[i][0:5]+strconv.Itoa(user_id)+time.Now().Format("2001-01-01_15:04:05")] = (pwd + images[ads.Image[i][0:5]+strconv.Itoa(user_id)+time.Now().Format("2001-01-01_15:04:05")])
	}

	err, image_flag := UploadFilesMass(rw, images, pwd)
	errorr(err)

	if token_flag && image_flag {
		err = repo.SignupAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, ads.Title, ads.Description, ads.Hourly_rate, ads.Daily_rate, user_id, ads.Category_id, ads.Location, time.Now(), images, pwd)

		errorr(err)
	}
}

func (a *MyApp) UpdAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var ads Ads

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(rw, token.Value)

	var pwd = "/home/beeline_project/media/ads/"
	var images map[string]string

	for i := 0; i < len(ads.Image); i++ {
		images[ads.Image[i][0:5]+strconv.Itoa(user_id)+time.Now().Format("2001-01-01_15:04:05")] = (pwd + images[ads.Image[i][0:5]+strconv.Itoa(user_id)+time.Now().Format("2001-01-01_15:04:05")])
	}

	err, image_flag := UploadFilesMass(rw, images, pwd)

	if token_flag && image_flag {
		err := repo.UpdAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, ads.Title, ads.Description, ads.Hourly_rate, ads.Daily_rate, user_id, ads.Category_id, ads.Location, pwd, ads.Id, time.Now(), pwd, images)

		errorr(err)
	}
}

func (a *MyApp) DelAdsPOST(rw http.ResponseWriter, r *http.Request) {
	type DelAds struct {
		Ads_id  int `json:"Ads_id"`
		User_id int `json:"User_id"`
	}

	var ads DelAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)
	if deleteMe {
		err := repo.DelAdsSQL(a.app.Ctx, ads.Ads_id, ads.User_id, rw, a.app.Repo.Pool)
		if err != nil {
			fmt.Errorf("Ошибка создания объявления: %v", err)
			return
		}
	}
	//}
}

type FavAds struct {
	User_id int `json:"User_id"`
	Ads_id  int `json:"Ads_id"`
}

func (a *MyApp) SigFavAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var ads FavAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)
	if deleteMe {
		err := repo.SigFavAdsSQL(a.app.Ctx, ads.User_id, ads.Ads_id, a.app.Repo.Pool, rw)
		if err != nil {
			fmt.Errorf("Ошибка создания объявления: %v", err)
			return
		}
	}
	//}
}

func (a *MyApp) DelFavAdsPOST(rw http.ResponseWriter, r *http.Request) {
	var ads FavAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)
	if deleteMe {
		err := repo.DelFavAdsSQL(a.app.Ctx, ads.User_id, ads.Ads_id, a.app.Repo.Pool, rw)
		if err != nil {
			fmt.Errorf("Ошибка создания объявления: %v", err)
			return
		}
	}
	//}
}

type Title struct {
	Title string `json:"Title"`
}

func (a *MyApp) SearchForTechPOST(rw http.ResponseWriter, r *http.Request) {
	var title Title

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&title)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SearchForTechSQL(a.app.Ctx, title.Title, rw, a.app.Repo.Pool)
	errorr(err)
}

type SortProductListCategoriez struct {
	Category []int `json:"Category"`
}

func (a *MyApp) SortProductListCategoriezPOST(rw http.ResponseWriter, r *http.Request) {
	var sortProductList SortProductListCategoriez

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sortProductList)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListCategoriezSQL(a.app.Ctx, rw, sortProductList.Category, a.app.Repo.Pool)

	errorr(err)
}

type SigChat struct {
	Id_user int `json:"Id_user"`
	Id_ads  int `json:"Id_ads"`
}

func (a *MyApp) SigChatPOST(rw http.ResponseWriter, r *http.Request) {
	var sigChat SigChat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sigChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.SigChatSQL(a.app.Ctx, rw, sigChat.Id_user, sigChat.Id_ads, a.app.Repo.Pool)

	errorr(err)
}

type Chat struct {
	Id_chat int `json:"Id_chat"`
}

func (a *MyApp) OpenChatPOST(rw http.ResponseWriter, r *http.Request) {
	var openChat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&openChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.OpenChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, user_id)
	}

	errorr(err)
}

type SendMess struct {
	Id_chat int    `json:"Id_chat"`
	Text    string `json:"Text"`
	Image   string `json:"Image"`
}

func (a *MyApp) SendMessagePOST(rw http.ResponseWriter, r *http.Request) {
	var sendMess SendMess

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	flag, user_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.SendMessageSQL(a.app.Ctx, rw, a.app.Repo.Pool, sendMess.Id_chat, user_id, sendMess.Text, sendMess.Image)

		errorr(err)
	}
}

func (a *MyApp) SigDisputInChatPOST(rw http.ResponseWriter, r *http.Request) {
	var chat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&chat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.SigDisputInChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, chat.Id_chat, 128)

	errorr(err)
}

type SigReview struct {
	Ads_id  int    `json:"Ads_id"`
	Rating  int    `json:"Rating"`
	Comment string `json:"Comment"`
}

func (a *MyApp) SigReviewPOST(rw http.ResponseWriter, r *http.Request) {
	var sigReview SigReview

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sigReview)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.SigReviewSQL(a.app.Ctx, rw, sigReview.Ads_id, 27, sigReview.Rating, sigReview.Comment, a.app.Repo.Pool)

	errorr(err)
}

type UpdReview struct {
	Review_id int    `json:"Review_id"`
	Rating    int    `json:"Rating"`
	Comment   string `json:"Comment"`
}

func (a *MyApp) UpdReviewPOST(rw http.ResponseWriter, r *http.Request) {
	var updReview UpdReview

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&updReview)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.UpdReviewSQL(a.app.Ctx, rw, 128, updReview.Review_id, updReview.Rating, updReview.Comment, a.app.Repo.Pool)

	errorr(err)
}

func (a *MyApp) MediatorStartWorkingPOST(rw http.ResponseWriter, r *http.Request) {
	var openChat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&openChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	errorr(err)

	flag, mediator_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.MediatorStartWorkingSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, mediator_id)
	}
	errorr(err)
}

func (a *MyApp) MediatorEnterInChatPOST(rw http.ResponseWriter, r *http.Request) {
	var openChat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&openChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	errorr(err)

	flag, mediator_id := jwt.IsAuthorized(rw, token)

	if flag {
		err = repo.MediatorEnterInChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, mediator_id)
	}

	errorr(err)
}

func (a *MyApp) MediatorFinishJobInChatPOST(rw http.ResponseWriter, r *http.Request) {
	var openChat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&openChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.MediatorFinishJobInChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat)

	errorr(err)
}

type Transact struct {
	User_id int `json:"User_id"`
	User_2  int `json:"User_2"`
	Amount  int `json:"Amount"`
}

func (a *MyApp) TransactionToAnotherPOST(rw http.ResponseWriter, r *http.Request) {
	var transact Transact

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&transact)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.TransactionToAnotherSQL(a.app.Ctx, rw, a.app.Repo.Pool, transact.User_id, transact.User_2, transact.Amount)

	errorr(err)
}

func (a *MyApp) TransactionToSomethingPOST(rw http.ResponseWriter, r *http.Request) {
	var transact Transact

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&transact)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.TransactionToSomethingSQL(a.app.Ctx, rw, a.app.Repo.Pool, transact.User_id, transact.Amount)

	errorr(err)
}

func (a *MyApp) TransactionToReturnAmountPOST(rw http.ResponseWriter, r *http.Request) {
	var transact Transact

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&transact)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.TransactionToReturnAmountSQL(a.app.Ctx, rw, a.app.Repo.Pool, transact.User_id, transact.Amount)

	errorr(err)
}

type Order struct {
	Ad_id       int   `json:"Ad_id"`
	User_id     int   `json:"User_id"`
	Total_price int   `json:"Total_price"`
	Starts_at   int64 `json:"Starts_at"`
	Ends_at     int64 `json:"Ends_at"`
}

func (a *MyApp) RegisterOrderPOST(rw http.ResponseWriter, r *http.Request) {
	var order Order

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем UNIX-время в тип time.Time
	Starts_at := time.Unix(order.Starts_at, 0).UTC()
	Ends_at := time.Unix(order.Ends_at, 0).UTC()

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.RegisterOrderSQL(a.app.Ctx, rw, a.app.Repo.Pool, order.Ad_id, order.User_id, order.Total_price, Starts_at, Ends_at)

	errorr(err)
}

func (a *MyApp) RegBookingPOST(rw http.ResponseWriter, r *http.Request) {
	type Booking struct {
		Order_id  int   `json:"Order_id"`
		Starts_at int64 `json:"Starts_at"`
		Ends_at   int64 `json:"Ends_at"`
		Amount    int   `json:"Amount"`
	}

	var booking Booking

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем UNIX-время в тип time.Time
	Starts_at := time.Unix(booking.Starts_at, 0).UTC()
	Ends_at := time.Unix(booking.Ends_at, 0).UTC()

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.RegBookingSQL(a.app.Ctx, rw, a.app.Repo.Pool, 27, booking.Order_id, Starts_at, Ends_at, booking.Amount)

	errorr(err)
}

func (a *MyApp) RebookBookingPOST(rw http.ResponseWriter, r *http.Request) {
	type Booking struct {
		Id_old_book int   `json:"Id_old_book"`
		User_id     int   `json:"User_id"`
		Starts_at   int64 `json:"Starts_at"`
		Ends_at     int64 `json:"Ends_at"`
		Amount      int   `json:"Amount"`
	}

	var booking Booking

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Преобразуем UNIX-время в тип time.Time
	Starts_at := time.Unix(booking.Starts_at, 0).UTC()
	Ends_at := time.Unix(booking.Ends_at, 0).UTC()

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.RebookBookingSQL(a.app.Ctx, rw, a.app.Repo.Pool, booking.Id_old_book, booking.User_id, Starts_at, Ends_at, booking.Amount)

	errorr(err)
}

type Report struct {
	Order_id int `json:"Order_id"`
}

func (a *MyApp) RegReportPOST(rw http.ResponseWriter, r *http.Request) {
	var report Report

	//запрос к счёту

	//если всё ок, то продолжаем

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	err = repo.RegReportSQL(a.app.Ctx, rw, a.app.Repo.Pool, report.Order_id)

	errorr(err)
}

type Passwd struct {
	Passwd_1 string `json:"Passwd_1"`
	Passwd_2 string `json:"Passwd_2"`
}

func (a *MyApp) SendCodeForRecoveryPassWithEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var passwd Passwd

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&passwd)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	if passwd.Passwd_1 == passwd.Passwd_2 {
		err = repo.RecoveryPassWithEmailSQL(a.app.Ctx, rw, a.app.Repo.Pool, redisClient, 184, passwd.Passwd_1)

		errorr(err)
	} else {
		type Response struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}

		response := Response{
			Status:  "fatal",
			Message: "Поля не совпадают!",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	errorr(err)
}

func (a *MyApp) EnterCodeForRecoveryPassWithEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var email Email_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	type Kesh struct {
		Passwd string
		Code   int
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Fatal("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Проверяем код
	if email.Email_code == storedKesh.Code {
		err = repo.EnterCodeForRecoveryPassWithEmailSQL(a.app.Ctx, rw, a.app.Repo.Pool, 184, storedKesh.Passwd)
		if err != nil {
			log.Fatal(err)
		}
		rw.Write([]byte("Смена завершена успешно"))
	} else {
		http.Error(rw, "Неверный код подтверждения", http.StatusUnauthorized)
	}
	errorr(err)
}

func (a *MyApp) AutorizLoginEmailSendPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var login Login

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&login)
	errorr(err)

	// Настройки SMTP-сервера
	smtpHost := "smtp.mail.ru"
	smtpPort := "587"

	// Данные отправителя (ваша почта и пароль приложения)
	senderEmail := "parpatt_test@mail.ru"
	password := "X0h72ndPXchhjWZ4vbyT" // Пароль приложения

	// Получатель
	recipientEmail := login.Login

	// Сообщение
	subject := "Subject: Тебя беспокоит служба безопасности сбербанка.\n"
	body := "Введи этот код.\n"
	codeNum := 777 // Здесь лучше использовать случайный код
	message := []byte(subject + "\n" + body + strconv.Itoa(codeNum))

	// Авторизация для отправки email
	auth := smtp.PlainAuth("", senderEmail, password, smtpHost)

	// Устанавливаем обычное нешифрованное соединение
	client, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		log.Fatal(err)
	}

	// Используем команду STARTTLS для начала TLS-сессии
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Это нужно убрать в продакшене
		ServerName:         smtpHost,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		log.Fatal(err)
	}

	// Старт авторизации
	if err = client.Auth(auth); err != nil {
		log.Fatal(err)
	}

	// Установка адреса отправителя
	if err = client.Mail(senderEmail); err != nil {
		log.Fatal(err)
	}

	// Установка адреса получателя
	if err = client.Rcpt(recipientEmail); err != nil {
		log.Fatal(err)
	}

	// Отправка сообщения
	w, err := client.Data()
	if err != nil {
		log.Fatal(err)
	}

	_, err = w.Write(message)
	if err != nil {
		log.Fatal(err)
	}

	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Завершение сеанса
	client.Quit()

	type Kesh struct {
		Login Login
		Code  int
	}

	// Преобразуем структуру kesh в JSON
	keshData, err := json.Marshal(Kesh{Login: login, Code: codeNum})
	if err != nil {
		log.Fatal("Ошибка при сериализации структуры kesh:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Сохраняем код подтверждения в Redis с TTL 10 минут
	err = redisClient.Set(ctx, "jwt", keshData, 10*time.Minute).Err()
	if err != nil {
		log.Fatal("Ошибка при сохранении кода в Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Код отправлен на вашу почту."))
}

func (a *MyApp) AutorizLoginEmailEnterPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	var code Email_code

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&code)
	errorr(err)

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, "jwt").Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	type Kesh struct {
		Login Login
		Code  int
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Fatal("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if storedKesh.Code == code.Email_code {
		err := repo.LoginSQL(a.app.Ctx, storedKesh.Login.Login, storedKesh.Login.Password, a.app.Repo.Pool, rw)
		errorr(err)
	}
}
