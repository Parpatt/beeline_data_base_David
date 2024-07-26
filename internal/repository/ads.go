package repository

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Ads struct {
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

func (r *Repository) AddAdsSQL(ctx context.Context, title string, description string, hourly_rate int, daily_rate int, owner_id int, category_id int, location string, created_at time.Time, updated_at time.Time) (err error) {
	_, err = r.pool.Exec(ctx, `WITH i AS (INSERT INTO Ads.ads (title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id) INSERT INTO Ads.Ad_photos (ad_id, file_path, uploaded_at, removed_at, status)  SELECT i.id, $10, $11, $12, $13  FROM i;`,
		title,
		description,
		hourly_rate,
		daily_rate,
		owner_id,
		category_id,
		location,
		created_at,
		updated_at,

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

func (r *Repository) DeleteAdsSQL(ctx context.Context, id int) (err error) {
	_, err = r.pool.Exec(ctx, `UPDATE Ads.ads SET status = false WHERE id = $1;`, id)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) AddFavoriteSQL(ctx context.Context, user_id, ads_id string) (err error) {
	_, err = r.pool.Exec(
		ctx,
		"INSERT INTO Ads.favorite_ads(user_id, ad_id) VALUES($1, $2);",

		user_id,
		ads_id,
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) DelFavoriteSQL(ctx context.Context, user_id, ads_id string) (err error) {
	_, err = r.pool.Exec(
		ctx,
		"DELETE FROM Ads.favorite_ads WHERE user_id = $1 AND ad_id = $2;",

		user_id,
		ads_id,
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) AddReviewsSQL(ctx context.Context, ad_id, reviewer_id, rating int, comment string) (err error) {
	_, err = r.pool.Exec(
		ctx,
		"INSERT INTO Ads.reviews (ad_id, reviewer_id, rating, comment, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6);",

		ad_id,
		reviewer_id,
		rating,
		comment,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) UpdReviewsSQL(ctx context.Context, review_id, rating int, comment string) (err error) {
	_, err = r.pool.Exec(
		ctx,
		"UPDATE Ads.reviews SET rating = $1, comment = $2, updated_at = $3 WHERE id = $4;",

		rating,
		comment,
		time.Now(),
		review_id,
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) GroupByHourlyRateSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		hourly_rate int
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, hourly_rate FROM Ads.ads GROUP BY id ORDER BY hourly_rate DESC;",
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления; почасовая ставка")

	for request.Next() {
		p := product{}
		err := request.Scan(&p.id, &p.hourly_rate)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupByDailyRateSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id         int
		daily_rate int
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, daily_rate FROM Ads.ads GROUP BY id ORDER BY daily_rate DESC;",
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления; суточная ставка")

	for request.Next() {
		p := product{}
		err := request.Scan(&p.id, &p.daily_rate)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupByCategorySQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		title       string
		description string
		hourly_rate int
		daily_rate  int
		owner_id    int
		category_id int
		location    string
		created_at  time.Time
		updated_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at FROM Ads.ads WHERE status = true GROUP BY id, category_id;",
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления,заголовок, описание, почасовая ставка, суточная ставка, хозяим объявления, id категории, локация, дата создания, дата обновления")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.title,
			&p.description,
			&p.hourly_rate,
			&p.daily_rate,
			&p.owner_id,
			&p.category_id,
			&p.location,
			&p.created_at,
			&p.updated_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupByLocatSQL(ctx context.Context, rw http.ResponseWriter, locat string) (err error) {
	type product struct {
		id       int
		location string
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, location FROM Ads.ads WHERE location = $1 GROUP BY id, category_id;",

		locat,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления; локация")

	for request.Next() {
		p := product{}
		err := request.Scan(&p.id, &p.location)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortFavByRecentSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		user_id int
		ad_id   int
		reg_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT * FROM Ads.favorite_ads GROUP BY reg_at, user_id, ad_id ORDER BY reg_at desc;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "user_id, ad_id, reg_id(выводит в хреновом формате, но если запустить через терминал, то всё норм)")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.user_id,
			&p.ad_id,
			&p.reg_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortFavByCheaperSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		ads_id      int
		hourly_rate int
		user_id     int
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"WITH i AS (SELECT Ads.id AS ads_id, ads.hourly_rate, favorite_ads.user_id FROM Ads.ads, Ads.favorite_ads WHERE ads.id = favorite_ads.ad_id) SELECT * FROM i GROUP BY i.ads_id, i.hourly_rate, i.user_id ORDER BY i.hourly_rate ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "ads_id, hourly_rate, user_id")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.ads_id,
			&p.hourly_rate,
			&p.user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortFavByDearlySQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		ads_id      int
		hourly_rate int
		user_id     int
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"WITH i AS (SELECT Ads.id AS ads_id, ads.hourly_rate, favorite_ads.user_id FROM Ads.ads, Ads.favorite_ads WHERE ads.id = favorite_ads.ad_id) SELECT * FROM i GROUP BY i.ads_id, i.hourly_rate, i.user_id ORDER BY i.hourly_rate DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "ads_id, hourly_rate, user_id")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.ads_id,
			&p.hourly_rate,
			&p.user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortReviewNewOnesFirstSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		reviewer_id int
		rating      int
		comment     string
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id ORDER BY id DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, reviewer_id, rating,")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.reviewer_id,
			&p.rating,
			&p.comment,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortReviewOldOnesFirstSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		reviewer_id int
		rating      int
		comment     string
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id ORDER BY id ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, reviewer_id, rating,")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.reviewer_id,
			&p.rating,
			&p.comment,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

// tyt
func (r *Repository) SortReviewLowRatOnesFirstSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		reviewer_id int
		rating      int
		comment     string
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id, rating ORDER BY rating ASC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, reviewer_id, rating,")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.reviewer_id,
			&p.rating,
			&p.comment,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) SortReviewHigRatOnesFirstSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		reviewer_id int
		rating      int
		comment     string
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, reviewer_id, rating, comment FROM Ads.reviews GROUP BY id, rating ORDER BY rating DESC;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, reviewer_id, rating,")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.reviewer_id,
			&p.rating,
			&p.comment,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupAdsByRentedSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		title       string
		description string
		hourly_rate int
		daily_rate  int
		owner_id    int
		category_id int
		location    string
		created_at  time.Time
		updated_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at FROM Ads.ads WHERE status = true;",
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления,заголовок, описание, почасовая ставка, суточная ставка, хозяим объявления, id категории, локация, дата создания, дата обновления")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.title,
			&p.description,
			&p.hourly_rate,
			&p.daily_rate,
			&p.owner_id,
			&p.category_id,
			&p.location,
			&p.created_at,
			&p.updated_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupAdsByArchivedSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		title       string
		description string
		hourly_rate int
		daily_rate  int
		owner_id    int
		category_id int
		location    string
		created_at  time.Time
		updated_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, location, created_at, updated_at FROM Ads.ads WHERE status = false;",
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id объявления,заголовок, описание, почасовая ставка, суточная ставка, хозяим объявления, id категории, локация, дата создания, дата обновления")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.title,
			&p.description,
			&p.hourly_rate,
			&p.daily_rate,
			&p.owner_id,
			&p.category_id,
			&p.location,
			&p.created_at,
			&p.updated_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}
