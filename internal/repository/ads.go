package repository

import (
	"context"
	"fmt"
	"time"
)

type Add struct {
	Id          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Hourly_rate int       `json:"hourly_rate" db:"hourly_rate"`
	Daily_rate  int       `json:"daily_rate" db:"daily_rate"`
	Owner_id    int       `json:"owner_id" db:"owner_id"`
	Category_id int       `json:"category_id" db:"category_id"`
	Location    string    `json:"location" db:"location"`
	Created_at  time.Time `json:"created_at" db:"created_at"`
	Updated_at  time.Time `json:"updated_at" db:"updated_at"`
	Status      int       `json:"status" db:"status"`
}

func (r *Repository) AddNewAds(ctx context.Context, title string, description string, hourly_rate int, daily_rate int, owner_id int, category_id int, location string, created_at time.Time, updated_at time.Time, status int) (err error) {
	_, err = r.pool.Exec(ctx, `WITH i AS (INSERT INTO Ads.ads (title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id) INSERT INTO Ads.Ad_photos (ad_id, file_path, uploaded_at, removed_at, status)  SELECT i.id, $11, $12, $13, $14  FROM i;`,
		title,
		description,
		hourly_rate,
		daily_rate,
		owner_id,
		category_id,
		location,
		created_at,
		updated_at,
		status,

		"C:/",
		time.Now(),
		time.Now(),
		false,
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}
