package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

func (repo *MyRepository) GroupReviewNewOnesFirstSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Reviewer_id int
		Rating      int
		Comment     string
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id ORDER BY id DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Reviewer_id,
			&p.Rating,
			&p.Comment,
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

func (repo *MyRepository) GroupReviewOldOnesFirstSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Reviewer_id int
		Rating      int
		Comment     string
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id ORDER BY id ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Reviewer_id,
			&p.Rating,
			&p.Comment,
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

func (repo *MyRepository) GroupReviewLowRatOnesFirstSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Reviewer_id int
		Rating      int
		Comment     string
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id, rating ORDER BY rating ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Reviewer_id,
			&p.Rating,
			&p.Comment,
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

func (repo *MyRepository) GroupReviewHigRatOnesFirstSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Reviewer_id int
		Rating      int
		Comment     string
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id, rating ORDER BY rating DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Reviewer_id,
			&p.Rating,
			&p.Comment,
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
