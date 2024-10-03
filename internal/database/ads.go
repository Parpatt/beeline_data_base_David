package database

import (
	"context"
	"encoding/json"
	"fmt"
	"myproject/internal"
	"net/http"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MyRepository struct {
	app *internal.Repository
}

func NewRepo(Ctx context.Context, dbpool *pgxpool.Pool) *MyRepository {
	return &MyRepository{&internal.Repository{}}
}

func (repo *MyRepository) ProductListSQL(ctx context.Context, rw http.ResponseWriter, ads_list []int, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		`SELECT id FROM orders.orders WHERE ad_id = 144 ORDER BY created_at LIMIT 1;`)
	errorr(err)

	var order_id int
	for request.Next() {
		err := request.Scan(
			&order_id,
		)
		if err != nil {
			fmt.Errorf("Error", err)
			continue
		}
	}

	/*
		SELECT starts_at, ends_at FROM orders.bookings
			WHERE order_id = 22 AND starts_at > '0002-01-01' AND ends_at > '0002-01-01'
				ORDER BY created_at LIMIT 1;
	*/

	type Product struct {
		File_path     []string
		Title         string
		Hourly_rate   float64
		Description   string
		Duration      []string
		Created_at    time.Time
		Favorite_flag []bool
		User_avatar   string
		User_name     string
		Rating        float64
		Review_count  []int

		Ads_id      int
		Owner_id    int
		Category_id int
	}
	products := []Product{}
	request, err = rep.Query(
		ctx,
		`
		SELECT
			t1.File_path,
			t2.Title,
			t2.Hourly_rate,
			t2.Description,
			Duration('{
				2025-09-13 07:56:12, 2025-07-12 07:56:12,
				2025-09-14 07:56:12, 2025-07-12 07:56:12,
				2025-09-15 07:56:12, 2025-07-12 07:56:12
			}'),
			t2.Created_at,
			Favorite_flag(32, $1),
			t4.Avatar_path as User_avatar,
			t3.Name as User_name,
			t4.Rating,
			Review_count($1),
			
			t2.Id as Ads_id,
			t2.Owner_id,
			t2.Category_id
		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2
			ON t2.id = t1.ad_id AND t2.status = true
		INNER JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		INNER JOIN
			users.users t4
			ON t4.id = t2.owner_id
		LIMIT 3;
		`, ads_list)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.File_path,
			&p.Title,
			&p.Hourly_rate,
			&p.Description,
			&p.Duration,
			&p.Created_at,
			&p.Favorite_flag,
			&p.User_avatar,
			&p.User_name,
			&p.Rating,
			&p.Review_count,

			&p.Ads_id,
			&p.Owner_id,
			&p.Category_id,
		)
		if err != nil {
			fmt.Errorf("Error", err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || products == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

type CustomerReviews struct {
	Name       string    `json:"Name"`
	Updated_at time.Time `json:"Updated_at"`
	Rating     int       `json:"Rating"`
	Comment    string    `json:"Comment"`
}

type ForPrintAds struct {
	Title            string            `json:"Title"`
	File_path        []string          `json:"File_path"`
	Updated_at       time.Time         `json:"Updated_at"`
	Description      string            `json:"Description"`
	Location         string            `json:"Location"`
	Position         pgtype.Point      `json:"Position"`
	Customer_reviews []CustomerReviews `json:"Customer_reviews"`
	Review_count     int               `json:"Review_count"`
	Hourly_rate      int               `json:"Hourly_rate"`
	Ads_id           int               `json:"Ads_id"`
	Owner_id         int               `json:"Owner_id"`
	Owner_host_name  string            `json:"Owner_host_name"`
	Rating           float64           `json:"Rating"`
	All_Review_count int               `json:"All_Review_count"`
	Ads_count        int               `json:"Ads_count"` //это
}

func (repo *MyRepository) PrintAdsSQL(ctx context.Context, rw http.ResponseWriter, id int, rep *pgxpool.Pool) (err error) {
	prod := []ForPrintAds{}

	request, err := rep.Query(ctx,
		`
		SELECT
			t2.title,
			t1.file_path,
			t2.updated_at,
			t2.description,
			t2.location,
			t2.position,
			t2.hourly_rate,
			t2.id as ads_id,
			t2.owner_id,
			t3.Name as owner_host_name,
			t4.rating

		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2
			ON t2.id = t1.ad_id AND t2.status = true AND t2.id = $1
		INNER JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		INNER JOIN
			users.users t4 ON t2.owner_id = t4.id;
		`, id)

	for request.Next() {
		p := ForPrintAds{}
		err := request.Scan(
			&p.Title,
			&p.File_path,
			&p.Updated_at,
			&p.Description,
			&p.Location,
			&p.Position,
			&p.Hourly_rate,
			&p.Ads_id,
			&p.Owner_id,
			&p.Owner_host_name,
			&p.Rating)
		if err != nil {
			fmt.Println(err)
			continue
		}

		prod = append(prod, p)
	}

	request, err = rep.Query(ctx,
		`
		SELECT
			t1.name,
			t2.updated_at,
			t2.rating,
			t2.comment
		FROM
			ads.reviews t2
		INNER JOIN
			users.individual_user t1
			ON t2.reviewer_id = t1.user_id
		WHERE t2.ad_id = $1;
		`, id)

	var mass []CustomerReviews

	for request.Next() {
		q := CustomerReviews{}
		err := request.Scan(
			&q.Name,
			&q.Updated_at,
			&q.Rating,
			&q.Comment)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mass = append(mass, CustomerReviews{q.Name, q.Updated_at, q.Rating, q.Comment})
	}

	prod[0].Customer_reviews = mass

	prod[0].Review_count = len(prod[0].Customer_reviews)

	request, err = rep.Query(ctx,
		`
		SELECT id FROM ads.reviews WHERE ad_id = $1;
		`, id)

	var mass_rev_count []int
	for request.Next() {
		var num int

		err := request.Scan(
			&num)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mass_rev_count = append(mass_rev_count, num)
	}
	prod[0].All_Review_count = len(mass_rev_count)

	var mass_ads_count []int
	request, err = rep.Query(ctx,
		`
		SELECT id FROM ads.ads WHERE owner_id = $1;
		`, prod[0].Owner_id)

	for request.Next() {
		var num int

		err := request.Scan(
			&num)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mass_ads_count = append(mass_ads_count, num)
	}
	prod[0].Ads_count = len(mass_ads_count)

	type Response struct {
		Status  string      `json:"status"`
		Data    ForPrintAds `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	if err != nil || len(prod) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Объявление не найдено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    prod[0],
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}
}

func (repo *MyRepository) SortProductListAllSQL(ctx context.Context, rw http.ResponseWriter, category []int, lowNum, higNum int, lowDate, higDate time.Time, location string, rating int, rep *pgxpool.Pool) (err error) {
	type Product struct {
		File_path   []string
		Hourly_rate int
		Title       string
		Category_id int
		Name        string
		Id          int
		Owner_id    int
		Ads_rating  float64
	}
	products := []Product{}

	errorr(err)

	request, err := rep.Query(ctx, ``)

	if location == "" {
		request, err = rep.Query(
			ctx,
			`SELECT * FROM sort_product_list($1, $2, $3, $4, $5, $6, $7);`,

			category,
			lowNum,
			higNum,
			lowDate,
			higDate,
			nil,
			rating,
		) //категория передается как массив
	} else {
		request, err = rep.Query(
			ctx,
			`SELECT * FROM sort_product_list($1, $2, $3, $4, $5, $6, $7);`,

			category,
			lowNum,
			higNum,
			lowDate,
			higDate,
			location,
			rating,
		) //категория передается как массив
	}

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)

		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.File_path,
			&p.Hourly_rate,
			&p.Title,
			&p.Category_id,
			&p.Name,
			&p.Id,
			&p.Owner_id,
			&p.Ads_rating,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err == nil && request != nil && len(products) != 0 {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SignupAdsSQL(
	ctx context.Context,
	rw http.ResponseWriter,
	rep *pgxpool.Pool,
	title,
	description string,
	hourly_rate,
	daily_rate,
	owner_id,
	category_id int,
	location string,
	updated_at time.Time,
	images map[string]string,
	pwd string) (err error) {
	request, err := rep.Query(ctx, `
			WITH i AS (
				INSERT INTO Ads.ads (title, description, hourly_rate, daily_rate, owner_id, category_id, location, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				RETURNING id
			),
			p AS (
				INSERT INTO Ads.Ad_photos (ad_id, file_path, removed_at, status)
				SELECT i.id, $9, $10, $11 
				FROM i
				RETURNING ad_id
			)
			SELECT i.id FROM i;
		`,
		title,
		description,
		hourly_rate,
		daily_rate,
		owner_id,
		category_id,
		location,
		updated_at,

		pwd,
		time.Now(),
		false,
	)
	errorr(err)

	var ad_id int
	for request.Next() {
		err := request.Scan(
			&ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ad_id <= 0 {
		flag, file_err := DeleteFileMass(rw, pwd, images)
		if flag {
			response := Response{
				Status:  "fatal",
				Data:    0,
				Message: "Проблемы с удалением фото: " + file_err.Error(),
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)

			return err
		}

		response := Response{
			Status:  "fatal",
			Data:    0,
			Message: "Объявление не зарегистирровано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    ad_id,
		Message: "Объявление показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) UpdAdsSQL(
	ctx context.Context,
	rw http.ResponseWriter,
	rep *pgxpool.Pool,
	title,
	description string,
	hourly_rate,
	daily_rate,
	owner_id,
	category_id int,
	location,
	photoPwd string,
	ad_id int,
	updated_at time.Time,
	pwd string,
	images map[string]string,
) (err error) {
	request, err := rep.Query(ctx,
		`SELECT title, description, hourly_rate, daily_rate, category_id, location FROM Ads.ads WHERE id = $1;`,

		ad_id,
	)
	errorr(err)

	type Ads struct {
		Id          int    `json:"id"`
		Title       string `json:"Title"`
		Description string `json:"Description"`
		Hourly_rate int    `json:"Hourly_rate"`
		Daily_rate  int    `json:"Daily_rate"`
		Category_id int    `json:"Category_id"`
		Location    string `json:"Location"`
	}
	mass := []Ads{}
	for request.Next() {
		p := Ads{}
		err := request.Scan(
			&p.Title,
			&p.Description,
			&p.Hourly_rate,
			&p.Daily_rate,
			&p.Category_id,
			&p.Location,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mass = append(mass, p)
	}

	if title == "" {
		title = mass[0].Title
	}
	if description == "" {
		description = mass[0].Description
	}
	if hourly_rate == 0 {
		hourly_rate = mass[0].Hourly_rate
	}
	if daily_rate == 0 {
		daily_rate = mass[0].Daily_rate
	}
	if category_id == 0 {
		category_id = mass[0].Category_id
	}
	if location == "" {
		location = mass[0].Location
	}

	fmt.Println(mass[0])

	if len(photoPwd) == 0 {
		request, err := rep.Query(ctx,
			`
			UPDATE Ads.ads
				SET title = $1,
				description = $2,
				hourly_rate = $3,
				daily_rate = $4,
				category_id = $5,
				location = $6,
				updated_at = NOW()
			WHERE id = $7 AND owner_id = $8
			RETURNING id;
		`,
			title,
			description,
			hourly_rate,
			daily_rate,
			category_id,
			location,

			ad_id,
			owner_id,
		)

		var ad_id int
		for request.Next() {
			err := request.Scan(
				&ad_id,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil {
			response := Response{
				Status:  "success",
				Data:    0,
				Message: "Объявление не изменено",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)
			return err
		} else {
			response := Response{
				Status:  "success",
				Data:    ad_id,
				Message: "Объявление показано",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)
			return err
		}
	} else {
		request, err_ads := rep.Query(ctx,
			`
				UPDATE Ads.ads
					SET title = $1,
					description = $2,
					hourly_rate = $3,
					daily_rate = $4,
					category_id = $5,
					location = $6,
					updated_at = NOW()
				WHERE id = $7 AND owner_id = $8
				RETURNING id;
			`,
			title,
			description,
			hourly_rate,
			daily_rate,
			category_id,
			location,

			ad_id,
			owner_id,
		)

		var ad_id int
		for request.Next() {
			err := request.Scan(
				&ad_id,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil {
			return err
		}

		_, err = rep.Exec(
			ctx, "INSERT INTO Ads.Ad_photos (ad_id, file_path, removed_at) VALUES($1, $2, $3);",
			ad_id,
			photoPwd,
			time.Now())
		if err != nil && err_ads != nil {
			flag, file_err := DeleteFileMass(rw, pwd, images)
			if flag {
				response := Response{
					Status:  "fatal",
					Data:    0,
					Message: "Проблемы с удалением фото: " + file_err.Error(),
				}

				rw.WriteHeader(http.StatusOK)
				json.NewEncoder(rw).Encode(response)

				return err
			}

			response := Response{
				Status:  "fatal",
				Data:    0,
				Message: "Объявление показано",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    ad_id,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

	}

	return
}

func (repo *MyRepository) DelAdsSQL(ctx context.Context, ads_id int, owner_id int, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(ctx,
		`UPDATE Ads.ads
		SET status = false
		WHERE id = $1 AND owner_id = $2
		RETURNING id;`, ads_id, owner_id)

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	var ad_id int
	for request.Next() {
		err := request.Scan(
			&ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	if err != nil || ad_id == 0 {
		err = fmt.Errorf("failed to exec data: %w", err)
		response := Response{
			Status:  "fatal",
			Message: "Объявление не удалено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else if ad_id == ads_id {
		response := Response{
			Status:  "success",
			Data:    ad_id,
			Message: "Объявление удалено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	return
}

type Int struct {
	Title  string `json:"Title"`
	Ads_id int    `json:"Ads_id"`
}

func (repo *MyRepository) SearchForTechSQL(ctx context.Context, title string, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	products := []Int{}

	request, err := rep.Query(ctx, `
			SELECT ads.title, ads.id FROM ads.ads WHERE ads.title ILIKE '%' || $1 || '%';
		`,
		title,
	)

	if err != nil {
		fmt.Println(err)
	}

	for request.Next() {
		p := Int{}
		err := request.Scan(
			&p.Title,
			&p.Ads_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string `json:"status"`
		Data    []Int  `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || len(products) <= 0 {
		response := Response{
			Status:  "fatal",
			Message: "Объявление не найдено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}
}

func (repo *MyRepository) SortProductListCategoriezSQL(ctx context.Context, rw http.ResponseWriter, category []int, rep *pgxpool.Pool) (err error) {
	type Product struct {
		File_path   []string
		Hourly_rate int
		Title       string
		Category_id int
		Name        string
		Id          int
		Owner_id    int
		Ads_rating  int
	}
	products := []Product{}

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)

		return
	}

	request, err := rep.Query(
		ctx,
		`
		SELECT
			t1.File_path,
			t2.Hourly_rate,
			t2.Title,
			t2.Category_id,
			t3.Name,
			t2.Id,
			t2.Owner_id,
			t4.rating
		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2 ON t2.id = t1.ad_id AND t2.status = true
		INNER JOIN
			users.individual_user t3 ON t3.user_id = t2.owner_id
		INNER JOIN
			users.users t4 ON t3.user_id = t4.id
		WHERE
			(COALESCE($1, ARRAY[]::int[]) = ARRAY[]::int[]
			OR t2.category_id = ANY($1::int[]));
		`,

		category,
	) //категория передается как массив

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)

		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.File_path,
			&p.Hourly_rate,
			&p.Title,
			&p.Category_id,
			&p.Name,
			&p.Id,
			&p.Owner_id,
			&p.Ads_rating,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err == nil && request != nil {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SigChatSQL(ctx context.Context, rw http.ResponseWriter, id_user int, id_ads int, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		"SELECT ads.owner_id FROM ads.ads WHERE id = $1;",

		id_ads,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	var id_buddy int
	for request.Next() {
		err := request.Scan(
			&id_buddy,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	request, err = rep.Query(
		ctx,
		"SELECT Chat.add_chat($1, $2, $3);",

		id_user,
		id_buddy,
		id_ads,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	var chat_id int
	for request.Next() {
		err := request.Scan(
			&chat_id,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && request != nil || request != nil {
		response := Response{
			Status:  "success",
			Data:    chat_id,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) OpenChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, user_id int) (err error) {
	type Product_user struct {
		Text string
		Date time.Time
	}
	Products_user := []Product_user{}

	type Product_buddy struct {
		Text string
		Date time.Time
	}
	Products_buddy := []Product_buddy{}

	type Product struct {
		Product_userr  []Product_user
		Product_buddyy []Product_buddy
	}

	request_1, err := rep.Query( //это запрос на вывод наших сообщений
		ctx,
		"SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id = $2;",

		id_chat,
		user_id,
	)
	errorr(err)

	var Text string
	var Date time.Time

	for request_1.Next() {
		err := request_1.Scan(
			&Text,
			&Date,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
		Products_user = append(Products_user, Product_user{Text: Text, Date: Date})
	}

	request_2, err := rep.Query( //это запрос на вывод сообщений нашего кента
		ctx,
		"SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id != $2;",

		id_chat,
		user_id,
	)
	errorr(err)

	for request_2.Next() {

		err := request_2.Scan(
			&Text,
			&Date,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
		Products_buddy = append(Products_buddy, Product_buddy{Text: Text, Date: Date})
	}

	type Response struct {
		Status  string  `json:"status"`
		Data    Product `json:"data,omitempty"`
		Message string  `json:"message"`
	}

	if err == nil && (Products_user != nil || Products_buddy != nil) {
		response := Response{
			Status: "success",
			Data: Product{
				Product_userr:  Products_user,
				Product_buddyy: Products_buddy,
			},
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SendMessageSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, id_user int, text, images string) (err error) {
	request, err := rep.Query( //это запрос на вывод наших сообщений
		ctx,
		"INSERT INTO Chat.messages(chat_id, sender_id, text) VALUES ($1, $2, $3) RETURNING id;",

		id_chat,
		id_user,
		text,
	)
	errorr(err)

	var mess_id int

	for request.Next() {
		err := request.Scan(
			&mess_id,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && mess_id != 0 {
		response := Response{
			Status:  "success",
			Data:    mess_id,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SigDisputInChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat int, id_user int) (err error) {
	request, err := rep.Query(
		ctx,
		`
		UPDATE chat.chats SET have_disput = true 
		WHERE ((id = $1 and user_1_id = $2) or (id = $1 and user_2_id = $2)) and have_disput = false
		RETURNING id;`,

		id_chat,
		id_user,
	)
	errorr(err)

	var chat_id int

	for request.Next() {
		err := request.Scan(
			&chat_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && chat_id != 0 {
		response := Response{
			Status:  "success",
			Data:    chat_id,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SigReviewSQL(ctx context.Context, rw http.ResponseWriter, ads_id int, reviewer_id int, rating int, comment string, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		"INSERT INTO Ads.reviews (ad_id, reviewer_id, rating, comment, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;",

		ads_id,
		reviewer_id,
		rating,
		comment,
		time.Now(),
		time.Now(),
	)

	errorr(err)

	var rev_id int
	for request.Next() {
		err := request.Scan(
			&rev_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || rev_id <= 0 {
		response := Response{
			Status:  "fatal",
			Message: "Ошибка",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    rev_id,
			Message: "Отзыв добавлен",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (repo *MyRepository) UpdReviewSQL(ctx context.Context, rw http.ResponseWriter, reviewer_id int, review_id int, rating int, comment string, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		`
			UPDATE ads.reviews 
			SET rating = $1, comment = $2, updated_at = NOW() 
			WHERE reviewer_id = $3 AND id = $4 RETURNING reviews.id;
		`,

		rating,
		comment,
		reviewer_id,
		review_id,
	)

	errorr(err)

	var reviews_id int
	for request.Next() {
		err := request.Scan(
			&reviews_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || reviews_id <= 0 {
		response := Response{
			Status:  "fatal",
			Data:    0,
			Message: "Ошибка",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	response := Response{
		Status:  "success",
		Data:    reviews_id,
		Message: "Отзыв обновлён",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

// func (repo *MyRepository) UpdReviewSQL(ctx context.Context, rw http.ResponseWriter, reviewer_id int, review_id int, rating int, comment string, rep *pgxpool.Pool) (err error) {
// 	request, err := rep.Query(
// 		ctx,
// 		`
// 			UPDATE ads.reviews
// 			SET rating = $1, comment = $2, updated_at = NOW()
// 			WHERE reviewer_id = $3 AND id = $4 RETURNING reviews.id;
// 		`,

// 		rating,
// 		comment,
// 		reviewer_id,
// 		review_id,
// 	)

// 	errorr(err)

// 	var reviews_id int
// 	for request.Next() {
// 		err := request.Scan(
// 			&reviews_id,
// 		)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 	}

// 	type Response struct {
// 		Status  string `json:"status"`
// 		Data    int    `json:"data,omitempty"`
// 		Message string `json:"message"`
// 	}

// 	if err != nil || reviews_id <= 0 {
// 		response := Response{
// 			Status:  "fatal",
// 			Data:    0,
// 			Message: "Ошибка",
// 		}

// 		rw.WriteHeader(http.StatusOK)
// 		json.NewEncoder(rw).Encode(response)
// 		return err
// 	}

// 	response := Response{
// 		Status:  "success",
// 		Data:    reviews_id,
// 		Message: "Отзыв обновлён",
// 	}

// 	rw.WriteHeader(http.StatusOK)
// 	json.NewEncoder(rw).Encode(response)
// 	return
// }

type DisputeChat struct {
	Review_id int    `json:"Review_id"`
	Rating    int    `json:"Rating"`
	Comment   string `json:"Comment"`
}

func (repo *MyRepository) MediatorStartWorkingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, chat_id int, mediator_id int) (err error) {
	// type User_1 struct {
	// 	Text string
	// 	Date time.Time
	// }
	// Userr_1 := []User_1{}

	// type User_2 struct {
	// 	Text string
	// 	Date time.Time
	// }
	// Userr_2 := []User_2{}

	// type Product struct {
	// 	Mediator_id     int
	// 	Id_1            int
	// 	Products_user_1 []User_1
	// 	Id_2            int
	// 	Products_user_2 []User_2
	// }

	// request, err := rep.Query(
	// 	ctx,
	// 	`SELECT user_1_id FROM chat.chats WHERE id = $1;`, chat_id)

	// errorr(err)

	// var User_1_id int
	// for request.Next() {
	// 	err := request.Scan(
	// 		&User_1_id,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// }

	// request, err = rep.Query(
	// 	ctx,
	// 	`SELECT user_2_id FROM chat.chats WHERE id = $1;`, chat_id)

	// errorr(err)

	// var User_2_id int
	// for request.Next() {
	// 	err := request.Scan(
	// 		&User_2_id,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// }

	// request, err = rep.Query(
	// 	ctx,
	// 	`SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id = $2;`, chat_id, User_1_id)

	// errorr(err)

	// for request.Next() {
	// 	p := User_1{}
	// 	err := request.Scan(
	// 		&p.Text,
	// 		&p.Date,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)

	// 		continue
	// 	}
	// 	Userr_1 = append(Userr_1, User_1{Text: p.Text, Date: p.Date})
	// }

	// request, err = rep.Query(
	// 	ctx,
	// 	`SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id = $2;`, chat_id, User_2_id)

	// errorr(err)

	// for request.Next() {
	// 	p := User_2{}
	// 	err := request.Scan(
	// 		&p.Text,
	// 		&p.Date,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)

	// 		continue
	// 	}
	// 	Userr_2 = append(Userr_2, User_2{Text: p.Text, Date: p.Date})
	// }

	request, err := rep.Query( //это запрос на вывод наших сообщений
		ctx,
		`
		UPDATE chat.chats SET mediator_id = $1 
		WHERE id = $2
		RETURNING mediator_id;`,

		mediator_id,
		chat_id,
	)
	errorr(err)

	var Id_mediator int

	for request.Next() {
		err := request.Scan(
			&Id_mediator,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && Id_mediator != 0 {
		response := Response{
			Status:  "success",
			Data:    Id_mediator,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) MediatorEnterInChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, chat_id int, mediator_id int) (err error) {
	type User_1 struct {
		Text string
		Date time.Time
	}
	Userr_1 := []User_1{}

	type User_2 struct {
		Text string
		Date time.Time
	}
	Userr_2 := []User_2{}

	type Product struct {
		Id_1            int
		Products_user_1 []User_1
		Id_2            int
		Products_user_2 []User_2
	}

	request, err := rep.Query(
		ctx,
		`SELECT user_1_id FROM chat.chats WHERE id = $1;`, chat_id)

	errorr(err)

	var User_1_id int
	for request.Next() {
		err := request.Scan(
			&User_1_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	request, err = rep.Query(
		ctx,
		`SELECT user_2_id FROM chat.chats WHERE id = $1;`, chat_id)

	errorr(err)

	var User_2_id int
	for request.Next() {
		err := request.Scan(
			&User_2_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	request, err = rep.Query(
		ctx,
		`SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id = $2;`, chat_id, User_1_id)

	errorr(err)

	for request.Next() {
		p := User_1{}
		err := request.Scan(
			&p.Text,
			&p.Date,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
		Userr_1 = append(Userr_1, User_1{Text: p.Text, Date: p.Date})
	}

	request, err = rep.Query(
		ctx,
		`SELECT text, sent_at FROM chat.messages WHERE chat_id = $1 AND sender_id = $2;`, chat_id, User_2_id)

	errorr(err)

	for request.Next() {
		p := User_2{}
		err := request.Scan(
			&p.Text,
			&p.Date,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
		Userr_2 = append(Userr_2, User_2{Text: p.Text, Date: p.Date})
	}

	type Response struct {
		Status  string  `json:"status"`
		Data    Product `json:"data,omitempty"`
		Message string  `json:"message"`
	}

	if err == nil && (len(Userr_1) != 0 || len(Userr_2) != 0) {
		response := Response{
			Status: "success",
			Data: Product{
				Id_1:            User_1_id,
				Products_user_1: Userr_1,
				Id_2:            User_2_id,
				Products_user_2: Userr_2,
			},
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) MediatorFinishJobInChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, chat_id int) (err error) {
	request, err := rep.Query( //это запрос на вывод наших сообщений
		ctx,
		`
		UPDATE chat.chats SET mediator_id = null, have_disput = false
		WHERE id = $1
		RETURNING id;`,

		chat_id,
	)
	errorr(err)

	var id_chat int

	for request.Next() {
		err := request.Scan(
			&id_chat,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && id_chat != 0 {
		response := Response{
			Status:  "success",
			Data:    id_chat,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SigFavAdsSQL(ctx context.Context, user_id int, ads_id int, rep *pgxpool.Pool, rw http.ResponseWriter) (err error) {
	request, err := rep.Query(
		ctx,
		"INSERT INTO Ads.favorite_ads(user_id, ad_id) VALUES ($1, $2) RETURNING ad_id;",

		user_id,
		ads_id,
	)

	var ad_id int
	for request.Next() {
		err := request.Scan(
			&ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ad_id == 0 {
		err = fmt.Errorf("failed to exec data: %w", err)
		response := Response{
			Status:  "fatal",
			Message: "Объявление не добавлено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    ad_id,
			Message: "Объявление добавлено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	return
}

func (repo *MyRepository) DelFavAdsSQL(ctx context.Context, user_id int, ads_id int, rep *pgxpool.Pool, rw http.ResponseWriter) (err error) {
	request, err := rep.Query(
		ctx,
		"DELETE FROM Ads.favorite_ads WHERE user_id = $1 AND ad_id = $2 RETURNING ad_id;",

		user_id,
		ads_id,
	)

	var ad_id int
	for request.Next() {
		err := request.Scan(
			&ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ad_id == 0 {
		err = fmt.Errorf("failed to exec data: %w", err)
		response := Response{
			Status:  "fatal",
			Message: "Объявление не удалено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    ad_id,
			Message: "Объявление удалено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}

	return
}

func (repo *MyRepository) GroupAdsByHourlyRateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		"SELECT id, hourly_rate FROM Ads.ads GROUP BY id ORDER BY hourly_rate DESC;",
	)

	type Product struct {
		Id          int
		Hourly_rate int
	}
	products := []Product{}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Hourly_rate)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}
	return
}

func (repo *MyRepository) GroupAdsByDailyRateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		"SELECT id, daily_rate FROM Ads.ads GROUP BY id ORDER BY daily_rate DESC;",
	)

	errorr(err)

	type Product struct {
		Id          int
		Hourly_rate int
	}
	products := []Product{}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Hourly_rate)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (repo *MyRepository) GroupFavByRecentSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		User_id int
		Ad_id   int
		Reg_at  time.Time

		Ad_photo_id int
		File_path   []string
		Title       string
		Hourly_rate float32
		Description string
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT * FROM Ads.favorite_ads GROUP BY reg_at, user_id, ad_id ORDER BY reg_at desc;",
	)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.User_id,
			&p.Ad_id,
			&p.Reg_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Обрабатываем ad_photos
		request_ad_photo, err_ad_photo := rep.Query(
			ctx,
			"SELECT id, file_path FROM Ads.ad_photos WHERE ad_id = $1;",
			p.Ad_id,
		)
		errorr(err_ad_photo)

		// Цикл для обработки всех строк
		for request_ad_photo.Next() {
			err = request_ad_photo.Scan(
				&p.Ad_photo_id,
				&p.File_path,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		// Обрабатываем ads
		request_ads, err_ads := rep.Query(
			ctx,
			"SELECT title, hourly_rate, description FROM Ads.ads WHERE id = $1;",
			p.Ad_id,
		)
		errorr(err_ads)

		// Цикл для обработки всех строк
		for request_ads.Next() {
			err = request_ads.Scan(
				&p.Title,
				&p.Hourly_rate,
				&p.Description,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err

	}
	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Сгруппировано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return err
}

func (repo *MyRepository) GroupFavByCheaperSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Ads_id      int
		Hourly_rate int
		User_id     int
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"WITH i AS (SELECT Ads.id AS ads_id, ads.hourly_rate, favorite_ads.user_id FROM Ads.ads, Ads.favorite_ads WHERE ads.id = favorite_ads.ad_id) SELECT * FROM i GROUP BY i.ads_id, i.hourly_rate, i.user_id ORDER BY i.hourly_rate ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_id,
			&p.Hourly_rate,
			&p.User_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		fmt.Println(products)
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (repo *MyRepository) GroupFavByDearlySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Ads_id      int
		Hourly_rate int
		User_id     int
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"WITH i AS (SELECT Ads.id AS ads_id, ads.hourly_rate, favorite_ads.user_id FROM Ads.ads, Ads.favorite_ads WHERE ads.id = favorite_ads.ad_id) SELECT * FROM i GROUP BY i.ads_id, i.hourly_rate, i.user_id ORDER BY i.hourly_rate DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_id,
			&p.Hourly_rate,
			&p.User_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		fmt.Println(products)
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (repo *MyRepository) GroupAdsByRentedSQL(ctx context.Context, rw http.ResponseWriter, user_id int, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Title       string
		Description string
		Hourly_rate int
		Daily_rate  int
		Owner_id    int
		Category_id int
		Location    string
		Created_at  time.Time
		Updated_at  time.Time
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at FROM Ads.ads WHERE status = true AND owner_id = $1;",

		user_id,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Title,
			&p.Description,
			&p.Hourly_rate,
			&p.Daily_rate,
			&p.Owner_id,
			&p.Category_id,
			&p.Location,
			&p.Created_at,
			&p.Updated_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}

func (repo *MyRepository) GroupAdsByArchivedSQL(ctx context.Context, rw http.ResponseWriter, user_id int, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Title       string
		Description string
		Hourly_rate int
		Daily_rate  int
		Owner_id    int
		Category_id int
		Location    string
		Created_at  time.Time
		Updated_at  time.Time
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at FROM Ads.ads WHERE status = false AND owner_id = $1;",

		user_id,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Title,
			&p.Description,
			&p.Hourly_rate,
			&p.Daily_rate,
			&p.Owner_id,
			&p.Category_id,
			&p.Location,
			&p.Created_at,
			&p.Updated_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	} else {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Сгруппировано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	}
}
