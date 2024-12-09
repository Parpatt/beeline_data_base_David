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
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	jwt_from_prod_list "github.com/golang-jwt/jwt"
)

// Создаем глобальный контекст и Redis-клиент
var ctx = context.Background()

// Конфигурация для OAuth
const (
	VKClientID     = "ваш_client_id"
	VKClientSecret = "ваш_client_secret"
	VKRedirectURI  = "http://localhost:8080/vk/callback"
	VKAuthURL      = "https://oauth.vk.com/authorize"
	VKTokenURL     = "https://oauth.vk.com/access_token"
	VKUserInfoURL  = "https://api.vk.com/method/users.get"
)

type MyApp struct {
	app internal.App
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Owner_id struct {
	Owner_id int `json:"owner_id"`
}

type Ads_id struct {
	Ads_id int `json:"ads_id"`
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

	Data string `json:"data"`
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

type Reg_code struct {
	Reg_code int `json:"reg_code"`
}

func errorr(err error) {
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
}

func validatePassword(rw http.ResponseWriter, password string) bool {
	if len(password) < 8 {
		response := Response{
			Status:  "falat",
			Message: "длина пароля должна быть не менее 8 символов",
		}

		json.NewEncoder(rw).Encode(response)

		return false
	}

	// Проверка наличия хотя бы одной цифры
	if matched, _ := regexp.MatchString(`[0-9]`, password); !matched {
		response := Response{
			Status:  "falat",
			Message: "пароль должен содержать хотя бы одну цифру",
		}

		json.NewEncoder(rw).Encode(response)

		return false
	}

	// Проверка наличия хотя бы одной строчной буквы
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		response := Response{
			Status:  "falat",
			Message: "пароль должен содержать хотя бы одну строчную букву",
		}

		json.NewEncoder(rw).Encode(response)

		return false
	}

	// Проверка наличия хотя бы одной заглавной буквы
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		response := Response{
			Status:  "falat",
			Message: "пароль должен содержать хотя бы одну заглавную букву",
		}

		json.NewEncoder(rw).Encode(response)

		return false
	}

	// Проверка наличия хотя бы одного специального символа
	// if matched, _ := regexp.MatchString(`*[!@#~$%^&*()_+{}":;'?/>.<,]`, password); !matched {
	// 	response := Response{
	// 		Status:  "falat",
	// 		Message: "пароль должен содержать хотя бы один специальный символ",
	// 	}

	// 	json.NewEncoder(rw).Encode(response)

	// 	return false
	// }

	return true
}

func NewRepository(pool *pgxpool.Pool) *internal.Repository {
	return &internal.Repository{Pool: pool}
}

func NewApp(Ctx context.Context, dbpool *pgxpool.Pool) *MyApp {
	return &MyApp{internal.App{Ctx: Ctx, Repo: NewRepository(dbpool), Cache: make(map[string]models.User)}}
}

func UploadImagesMass(rw http.ResponseWriter, images []string, pwd string, user_id string) (error, bool, []string) {
	var Pwd_path []string
	for image := range images {
		_, _, element := UploadImage(rw, images[image], pwd, user_id, strconv.Itoa(image))

		Pwd_path = append(Pwd_path, element)
	}

	return nil, true, Pwd_path
}

// Генерация имени файла на основе временной метки
func generateFileName(extension, user_id, index string) string {
	timestamp := time.Now().Format("010203084503") // ГГГГММДДччммсс

	return fmt.Sprintf("image_%s_%s_%s.%s", timestamp, user_id, index, extension)
}

func UploadImage(rw http.ResponseWriter, imageBase64, directory, user_id, index string) (error, bool, string) {
	// Проверяем, содержит ли строка базовые метаданные
	if strings.HasPrefix(imageBase64, "data:image/png;base64,") {
		// Извлекаем данные после запятой
		commaIndex := strings.Index(imageBase64, ",")
		if commaIndex == -1 {
			http.Error(rw, "Invalid base64 data", http.StatusBadRequest)
			return fmt.Errorf("invalid base64 data"), false, ""
		}
		imageBase64 = imageBase64[commaIndex+1:]
	}

	// Декодируем данные base64
	data, err := base64.StdEncoding.DecodeString(imageBase64)
	errorr(err)

	// Генерируем уникальное имя файла
	fileName := generateFileName("png", user_id, index)

	// Полный путь до файла
	filePath := filepath.Join(directory, fileName)

	// Проверяем и создаём директорию, если она не существует
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755) // Создаем директорию с правами 0755
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error creating directory %s: %v", directory, err), http.StatusInternalServerError)
			return err, false, ""
		}
	}

	// Создаем файл на сервере для сохранения декодированного файла
	dst, err := os.Create(filePath) // Используем безопасное имя файла
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error creating file %s: %v", filePath, err), http.StatusInternalServerError)
		return err, false, ""
	}
	defer dst.Close()

	// Записываем данные в файл
	if _, err := dst.Write(data); err != nil {
		http.Error(rw, fmt.Sprintf("Error writing to file %s: %v", fileName, err), http.StatusInternalServerError)
		return err, false, ""
	}

	return nil, true, filePath
}

func UploadAvatar(rw http.ResponseWriter, imageBase64, directory, user_id, index string) (error, bool, string) {
	// Проверяем, содержит ли строка базовые метаданные
	if strings.HasPrefix(imageBase64, "data:image/png;base64,") {
		// Извлекаем данные после запятой
		commaIndex := strings.Index(imageBase64, ",")
		if commaIndex == -1 {
			http.Error(rw, "Invalid base64 data", http.StatusBadRequest)
			return fmt.Errorf("invalid base64 data"), false, ""
		}
		imageBase64 = imageBase64[commaIndex+1:]
	}

	// Декодируем данные base64
	data, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error decoding base64: %v", err), http.StatusInternalServerError)
		return err, false, ""
	}

	fmt.Println("Decoded data length:", len(data))

	// Генерируем уникальное имя файла
	fileName := generateFileName("png", user_id, index)

	// Полный путь до файла
	filePath := filepath.Join(directory, fileName)

	// Проверяем и создаём директорию, если она не существует
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755) // Создаем директорию с правами 0755
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error creating directory %s: %v", directory, err), http.StatusInternalServerError)
			return err, false, ""
		}
	}

	// Создаем файл на сервере для сохранения декодированного файла
	fmt.Println("Creating file at path:", filePath)
	dst, err := os.Create(filePath) // Используем безопасное имя файла
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error creating file %s: %v", filePath, err), http.StatusInternalServerError)
		return err, false, ""
	}
	defer dst.Close()

	// Записываем данные в файл
	fmt.Println("Writing data to file")
	if _, err := dst.Write(data); err != nil {
		http.Error(rw, fmt.Sprintf("Error writing to file %s: %v", fileName, err), http.StatusInternalServerError)
		return err, false, ""
	}

	fmt.Println("File successfully created at:", filePath)

	return nil, true, filePath
}

// Функция для массовой загрузки множества видеофайлов
func UploadVideosMass(rw http.ResponseWriter, videos []string, pwd, user_id string) (_ error, _ bool, file_path []string) {
	for i := range videos {
		err, success, file := UploadVideo(rw, videos[i], pwd, user_id, strconv.Itoa(i))
		if err != nil || !success {
			return err, false, nil
		}
		file_path = append(file_path, file)
	}
	return nil, true, file_path
}

// // Функция для загрузки одного видеофайла
func UploadVideo(rw http.ResponseWriter, videoBase64, directory, user_id, index string) (error, bool, string) {
	// Проверяем, содержит ли строка базовые метаданные для MP4
	if strings.HasPrefix(videoBase64, "data:video/mp4;base64,") {
		// Извлекаем данные после запятой
		commaIndex := strings.Index(videoBase64, ",")
		if commaIndex != -1 {
			videoBase64 = videoBase64[commaIndex+1:]
		}
	}

	// Декодируем данные base64
	data, err := base64.StdEncoding.DecodeString(videoBase64)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error decoding base64: %v", err), http.StatusInternalServerError)
		return err, false, ""
	}

	// Генерируем уникальное имя файла
	fileName := generateFileName("mp4", user_id, index)

	// Полный путь до файла
	filePath := filepath.Join(directory, fileName)

	// Проверяем и создаём директорию, если она не существует
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Error creating directory %s: %v", directory, err), http.StatusInternalServerError)
			return err, false, ""
		}
	}

	// Создаем файл на сервере для сохранения декодированного видео
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Error creating file %s: %v", filePath, err), http.StatusInternalServerError)
		return err, false, ""
	}
	defer dst.Close()

	// Записываем данные в файл
	if _, err := dst.Write(data); err != nil {
		http.Error(rw, fmt.Sprintf("Error writing to file %s: %v", fileName, err), http.StatusInternalServerError)
		return err, false, ""
	}

	return nil, true, filePath
}

func DeleteImage(rw http.ResponseWriter, pwd, imageName string) (error, bool) {
	// Формируем полный путь к файлу
	filePath := pwd + imageName

	// Удаляем файл
	err := os.Remove(filePath)
	errorr(err)

	return nil, true
}

func (a *MyApp) SignupUserByEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) error {
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
	CodeNum := 777 // Здесь лучше использовать случайный код
	message := []byte(subject + "\n" + body + strconv.Itoa(CodeNum))

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
	keshData, err := json.Marshal(Kesh{Email: kesh.Email, Code: CodeNum})
	errorr(err)

	// Генерация JWT токена
	ValidToken_jwt, err := jwt.GenerateJWT("jwt_for_reg", 0)
	errorr(err)

	// Установка куки
	livingTime := 40 * time.Minute //не смог найти чего-то получше
	expiration := time.Now().Add(livingTime)
	cookie := http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(ValidToken_jwt),
		Expires:  expiration,
		Path:     "/",             // Убедитесь, что путь корректен
		Domain:   "185.112.83.36", // IP-адрес вашего сервера
		HttpOnly: true,
		Secure:   false, // Для HTTP можно оставить false
		SameSite: http.SameSiteLaxMode,
	}

	fmt.Printf("Кука установлена: %v\n", cookie)

	// Сохраняем код подтверждения в Redis с TTL 40 минут
	err = redisClient.Set(ctx, ValidToken_jwt, keshData, livingTime).Err()
	if err != nil {
		log.Fatal("Ошибка при сохранении кода в Redis:", err)
		return err
	}

	type Data struct {
		CodeNum        int    `json:"CodeNum"`
		ValidToken_jwt string `json:"ValidToken_jwt"`
	}

	rw.WriteHeader(http.StatusOK)
	type Response struct {
		Status  string `json:"status"`
		Data    Data   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	// Лог с контекстом
	logger.Info().
		Str("service", "login").
		Int("port", 8080).
		Msg("User enter code and email")

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Почта не принята",
		}

		// Лог с контекстом
		logger.Info().
			Str("service", "login").
			Int("port", 8080).
			Msg("User enter code and email")

		json.NewEncoder(rw).Encode(response)

		return err
	} else if ValidToken_jwt == "" {

	}

	response := Response{
		Status:  "success",
		Data:    Data{CodeNum, ValidToken_jwt},
		Message: "Почта принята",
	}

	json.NewEncoder(rw).Encode(response)

	return err
}

func (a *MyApp) SignupUserByPhonePOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) error {
	type Kesh struct {
		Phone_number string `json:"Phone_num"`
		Code         int    `json:"Code"`
	}

	var kesh Kesh

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&kesh)
	errorr(err)

	CodeNum := 777 // Здесь лучше использовать случайный код

	// Преобразуем структуру kesh в JSON
	keshData, err := json.Marshal(Kesh{Phone_number: kesh.Phone_number, Code: CodeNum})
	errorr(err)

	// Генерация JWT токена
	ValidToken_jwt, err := jwt.GenerateJWT("jwt_for_reg", 0)
	errorr(err)

	// Установка куки
	livingTime := 40 * time.Minute //не смог найти чего-то получше
	expiration := time.Now().Add(livingTime)
	cookie := http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(ValidToken_jwt),
		Expires:  expiration,
		Path:     "/",             // Убедитесь, что путь корректен
		Domain:   "185.112.83.36", // IP-адрес вашего сервера
		HttpOnly: true,
		Secure:   false, // Для HTTP можно оставить false
		SameSite: http.SameSiteLaxMode,
	}

	fmt.Printf("Кука установлена: %v\n", cookie)

	// Сохраняем код подтверждения в Redis с TTL 40 минут
	err = redisClient.Set(ctx, ValidToken_jwt, keshData, livingTime).Err()
	if err != nil {
		log.Fatal("Ошибка при сохранении кода в Redis:", err)
		return err
	}

	type Data struct {
		CodeNum        int    `json:"CodeNum"`
		ValidToken_jwt string `json:"ValidToken_jwt"`
	}

	rw.WriteHeader(http.StatusOK)
	type Response struct {
		Status  string `json:"status"`
		Data    Data   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ValidToken_jwt == "" {
		response := Response{
			Status:  "fatal",
			Message: "Почта не принята",
		}

		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    Data{CodeNum, ValidToken_jwt},
		Message: "Почта принята",
	}

	json.NewEncoder(rw).Encode(response)

	return err
}

func (a *MyApp) EnterCodeFromEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) error {
	type RegCode struct {
		Email string `json:"Email"`
		Code  int    `json:"Code"`
	}

	var email RegCode

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(rw, "Кука не найдена", http.StatusUnauthorized)
			return err
		}
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}

	if token.Value == "" {
		http.Error(rw, "Кука пуста", http.StatusUnauthorized)
		return fmt.Errorf("пустая кука")
	}

	token_flag, _ := jwt.IsAuthorized(token.Value)
	if !token_flag {
		fmt.Println("Что-то не так с токеном")
		http.Error(rw, "Неверный токен", http.StatusUnauthorized)
		return fmt.Errorf("неверный токен")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token.Value).Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return err
	} else if err != nil {
		log.Println("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh RegCode
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Println("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Создаем структуру ответа
	type Response struct {
		Status  string `json:"status"`
		Data    string `json:"data,omitempty"`
		Message string `json:"message"`
	}

	var response Response

	fmt.Println(email.Code)
	fmt.Println(storedKesh.Code)

	// Проверяем код
	if storedKesh.Code == storedKesh.Code {
		response = Response{
			Status:  "success",
			Data:    storedKesh.Email,
			Message: "Код принят",
		}
	} else {
		response = Response{
			Status:  "fatal",
			Message: "Неверный код",
		}
	}

	// Отправляем ответ
	err = json.NewEncoder(rw).Encode(response)
	if err != nil {
		log.Println("Ошибка при отправке ответа:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	return nil
}

func (a *MyApp) EnterCodeFromPhonePOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) error {
	type Kesh struct {
		Phone string `json:"Phone_num"`
		Code  int    `json:"Code"`
	}

	var phone Reg_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&phone)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return err
	}

	// Попытка прочитать куку
	token, err := internal.ReadCookie("token", r)
	errorr(err)

	token_flag, _ := jwt.IsAuthorized(token)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token).Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return err
	} else if err != nil {
		log.Println("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Десериализуем JSON обратно в структуру kesh
	var storedKesh Kesh
	err = json.Unmarshal([]byte(keshData), &storedKesh)
	if err != nil {
		log.Println("Ошибка при десериализации данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	// Создаем структуру ответа
	type Response struct {
		Status  string `json:"status"`
		Data    string `json:"data,omitempty"`
		Message string `json:"message"`
	}

	var response Response

	// Проверяем код
	if phone.Reg_code == storedKesh.Code {
		fmt.Println("Phone", storedKesh.Phone)
		response = Response{
			Status:  "success",
			Data:    storedKesh.Phone,
			Message: "Код принят",
		}
		rw.WriteHeader(http.StatusOK)
	} else {
		response = Response{
			Status:  "fatal",
			Message: "Неверный код",
		}
		rw.WriteHeader(http.StatusUnauthorized) // Для неверного кода лучше использовать статус 401 Unauthorized
	}

	// Отправляем ответ
	err = json.NewEncoder(rw).Encode(response)
	if err != nil {
		log.Println("Ошибка при отправке ответа:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return err
	}

	return nil
}

func (a *MyApp) SignupLegalEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) (err error) {
	type Kesh struct {
		Email string
		Code  int
	}

	var user LegalUser

	// Парсинг JSON-запроса
	err = json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	flag := validatePassword(rw, user.Password_hash)

	if !flag {
		return
	}

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, _ := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token.Value).Result()
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

	var pwd = "/root/home/beeline_project/media/user/"

	err, image_flag, file_path := UploadAvatar(rw, user.Avatar, pwd, strings.Split(storedKesh.Email, ".")[0], "ava")
	errorr(err)

	if image_flag {
		err = repo.SigLegalUserEmailSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Ind_num_taxp,
			user.Name_of_company,
			user.Address_name,
			storedKesh.Email,
			user.Password_hash,

			user.Data,
			file_path,
		)
	}

	return err
}

func (a *MyApp) SignupLegalPhonePOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) (err error) {
	type Kesh struct {
		Phone string `json:"Phone_num"`
		Code  int    `json:"Code"`
	}

	var user LegalUser

	// Парсинг JSON-запроса
	err = json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	flag := validatePassword(rw, user.Password_hash)

	if !flag {
		return
	}

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, _ := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token.Value).Result()
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

	var pwd = "/root/home/beeline_project/media/user/"

	err, image_flag, file_path := UploadAvatar(rw, user.Avatar, pwd, strings.Split(storedKesh.Phone, ".")[0], "ava")
	errorr(err)
	if image_flag {
		err = repo.SigLegalUserPhoneSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Ind_num_taxp,
			user.Name_of_company,
			user.Address_name,
			storedKesh.Phone,
			user.Password_hash,

			user.Data,
			file_path,
		)
	}

	return err
}

func (a *MyApp) SignupNaturEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) (err error) {
	type Kesh struct {
		Email string
		Code  int
	}
	fmt.Println("dwa")

	var user NaturUser

	// Парсинг JSON-запроса
	err = json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	flag := validatePassword(rw, user.Password_hash)

	if !flag {
		return
	}

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, _ := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token.Value).Result()
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

	var pwd = "/root/home/beeline_project/media/user/"
	fmt.Println("dwa1")

	err, image_flag, file_path := UploadAvatar(rw, user.Avatar, pwd, strings.Split(storedKesh.Email, ".")[0], "ava")
	errorr(err)
	if image_flag {
		err = repo.SigNaturUserEmailSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Name,
			user.Surname,
			user.Patronymic,
			storedKesh.Email,
			user.Password_hash,

			user.Data,
			file_path)

		errorr(err)
	}

	return err
}

func (a *MyApp) SignupNaturPhonePOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) (err error) {
	type Kesh struct {
		Phone string `json:"Phone_num"`
		Code  int    `json:"Code"`
	}

	var user NaturUser

	// Парсинг JSON-запроса
	err = json.NewDecoder(r.Body).Decode(&user)
	errorr(err)

	flag := validatePassword(rw, user.Password_hash)

	if !flag {
		return
	}

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, _ := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	keshData, err := redisClient.Get(ctx, token.Value).Result()
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

	var pwd = "/root/home/beeline_project/media/user/"

	err, image_flag, file_path := UploadAvatar(rw, user.Avatar, pwd, strings.Split(storedKesh.Phone, ".")[0], "ava")
	errorr(err)

	if image_flag {
		err = repo.SigNaturUserPhoneSQL(
			ctx,
			a.app.Repo.Pool,
			rw,
			r,
			user.Name,
			user.Surname,
			user.Patronymic,
			storedKesh.Phone,
			user.Password_hash,

			user.Data,
			file_path,
		)
	}

	return err
}

func (a *MyApp) EditingLegalUserDataPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	flag, user_id := jwt.IsAuthorized(token)

	if flag {
		if legalUser.Avatar != "" {
			file_path := "/root/home/beeline_project/media/user/"
			err, ava_flag, pwd := UploadAvatar(rw, legalUser.Avatar, file_path, strconv.Itoa(user_id), "ava")

			if ava_flag {
				err = repo.EditingLegalUserAvaSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, pwd)

				errorr(err)
			}
		}

		if legalUser.Ind_num_taxp != 0 {
			err = repo.EditingLegalUserIndNumSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, legalUser.Ind_num_taxp)

			errorr(err)
		}
		if legalUser.Name_of_company != "" {
			err = repo.EditingLegalUserNameCompSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, legalUser.Name_of_company)

			errorr(err)
		}
		if legalUser.Address_name != "" {
			err = repo.EditingLegalUserAddressNameSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, legalUser.Address_name)

			errorr(err)
		}
	}
}

func (a *MyApp) EditingNaturUserDataPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	flag, user_id := jwt.IsAuthorized(token)

	if flag {
		if naturUser.Avatar != "" {
			file_path := "/root/home/beeline_project/media/user/"
			err, ava_flag, pwd := UploadAvatar(rw, naturUser.Avatar, file_path, strconv.Itoa(user_id), "ava")

			if ava_flag {
				err = repo.EditingNaturUserAvaSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, pwd)

				errorr(err)
			}
		}

		if naturUser.Surname != "" {
			err = repo.EditingNaturUserSurnameSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, naturUser.Surname)

			errorr(err)
		}
		if naturUser.Name != "" {
			err = repo.EditingNaturUserNameSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, naturUser.Name)

			errorr(err)
		}
		if naturUser.Patronymic != "" {
			err = repo.EditingNaturUserPatronomSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, naturUser.Patronymic)

			errorr(err)
		}
	}
}

type Email_name struct {
	Email_name string `json:"email_name"`
}

type Phone_num struct {
	Phone_num string `json:"phone_num"`
}

func (a *MyApp) SendCodForEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
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

	// Генерация JWT токена
	ValidToken_jwt, err := jwt.GenerateJWT("jwt_for_proof", 0)
	errorr(err)

	fmt.Println("JWT токен в строковом формате: ", ValidToken_jwt)

	// Установка куки
	livingTime := 10 * time.Minute //не смог найти чего-то получше
	expiration := time.Now().Add(livingTime)
	cookie := http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(ValidToken_jwt),
		Expires:  expiration,
		Path:     "/",             // Убедитесь, что путь корректен
		Domain:   "185.112.83.36", // IP-адрес вашего сервера
		HttpOnly: true,
		Secure:   false, // Для HTTP можно оставить false
		SameSite: http.SameSiteLaxMode,
	}

	fmt.Printf("Кука установлена: %v\n", cookie)

	CodeNum := 777 // Здесь лучше использовать случайный код

	type Kesh struct {
		CodeNum    int
		Email_name string
	}

	// Преобразуем структуру kesh в JSON
	keshData, err := json.Marshal(Kesh{CodeNum: CodeNum, Email_name: email.Email_name})
	errorr(err)

	// Сохраняем код подтверждения в Redis с TTL 10 минут
	err = redisClient.Set(ctx, ValidToken_jwt, keshData, 10*time.Minute).Err()
	errorr(err)

	type Data struct {
		CodeNum        int    `json:"CodeNum"`
		Email_name     string `json:"Email_name"`
		ValidToken_jwt string `json:"ValidToken_jwt"`
	}

	rw.WriteHeader(http.StatusOK)
	type Response struct {
		Status  string `json:"status"`
		Data    Data   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ValidToken_jwt == "" {
		response := Response{
			Status:  "fatal",
			Message: "Почта не принята",
		}

		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "success",
		Data:    Data{CodeNum, email.Email_name, ValidToken_jwt},
		Message: "Почта принята",
	}

	json.NewEncoder(rw).Encode(response)

	return
}

func (a *MyApp) EnterCodFromEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var email Reg_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Попытка прочитать куку
	token_kesh, err := r.Cookie("token")
	errorr(err)

	token_flag, _ = jwt.IsAuthorized(token_kesh.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	kash, err := redisClient.Get(ctx, token_kesh.Value).Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Структура для хранения данных
	var data map[string]interface{}

	// Парсинг JSON в структуру
	err = json.Unmarshal([]byte(kash), &data)
	if err != nil {
		fmt.Println("Ошибка парсинга JSON:", err)
		return
	}

	// Извлечение значения "Phone_num"
	if phoneNum, ok := data["Email_name"].(string); ok {
		if value, ok := data["CodeNum"].(float64); ok {
			if email.Reg_code == int(value) {
				err := repo.EnterCodFromEmailSQL(ctx, rw, a.app.Repo.Pool, user_id, phoneNum)
				errorr(err)
			}
		} else {
			fmt.Println("CodeNum не является int")
		}
	} else {
		fmt.Println("Email_name не найден или имеет неверный тип")
	}
}

func (a *MyApp) SendCodForPhoneNumPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var phone Phone_num

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&phone)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерация JWT токена
	ValidToken_jwt, err := jwt.GenerateJWT("jwt_for_proof", 0)
	errorr(err)

	fmt.Println("JWT токен в строковом формате: ", ValidToken_jwt)

	// Установка куки
	livingTime := 10 * time.Minute //не смог найти чего-то получше
	expiration := time.Now().Add(livingTime)
	cookie := http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(ValidToken_jwt),
		Expires:  expiration,
		Path:     "/",             // Убедитесь, что путь корректен
		Domain:   "185.112.83.36", // IP-адрес вашего сервера
		HttpOnly: true,
		Secure:   false, // Для HTTP можно оставить false
		SameSite: http.SameSiteLaxMode,
	}

	CodeNum := 777

	fmt.Printf("Кука установлена: %v\n", cookie)
	fmt.Print("JWT:  ")
	fmt.Println(ValidToken_jwt)

	type Kesh struct {
		CodeNum   int
		Phone_num string
	}

	// Преобразуем структуру kesh в JSON
	keshData, err := json.Marshal(Kesh{CodeNum: CodeNum, Phone_num: phone.Phone_num})
	errorr(err)

	// Сохраняем код подтверждения в Redis с TTL 10 минут
	err = redisClient.Set(ctx, ValidToken_jwt, keshData, 10*time.Minute).Err()
	errorr(err)

	type Data struct {
		CodeNum        int    `json:"CodeNum"`
		Phone_num      string `json:"Phone_num"`
		ValidToken_jwt string `json:"ValidToken_jwt"`
	}

	rw.WriteHeader(http.StatusOK)
	type Response struct {
		Status  string `json:"status"`
		Data    Data   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ValidToken_jwt == "" || phone.Phone_num == "" {
		response := Response{
			Status:  "fatal",
			Message: "Телефон не принят",
		}

		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "success",
		Data:    Data{CodeNum, phone.Phone_num, ValidToken_jwt},
		Message: "Почта принята",
	}

	json.NewEncoder(rw).Encode(response)

	return
}

func (a *MyApp) EnterCodFromPhoneNumPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var phone Reg_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&phone)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Попытка прочитать куку
	token_kesh, err := r.Cookie("token_kesh")
	errorr(err)

	token_flag, _ = jwt.IsAuthorized(token_kesh.Value)

	if !token_flag {
		fmt.Println("Что-то не так с токеном")
	}

	// Получаем данные из Redis
	kash, err := redisClient.Get(ctx, token_kesh.Value).Result()
	if err == redis.Nil {
		log.Println("Ключ не найден")
		http.Error(rw, "Код не найден или истек", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Fatal("Ошибка при получении данных из Redis:", err)
		http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Структура для хранения данных
	var data map[string]interface{}

	// Парсинг JSON в структуру
	err = json.Unmarshal([]byte(kash), &data)
	if err != nil {
		fmt.Println("Ошибка парсинга JSON:", err)
		return
	}

	// Извлечение значения "Phone_num"
	if phoneNum, ok := data["Phone_num"].(string); ok {
		if value, ok := data["CodeNum"].(float64); ok {
			if phone.Reg_code == int(value) {
				err := repo.EnterCodFromPhoneNumSQL(ctx, rw, a.app.Repo.Pool, user_id, phoneNum)
				errorr(err)
			}
		} else {
			fmt.Println("CodeNum не является int")
		}
	} else {
		fmt.Println("Phone_num не найден или имеет неверный тип")
	}
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *MyApp) LoginPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var login Login

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&login)
	errorr(err)

	// Лог с контекстом
	logger.Info().
		Str("service", "login").
		Int("port", 8080).
		Msg("User login")

	err = repo.LoginSQL(a.app.Ctx, a.app.Repo.Pool, rw, login.Login, login.Password, logger)

	errorr(err)
}

type ProductList struct {
	Ads_list []int `json:"Ads_list"`
}

func (a *MyApp) ProductListPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var productList ProductList

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&productList)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	tokenn, err := r.Cookie("token")
	errorr(err)

	// Парсинг и верификация токена
	token, err := jwt_from_prod_list.Parse(tokenn.Value, func(token *jwt_from_prod_list.Token) (interface{}, error) {
		// Проверка метода подписи токена
		if _, ok := token.Method.(*jwt_from_prod_list.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return "superSecretKey", nil
	})

	errorr(err)

	var user_id *int

	// Извлечение полезной нагрузки (claims) из токена
	if claims, ok := token.Claims.(jwt_from_prod_list.MapClaims); ok {
		// Извлекаем user_id из claims
		if userIDFloat, ok := claims["id"].(float64); ok {
			user_id = new(int)          // Создаем указатель
			*user_id = int(userIDFloat) // Записываем значение по указателю
		}
	}

	err = repo.ProductListSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id, productList.Ads_list)

	errorr(err)
}

type PrintAds struct {
	Ads_id int `json:"Ads_id"`
}

func (a *MyApp) PrintAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var printAds PrintAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&printAds)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.PrintAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, printAds.Ads_id)

	errorr(err)
}

type SortProductListAll struct {
	Category []int     `json:"Category"`
	LowNum   int       `json:"LowNum"`
	HigNum   int       `json:"HigNum"`
	LowDate  int64     `json:"LowDate"`
	HigDate  int64     `json:"HigDate"`
	Position []float64 `json:"Position"`
	Distance int       `json:"Distance"`
	Rating   int       `json:"Rating"`
}

func (a *MyApp) SortProductListDailyRatePOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListDailyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, sortProductList.Category, sortProductList.LowNum, sortProductList.HigNum, lowDate, higDate, sortProductList.Position, sortProductList.Distance, sortProductList.Rating)

	errorr(err)
}

func (a *MyApp) SortProductListHourlyRatePOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListHourlyRateSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, sortProductList.Category, sortProductList.LowNum, sortProductList.HigNum, lowDate, higDate, sortProductList.Position, sortProductList.Distance, sortProductList.Rating)

	errorr(err)
}

type Ads struct {
	Image []string `json:"Image"`

	Id          int     `json:"Id"`
	Title       string  `json:"Title"`
	Description string  `json:"Description"`
	Hourly_rate int     `json:"Hourly_rate"`
	Daily_rate  int     `json:"Daily_rate"`
	Category_id int     `json:"Category_id"`
	PositionX   float64 `json:"PositionX"`
	PositionY   float64 `json:"PositionY"`
}

func (a *MyApp) SignupAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads Ads

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	var pwd = "/root/home/beeline_project/media/ads/"

	err, image_flag, Pwd_mass := UploadImagesMass(rw, ads.Image, pwd, strconv.Itoa(user_id))
	errorr(err)

	if token_flag && image_flag {
		err = repo.SignupAdsSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			ads.Title,
			ads.Description,
			ads.Hourly_rate,
			ads.Daily_rate,
			user_id,
			ads.Category_id,
			ads.PositionX,
			ads.PositionY,
			time.Now(),
			ads.Image,
			Pwd_mass,
			pwd)

		errorr(err)
	}
}

func (a *MyApp) EditAdsListPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Ads struct {
		Ads_id int `json:"Ads_id"`
	}
	var ads Ads

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		err := repo.EditAdsListSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			ads.Ads_id,
			user_id)

		errorr(err)
	} else {

		json.NewEncoder(rw).Encode(Response{
			Status:  "fatal",
			Message: "Что-то не то с токеном",
		})
	}
}

func (a *MyApp) UpdAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Ads struct {
		Ads_id             int          `json:"Ads_id"`
		Id_images_from_del []int        `json:"Images_from_del"`
		Title              string       `json:"Title"`
		Description        string       `json:"Description"`
		Hourly_rate        int          `json:"Hourly_rate"`
		Daily_rate         int          `json:"Daily_rate"`
		Category_id        int          `json:"Category_id"`
		Position           pgtype.Point `json:"Position"`
		Images             []string     `json:"Images"`
	}
	var ads Ads
	fmt.Println(ads)

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		if len(ads.Id_images_from_del) != 0 {
			for i := range len(ads.Id_images_from_del) {
				repo.UpdAdsDelImgSQL(ctx, rw, a.app.Repo.Pool, ads.Id_images_from_del[i])
			}
		}
		if ads.Title != "" {
			repo.UpdAdsTitleSQL(ctx, rw, a.app.Repo.Pool, ads.Title, ads.Ads_id, user_id)
		}
		if ads.Description != "" {
			repo.UpdAdsDescriptionSQL(ctx, rw, a.app.Repo.Pool, ads.Description, ads.Ads_id, user_id)
		}
		if ads.Hourly_rate != 0 {
			repo.UpdAdsHourly_rateSQL(ctx, rw, a.app.Repo.Pool, ads.Hourly_rate, ads.Ads_id, user_id)
		}
		if ads.Daily_rate != 0 {
			repo.UpdAdsDaily_rateSQL(ctx, rw, a.app.Repo.Pool, ads.Daily_rate, ads.Ads_id, user_id)
		}
		if ads.Category_id != 0 {
			repo.UpdAdsCategory_idSQL(ctx, rw, a.app.Repo.Pool, ads.Category_id, ads.Ads_id, user_id)
		}
		// if ads.Location != "" {
		// 	repo.UpdAdsLocationSQL(ctx, rw, a.app.Repo.Pool, ads.Location, ads.Ads_id, user_id)
		// }
		// if ads.Position.Status != pgtype.Null {
		// 	repo.UpdAdsPositionSQL(ctx, rw, a.app.Repo.Pool, ads.Position, ads.Ads_id, user_id)
		// }
		if len(ads.Images) != 0 {
			var pwd = "/root/home/beeline_project/media/ads/"
			for i := range len(ads.Images) {
				err, flag, file_path := UploadImage(rw, ads.Images[i], pwd, strconv.Itoa(user_id), strconv.Itoa(i))
				errorr(err)

				if flag {
					repo.UpdAdsAddImgSQL(ctx, rw, a.app.Repo.Pool, file_path, ads.Ads_id, user_id)
				}
			}
		}
	} else {
		json.NewEncoder(rw).Encode(Response{
			Status:  "fatal",
			Message: "Что-то не то с токеном",
		})
	}

	// var pwd = "/home/beeline_project/media/ads/"

	// err, image_flag, Pwd_mass := UploadImagesMass(rw, ads.Image, pwd, strconv.Itoa(user_id))
	// errorr(err)

	// if token_flag && image_flag {
	// 	err := repo.UpdAdsSQL(
	// 		a.app.Ctx,
	// 		rw,
	// 		a.app.Repo.Pool,
	// 		ads.Title,
	// 		ads.Description,
	// 		ads.Hourly_rate,
	// 		ads.Daily_rate,
	// 		user_id,
	// 		ads.Category_id,
	// 		ads.Location,
	// 		ads.Id,
	// 		time.Now(),
	// 		Pwd_mass,
	// 		ads.Image)

	// 	errorr(err)
	// }
}

func (a *MyApp) DelAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type DelAds struct {
		Ads_id int `json:"Ads_id"`
	}

	var ads DelAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	flag, user_id := jwt.IsAuthorized(token)
	if flag {
		err := repo.DelAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, ads.Ads_id, user_id)
		if err != nil {
			fmt.Errorf("Ошибка создания объявления: %v", err)
			return
		}
	} else {
		json.NewEncoder(rw).Encode(Response{
			Status:  "fatal",
			Message: "Что-то не то с токеном",
		})
	}
}

type FavAds struct {
	User_id int `json:"User_id"`
	Ads_id  int `json:"Ads_id"`
}

func (a *MyApp) SigFavAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads FavAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	flag, user_id := jwt.IsAuthorized(token)
	if flag {
		fmt.Println(user_id)
		err := repo.SigFavAdsSQL(a.app.Ctx, a.app.Repo.Pool, rw, user_id, ads.Ads_id)

		errorr(err)
	}
}

func (a *MyApp) DelFavAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads FavAds

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	flag, user_id := jwt.IsAuthorized(token)
	if flag {
		err := repo.DelFavAdsSQL(a.app.Ctx, a.app.Repo.Pool, rw, user_id, ads.Ads_id)

		errorr(err)
	}
}

type Title struct {
	Title string `json:"Title"`
}

func (a *MyApp) SearchForTechPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) SortProductListCategoriezPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sortProductList SortProductListCategoriez

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sortProductList)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.SortProductListCategoriezSQL(a.app.Ctx, rw, a.app.Repo.Pool, sortProductList.Category)

	errorr(err)
}

type SigChat struct {
	Id_ads int `json:"Ads_id"`
}

func (a *MyApp) ChatButtonInAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sigChat SigChat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sigChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	errorr(err)

	flag, user_id := jwt.IsAuthorized(token.Value)

	if flag {
		err = repo.ChatButtonInAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, sigChat.Id_ads)

		errorr(err)
	}
}

func (a *MyApp) SigChatPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sigChat SigChat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sigChat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	errorr(err)

	flag, user_id := jwt.IsAuthorized(token.Value)

	if flag {
		err = repo.SigChatSQL(a.app.Ctx, rw, user_id, sigChat.Id_ads, a.app.Repo.Pool)

		errorr(err)
	}
}

type Chat struct {
	Id_chat int `json:"Id_chat"`
}

func (a *MyApp) OpenChatPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	flag, user_id := jwt.IsAuthorized(token)

	if flag {
		err = repo.OpenChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, user_id)
	}

	errorr(err)
}

type SendMessAndImg struct {
	Id_chat int      `json:"Id_chat"`
	Text    string   `json:"Text"`
	Images  []string `json:"Image"`
}

type SendImg struct {
	Id_chat int      `json:"Id_chat"`
	Images  []string `json:"Image"`
}

type SendMess struct {
	Id_chat int    `json:"Id_chat"`
	Text    string `json:"Text"`
}

type SendMessAndVideo struct {
	Id_chat int      `json:"Id_chat"`
	Text    string   `json:"Text"`
	Videos  []string `json:"Videos"`
}

type SendVideo struct {
	Id_chat int      `json:"Id_chat"`
	Videos  []string `json:"Videos"`
}

func (a *MyApp) SendMessageAndImagePOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sendMess SendMessAndImg

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)

	var pwd = "/home/beeline_project/media/chat/"

	err, img_flag, file_paths := UploadImagesMass(rw, sendMess.Images, pwd, strconv.Itoa(user_id)) // добавляет изображения
	errorr(err)

	if flag && img_flag {
		err = repo.SendMessageAndMediaSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			sendMess.Id_chat,
			user_id,
			sendMess.Text,
			file_paths)

		errorr(err)
	}
}

func (a *MyApp) SendImagePOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sendMess SendImg

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)

	var pwd = "/home/beeline_project/media/chat/"

	err, img_flag, file_paths := UploadImagesMass(rw, sendMess.Images, pwd, strconv.Itoa(user_id)) // добавляет изображения
	errorr(err)

	if flag && img_flag {
		err = repo.SendImageSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			sendMess.Id_chat,
			user_id,
			file_paths)

		errorr(err)

	}
}

func (a *MyApp) SendMessagePOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sendMess SendMess

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)

	if flag {
		err = repo.SendMessageSQL(a.app.Ctx, rw, a.app.Repo.Pool, sendMess.Id_chat, user_id, sendMess.Text)

		errorr(err)
	}
}

func (a *MyApp) SendMessageAndVideoPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sendMess SendMessAndVideo

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)

	var pwd = "/home/beeline_project/media/chat/"

	err, img_flag, file_paths := UploadVideosMass(rw, sendMess.Videos, pwd, strconv.Itoa(user_id)) // добавляет изображения
	errorr(err)

	if flag && img_flag {
		err = repo.SendMessageAndMediaSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			sendMess.Id_chat,
			user_id,
			sendMess.Text,
			file_paths)

		errorr(err)
	}
}

func (a *MyApp) SendVideoPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sendMess SendVideo

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sendMess)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)
	errorr(err)

	flag, user_id := jwt.IsAuthorized(token)

	var pwd = "/home/beeline_project/media/chat/"

	err, img_flag, file_paths := UploadVideosMass(rw, sendMess.Videos, pwd, strconv.Itoa(user_id)) // добавляет изображения
	errorr(err)

	if flag && img_flag {
		err = repo.SendVideoSQL(
			a.app.Ctx,
			rw,
			a.app.Repo.Pool,
			sendMess.Id_chat,
			user_id,
			file_paths)

		errorr(err)
	}
}

func (a *MyApp) SigDisputInChatPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var chat Chat

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&chat)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// token, err := internal.ReadCookie("token", r)

	// if err != nil {
	// 	fmt.Errorf("Ошибка создания объявления: %v", err)
	// 	return
	// } else {
	// 	flag , user_id := jwt.IsAuthorized(rw, token)

	// err = repo.SigDisputInChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, chat.Id_chat, 128)

	// errorr(err)
}

type SigReview struct {
	Order_id int    `json:"Order_id"`
	Rating   int    `json:"Rating"`
	Comment  string `json:"Comment"`
	State    int    `json:"State"`
}

func (a *MyApp) SigReviewPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var sigReview SigReview

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&sigReview)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Errorf("Ошибка создания объявления: %v", err)
		return
	} else {
		flag, user_id := jwt.IsAuthorized(token)
		if flag {
			err = repo.SigReviewSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, sigReview.Order_id, sigReview.Rating, sigReview.Comment, sigReview.State)

			errorr(err)
		}
	}
}

type UpdReview struct {
	Review_id int    `json:"Review_id"`
	Rating    int    `json:"Rating"`
	Comment   string `json:"Comment"`
}

func (a *MyApp) UpdReviewPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var updReview UpdReview

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&updReview)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		return
	} else {

		flag, user_id := jwt.IsAuthorized(token)
		if flag {
			err = repo.UpdReviewSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, updReview.Review_id, updReview.Rating, updReview.Comment)

			errorr(err)
		}
	}
}

func (a *MyApp) MediatorStartWorkingPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	flag, mediator_id := jwt.IsAuthorized(token)

	if flag {
		err = repo.MediatorStartWorkingSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, mediator_id)
	}
	errorr(err)
}

func (a *MyApp) MediatorEnterInChatPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

	flag, mediator_id := jwt.IsAuthorized(token)

	if flag {
		err = repo.MediatorEnterInChatSQL(a.app.Ctx, rw, a.app.Repo.Pool, openChat.Id_chat, mediator_id)
	}

	errorr(err)
}

func (a *MyApp) MediatorFinishJobInChatPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) TransactionToAnotherPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) TransactionToSomethingPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) TransactionToReturnAmountPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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
	Total_price int   `json:"Total_price"`
	Starts_at   int64 `json:"Starts_at"`
	Ends_at     int64 `json:"Ends_at"`
}

func (a *MyApp) RegOrderHourlyPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Ads_id    int     `json:"Ads_id"`
		Starts_at int64   `json:"Starts_at"`
		Ends_at   int64   `json:"Ends_at"`
		PositionX float64 `json:"PositionX"`
		PositionY float64 `json:"PositionY"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.RegOrderHourlySQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, booking.Ads_id, time.Unix(booking.Starts_at, 0).UTC(), time.Unix(booking.Ends_at, 0).UTC(), booking.PositionX, booking.PositionY)

			errorr(err)
		}
	}
}

func (a *MyApp) RegOrderDailyPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Ads_id    int     `json:"Ads_id"`
		Starts_at int64   `json:"Starts_at"`
		Ends_at   int64   `json:"Ends_at"`
		PositionX float64 `json:"PositionX"`
		PositionY float64 `json:"PositionY"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.RegOrderDailySQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, booking.Ads_id, time.Unix(booking.Starts_at, 0).UTC(), time.Unix(booking.Ends_at, 0).UTC(), booking.PositionX, booking.PositionY)

			errorr(err)
		}
	}
}

func (a *MyApp) BiddingPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Bidding struct {
		Ads_id      int     `json:"Ads_id"`
		Global_rate int     `json:"Global_rate"`
		Start_at    int64   `json:"Start_at"`
		End_at      int64   `json:"End_at"`
		PositionX   float64 `json:"PositionX"`
		PositionY   float64 `json:"PositionY"`
	}

	var bidding Bidding

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&bidding)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.BiddingSQL(a.app.Ctx, rw, a.app.Repo.Pool, bidding.Ads_id, bidding.Global_rate, user_id, time.Unix(bidding.Start_at, 0).UTC(), time.Unix(bidding.End_at, 0).UTC(), bidding.PositionX, bidding.PositionY)

			errorr(err)
		}
	}
}

func (a *MyApp) RegOrderWithBiddingPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Bidding_id int `json:"Bidding_id"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.RegOrderWithBiddingSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, booking.Bidding_id)

			errorr(err)
		}
	}
}

func (a *MyApp) RebookOrderHourlyPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Order_id  int   `json:"Order_id"`
		Starts_at int64 `json:"Starts_at"`
		Ends_at   int64 `json:"Ends_at"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.RebookOrderHourlySQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, booking.Order_id, time.Unix(booking.Starts_at, 0).UTC(), time.Unix(booking.Ends_at, 0).UTC())

			errorr(err)
		}
	}
}

func (a *MyApp) RebookOrderDailyPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Order_id  int   `json:"Order_id"`
		Starts_at int64 `json:"Starts_at"`
		Ends_at   int64 `json:"Ends_at"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Println("что-то не так с токеном")
	} else {
		flag, user_id := jwt.IsAuthorized(token)

		if flag {
			err = repo.RebookOrderDailySQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, booking.Order_id, time.Unix(booking.Starts_at, 0).UTC(), time.Unix(booking.Ends_at, 0).UTC())

			errorr(err)
		}
	}
}

func (a *MyApp) ComplBookingPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	type Booking struct {
		Bookings_id []int `json:"Bookings_id"`
	}

	var booking Booking

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&booking)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := internal.ReadCookie("token", r)

	if err != nil {
		fmt.Errorf("Ошибка создания объявления: %v", err)
		return
	} else {
		flag, user_id := jwt.IsAuthorized(token)
		if flag {

			err = repo.SucBookingSQL(a.app.Ctx, rw, a.app.Repo.Pool, booking.Bookings_id, user_id)

			errorr(err)
		}
	}
}

type Report struct {
	Order_id int `json:"Order_id"`
}

func (a *MyApp) RegReportPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
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

func (a *MyApp) SendCodeForRecoveryPassWithEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var passwd Passwd

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&passwd)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		if passwd.Passwd_1 == passwd.Passwd_2 {
			flag := validatePassword(rw, passwd.Passwd_1)

			if flag {
				return
			}

			// Генерация JWT токена
			Jwt_for_proof, err := jwt.GenerateJWT("jwt_for_reg", 0)
			errorr(err)

			// Установка куки
			livingTime := 20 * time.Minute //не смог найти чего-то получше
			expiration := time.Now().Add(livingTime)
			cookie := http.Cookie{
				Name:     "token",
				Value:    url.QueryEscape(Jwt_for_proof),
				Expires:  expiration,
				Path:     "/",             // Убедитесь, что путь корректен
				Domain:   "185.112.83.36", // IP-адрес вашего сервера
				HttpOnly: true,
				Secure:   false, // Для HTTP можно оставить false
				SameSite: http.SameSiteLaxMode,
			}

			fmt.Printf("Кука установлена: %v\n", cookie)

			err = repo.RecoveryPassWithEmailSQL(a.app.Ctx, rw, a.app.Repo.Pool, redisClient, user_id, passwd.Passwd_1, Jwt_for_proof)

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
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Поля не совпадают!",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	errorr(err)
}

func (a *MyApp) EnterCodeForRecoveryPassWithEmailPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var email Reg_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	Jwt_for_proof, err := r.Cookie("jwt_for_proof")

	Jwt_for_proof_flag, _ := jwt.IsAuthorized(Jwt_for_proof.Value)

	if token_flag && Jwt_for_proof_flag {
		// Получаем данные из Redis
		keshData, err := redisClient.Get(ctx, Jwt_for_proof.Value).Result()
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
		if email.Reg_code == storedKesh.Code {
			err = repo.EnterCodeForRecoveryPassWithEmailSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, storedKesh.Passwd)
			if err != nil {
				log.Fatal(err)
			}
			rw.Write([]byte("Смена завершена успешно"))
		} else {
			http.Error(rw, "Неверный код подтверждения", http.StatusUnauthorized)
		}
		errorr(err)
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Поля не совпадают!",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
}

func (a *MyApp) SendCodeForRecoveryPassWithPhoneNumPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var passwd Passwd

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&passwd)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	// Попытка прочитать куку
	token, err := r.Cookie("token")
	errorr(err)

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		if passwd.Passwd_1 == passwd.Passwd_2 {
			flag := validatePassword(rw, passwd.Passwd_1)

			if flag {
				return
			}
			// Генерация JWT токена
			Jwt_for_proof, err := jwt.GenerateJWT("jwt_for_reg", 0)
			errorr(err)

			// Установка куки
			livingTime := 20 * time.Minute //не смог найти чего-то получше
			expiration := time.Now().Add(livingTime)
			cookie := http.Cookie{
				Name:     "token",
				Value:    url.QueryEscape(Jwt_for_proof),
				Expires:  expiration,
				Path:     "/",             // Убедитесь, что путь корректен
				Domain:   "185.112.83.36", // IP-адрес вашего сервера
				HttpOnly: true,
				Secure:   false, // Для HTTP можно оставить false
				SameSite: http.SameSiteLaxMode,
			}

			fmt.Printf("Кука установлена: %v\n", cookie)

			err = repo.RecoveryPassWithPhoneNumSQL(a.app.Ctx, rw, a.app.Repo.Pool, redisClient, user_id, passwd.Passwd_1, Jwt_for_proof)

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
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Поля не совпадают!",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	errorr(err)
}

func (a *MyApp) EnterCodeForRecoveryPassWithPhoneNumPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var email Reg_code

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	Jwt_for_proof, err := r.Cookie("jwt_for_proof")

	Jwt_for_proof_flag, _ := jwt.IsAuthorized(Jwt_for_proof.Value)

	if token_flag && Jwt_for_proof_flag {
		// Получаем данные из Redis
		keshData, err := redisClient.Get(ctx, Jwt_for_proof.Value).Result()
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
		if email.Reg_code == storedKesh.Code {
			err = repo.EnterCodeForRecoveryPassWithEmailSQL(a.app.Ctx, rw, a.app.Repo.Pool, user_id, storedKesh.Passwd)
			if err != nil {
				log.Fatal(err)
			}
			rw.Write([]byte("Смена завершена успешно"))
		} else {
			http.Error(rw, "Неверный код подтверждения", http.StatusUnauthorized)
		}
		errorr(err)
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Поля не совпадают!",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
}

func (a *MyApp) AutorizLoginEmailSendPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
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

func (a *MyApp) AutorizLoginEmailEnterPOST(rw http.ResponseWriter, r *http.Request, redisClient *redis.Client, logger zerolog.Logger) {
	var code Reg_code

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

	if storedKesh.Code == code.Reg_code {
		err := repo.LoginSQL(a.app.Ctx, a.app.Repo.Pool, rw, storedKesh.Login.Login, storedKesh.Login.Password, logger)
		errorr(err)
	}
}

func (a *MyApp) AllUserAdsPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var owner_id Owner_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&owner_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)
	err = repo.AllUserAdsSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, owner_id.Owner_id)

	errorr(err)
}

func (a *MyApp) AllAdsOfThisUserPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var owner_id Owner_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&owner_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		err = repo.AllAdsOfThisUserSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id, owner_id.Owner_id)

		errorr(err)
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Что-то не так с токеном",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
}

type WalletHistory struct {
	Typee int `json:"Type"`
}

func (a *MyApp) WalletHistoryPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var walletHist WalletHistory

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&walletHist)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	token, err := r.Cookie("token")

	token_flag, user_id := jwt.IsAuthorized(token.Value)

	if token_flag {
		err = repo.WalletHistorySQL(a.app.Ctx, rw, a.app.Repo.Pool, r, user_id, walletHist.Typee)

		errorr(err)
	} else {
		response := Response{
			Status:  "fatal",
			Message: "Что-то не так с токеном",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
}

func (a *MyApp) GroupReviewNewOnesFirstPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads_id Ads_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.GroupReviewNewOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, ads_id.Ads_id)

	errorr(err)
}

func (a *MyApp) GroupReviewOldOnesFirstPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads_id Ads_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.GroupReviewOldOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, ads_id.Ads_id)

	errorr(err)
}

func (a *MyApp) GroupReviewLowRatOnesFirstPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads_id Ads_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.GroupReviewLowRatOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, ads_id.Ads_id)

	errorr(err)
}

func (a *MyApp) GroupReviewHigRatOnesFirstPOST(rw http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	var ads_id Ads_id

	// Парсинг JSON-запроса
	err := json.NewDecoder(r.Body).Decode(&ads_id)
	errorr(err)

	repo := database.NewRepo(a.app.Ctx, a.app.Repo.Pool)

	err = repo.GroupReviewHigRatOnesFirstSQL(a.app.Ctx, rw, a.app.Repo.Pool, r, ads_id.Ads_id)

	errorr(err)
}

type Token struct {
	Refresh_token int `json:"refresh_token"`
}
