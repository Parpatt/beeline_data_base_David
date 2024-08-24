package database

import (
	"context"
	"encoding/json"
	"fmt"
	"myproject/internal"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type MyRepository struct {
	app *internal.Repository
}

func NewRepo(Ctx context.Context, dbpool *pgxpool.Pool) *MyRepository {
	return &MyRepository{&internal.Repository{}}
}

func (repo *MyRepository) ProductListSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		File_path   string
		Hourly_rate int
		Title       string
		Category_id int
		Name        string
		Id          int
		Owner_id    int
	}
	products := []Product{}

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
			t2.Owner_id
		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2
			ON t2.id = t1.ad_id AND t2.status = true
		INNER JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id;
		`)
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

	if err != nil {
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
	return
}

type ForPrintAds struct {
	Title           string `json:"Title"`
	File_path       string `json:"File_path"`
	Description     string `json:"Description"`
	Location        string `json:"Location"`
	Renter_name     string `json:"Renter_name"`
	Rating          int    `json:"Rating"`
	Comment         string `json:"Comment"`
	Hourly_rate     int    `json:"Hourly_rate"`
	Ads_id          int    `json:"Ads_id"`
	Owner_id        int    `json:"Owner_id"`
	Owner_host_name string `json:"Owner_host_name"`
}

func (repo *MyRepository) PrintAdsSQL(ctx context.Context, rw http.ResponseWriter, id int, rep *pgxpool.Pool) (err error) {
	prod := []ForPrintAds{}

	request, err := rep.Query(ctx,
		`
		SELECT
			t2.title,
			t1.file_path,
			t2.description,
			t2.location,
			t4.Name as renter_name,
			t3.Rating,
			t3.Comment,
			t2.hourly_rate,
			t2.id as ads_id,
			t2.owner_id,
			t5.Name as owner_host_name
		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2
			ON t2.id = t1.ad_id AND t2.status = true AND t2.id = $1
		INNER JOIN
			ads.reviews t3
			ON true
		INNER JOIN
			users.individual_user t4
			ON t4.user_id = t3.reviewer_id
		INNER JOIN
			users.individual_user t5
			ON t5.user_id = t2.owner_id;
		`, id)

	for request.Next() {
		p := ForPrintAds{}
		err := request.Scan(
			&p.Title,
			&p.File_path,
			&p.Description,
			&p.Location,
			&p.Renter_name,
			&p.Rating,
			&p.Comment,
			&p.Hourly_rate,
			&p.Ads_id,
			&p.Owner_id,
			&p.Owner_host_name)
		if err != nil {
			fmt.Println(err)
			continue
		}
		prod = append(prod, p)
	}

	type Response struct {
		Status  string      `json:"status"`
		Data    ForPrintAds `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	if err != nil || len(prod) == 0 {
		response := Response{
			Status: "fatal",
			Data: ForPrintAds{
				Title:           "",
				File_path:       "",
				Description:     "",
				Location:        "",
				Renter_name:     "",
				Rating:          0,
				Comment:         "",
				Hourly_rate:     0,
				Ads_id:          0,
				Owner_id:        0,
				Owner_host_name: "",
			},
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

func (repo *MyRepository) SortProductListSQL(ctx context.Context, rw http.ResponseWriter, category int, rep *pgxpool.Pool) (err error) {
	type Product struct {
		File_path   string
		Hourly_rate int
		Title       string
		Category_id int
		Name        string
		Id          int
		Owner_id    int
	}
	products := []Product{}

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
			t2.Owner_id
		FROM
			ads.ad_photos t1
		INNER JOIN
			ads.ads t2
			ON t2.id = t1.ad_id AND t2.status = true AND category_id = $1
		INNER JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id;
		`,

		category)
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

func (repo *MyRepository) SignupAdsSQL(ctx context.Context, title string, description string, hourly_rate int, daily_rate int, owner_id int, category_id int, location string, updated_at time.Time, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(ctx, `
			WITH i AS (
				INSERT INTO Ads.ads (title, description, hourly_rate, daily_rate, owner_id, category_id, location, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
				RETURNING id
			)
			INSERT INTO Ads.Ad_photos (ad_id, file_path, removed_at, status)
			SELECT i.id, $9, $10, $11 
			FROM i
			RETURNING id;
		`,
		title,
		description,
		hourly_rate,
		daily_rate,
		owner_id,
		category_id,
		location,
		updated_at,

		"C:/",
		time.Now(),
		false,
	)

	if err != nil {
		fmt.Println(err)
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

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || ad_id <= 0 {
		response := Response{
			Status:  "fatal",
			Data:    0,
			Message: "Объявление не зарегистирровано",
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
}

type Int struct {
	Ads_id int `json:"Ads_id"`
}

func (repo *MyRepository) SearchForTechSQL(ctx context.Context, title string, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	products := []Int{}

	request, err := rep.Query(ctx, `
			SELECT ads.id FROM ads.ads WHERE ads.title ILIKE '%' || $1 || '%';
		`,
		title,
	)

	if err != nil {
		fmt.Println(err)
	}

	for request.Next() {
		p := Int{}
		err := request.Scan(
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
	fmt.Println(products)
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
