package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

func ConvertToString(oldStr *string) string {
	if oldStr != nil {
		return *oldStr // разыменуем указатель и присвоим значение в str
	} else {
		return "" // можно использовать пустую строку или любое другое значение по умолчанию
	}
}

func (repo *MyRepository) ProductListSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, user_id *int, ads_list []int) (err error) {
	type Product struct {
		Ads_path    string //это фотки объявления
		Avatar_path string //это аватарка юзера

		Ads_photo string
		Avatar    string

		Title         string
		Hourly_rate   int
		Description   string
		Duration      string
		Created_at    time.Time
		Favorite_flag bool
		User_name     string
		Rating        float64
		Review_count  int

		Ads_id      int
		Owner_id    int
		Category_id int
	}

	var Favorite_flag_mass []bool
	var Review_count_mass []int

	products := []Product{}
	request, err := rep.Query(
		ctx,
		`
		WITH duration AS (
			SELECT 
				ads.id,
				ads.owner_id,
				ARRAY[
					MAX(bookings.starts_at),
					MAX(bookings.ends_at)
				] AS date_range
			FROM ads.ads, orders.bookings
			WHERE ads.id = ANY($2::INT[]) AND bookings.ads_id = ads.id
			GROUP BY ads.id, ads.owner_id
		)
		SELECT
			COALESCE(t1.File_path::TEXT, '/root/') AS Ads_path,
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			COALESCE((Duration(
				(SELECT d.date_range::date[] FROM duration d WHERE d.id = t2.id)
			))[1], 'Нет точной информации'),
			t2.Created_at,
			Favorite_flag($1, $2::INT[]),
			t4.Avatar_path::TEXT as User_avatar,
			COALESCE(t3.Name::TEXT, t5.name_of_company::TEXT) as User_name,
			t4.Rating,
			Review_count($2::INT[]),
			t2.Id as Ads_id,
			t2.Owner_id,
			t2.Category_id
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1
			ON t2.id = t1.ad_id  -- Соединение на уровне объявления
		LEFT JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5
			ON t5.user_id = t2.owner_id
		LEFT JOIN
			users.users t4
			ON t4.id = t2.owner_id
		WHERE
			t2.status = true
			AND t2.id = ANY($2::INT[]);
		`,
		user_id,
		ads_list)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_path, //кладем сюда множество, длиною в три ӕлемента, с путями фоток
			&p.Title,
			&p.Hourly_rate,
			&p.Description,
			&p.Duration,
			&p.Created_at,
			&Favorite_flag_mass,
			&p.Avatar_path,
			&p.User_name,
			&p.Rating,
			&Review_count_mass,

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

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		products[i].Favorite_flag = Favorite_flag_mass[i]
		products[i].Review_count = Review_count_mass[i]

		for j := 0; j < len(products[i].Ads_path); j++ {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, products[i].Ads_path)
		}

		products[i].Avatar = ServeSpecificMediaBase64(rw, r, products[i].Avatar_path)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

type Review_type struct {
	Review_id          *int    `json:"Review_id"`
	User_id            *int    `json:"User_id"`
	Review_name        *string `json:"Review_name"`
	Review_avatar      *string `json:"Review_avatar"`
	Review_avatar_path *string `json:"Review_avatar_path"`
	Updated_at_comment *int    `json:"Updated_at_comment"`
	Rating             *int    `json:"Rating"`
	Comment            *string `json:"Comment"`
	State              *int    `json:"State"`
}

type ForPrintAds struct {
	Title        string        `json:"Title"`
	Images       []string      `json:"Imags"`
	File_path    []string      `json:"File_path"`
	Updated_at   int           `json:"Updated_at"`
	Description  string        `json:"Description"`
	Position     pgtype.Point  `json:"Position"`
	Reviews      []Review_type `json:"reviews"`
	Review_count int           `json:"Review_count"`
	Hourly_rate  int           `json:"Hourly_rate"`
	Daily_rate   int           `json:"Daily_rate"`
	Ads_id       int           `json:"Ads_id"`
	Owner_id     int           `json:"Owner_id"`
	Owner_name   string        `json:"Owner_name"`
	Avatar       string        `json:"Avatar"`
	Avatar_path  string        `json:"Avatar_path"`
	Owner_rating float32       `json:"Owner_rating"`
	Duration     string        `json:"Duration"`
	Ads_count    int           `json:"Ads_count"`
}

func (repo *MyRepository) PrintAdsSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, ads_id int) (err error) {
	prod := []ForPrintAds{}

	request, err := rep.Query(ctx, `
		WITH CustomerReviews AS (
			SELECT DISTINCT
			reviews.id AS reviews_id,
			users.id AS user_id,
			COALESCE(individual_user.name, company_user.name_of_company) AS name,
			users.avatar_path,
			COALESCE(reviews.updated_at, reviews.created_at) AS updated_at,
			reviews.rating,
			reviews.state,
			COALESCE(reviews.comment, 'Пустой комментарий') AS comment
		FROM
			ads.reviews
		JOIN
			orders.orders ON reviews.order_id = orders.id
		JOIN
			orders.bookings ON bookings.order_id = orders.id
		JOIN
			finance.transactions ON bookings.transaction_id = transactions.id
		JOIN
			finance.wallets ON transactions.wallet_id = wallets.id
		JOIN
			users.users ON wallets.user_id = users.id
		LEFT JOIN
			users.individual_user ON users.id = individual_user.user_id
		LEFT JOIN
			users.company_user ON users.id = company_user.user_id
		WHERE
			bookings.ads_id = $1
			),
			duration AS (
				SELECT 
				ads.id,
				ads.owner_id,
				ARRAY[
					MAX(bookings.starts_at),
					MAX(bookings.ends_at)
					] AS date_range
				FROM ads.ads, orders.bookings
				WHERE bookings.ads_id = ads.id
				GROUP BY ads.id, ads.owner_id
		),
		owner_info AS (
			SELECT 
				ads.owner_id AS owner_id
			FROM ads.ads
			WHERE ads.id = $1
		),
		ads_info AS (
			SELECT COUNT(*) AS ads_count
			FROM ads.ads
			WHERE ads.owner_id = (SELECT owner_id FROM owner_info)
		)
		SELECT
			t2.title,
			COALESCE(
			CASE WHEN t1.file_path IS NULL THEN ARRAY['/home/'] ELSE ARRAY[t1.file_path] END,
			ARRAY['/home/']
		) AS file_path,
		t2.updated_at,
		t2.description,
		t2.position::TEXT,

		ARRAY(SELECT cr.reviews_id FROM CustomerReviews cr LIMIT 2) AS id,
		ARRAY(SELECT cr.user_id FROM CustomerReviews cr LIMIT 2) AS id,
		ARRAY(SELECT cr.name FROM CustomerReviews cr LIMIT 2) AS name,
		ARRAY(SELECT cr.avatar_path FROM CustomerReviews cr LIMIT 2) AS avatar_path,
		ARRAY(SELECT cr.updated_at FROM CustomerReviews cr LIMIT 2) AS updated_at,
		ARRAY(SELECT cr.rating FROM CustomerReviews cr LIMIT 2) AS rating,
		ARRAY(SELECT COALESCE(cr.comment, 'Пустой коммент') FROM CustomerReviews cr LIMIT 2) AS comment,
		ARRAY(SELECT cr.state FROM CustomerReviews cr LIMIT 2) AS state,
		CARDINALITY(ARRAY_AGG(cr.state)) AS Review_count, 

		t2.hourly_rate,
		t2.daily_rate, 
		t2.id AS ads_id,
		t2.owner_id,
		COALESCE(ind_us.Name::TEXT, comp_us.name_of_company::TEXT) AS owner_name,
		t4.avatar_path AS Avatar_path,
		t4.rating,
		COALESCE((Duration(
			(SELECT d.date_range::date[] FROM duration d WHERE d.id = t2.id)
		))[1], 'Нет точной информации') AS Duration,
		ads_info.ads_count
	FROM
		ads.ads t2
	LEFT JOIN
		ads.ad_photos t1 ON t2.id = t1.ad_id
	LEFT JOIN
		users.individual_user ind_us ON ind_us.user_id = t2.owner_id
	LEFT JOIN
		users.company_user comp_us ON comp_us.user_id = t2.owner_id
	INNER JOIN
		users.users t4 ON t2.owner_id = t4.id
	LEFT JOIN ads_info ON TRUE
	LEFT JOIN
		CustomerReviews cr ON TRUE
	WHERE
		t2.status = true AND t2.id = $1
	GROUP BY
		t2.id, t2.title, t2.updated_at, t2.description, t2.position::TEXT, t2.hourly_rate, t2.daily_rate, t2.owner_id, t4.rating, t4.avatar_path, ind_us.Name, comp_us.name_of_company, t1.file_path, ads_info.ads_count
	LIMIT 1;
		`, ads_id)
	errorr(err)

	var Avatar_cost []string
	var Updated_at_int time.Time

	for request.Next() {
		rev := ForPrintAds{}
		var Review_id *[]int
		var User_id *[]int
		var Name_cost *[]string
		var Avatar_path_cost *[]string
		var Updated_at_cost *[]time.Time
		var Rating_cost *[]int
		var Comment_cost *[]string
		var State_cost *[]int

		err := request.Scan(
			&rev.Title,
			&rev.File_path,
			&Updated_at_int,
			&rev.Description,
			&rev.Position,

			&Review_id,
			&User_id,
			&Name_cost,
			&Avatar_path_cost,
			&Updated_at_cost,
			&Rating_cost,
			&Comment_cost,
			&State_cost,

			&rev.Review_count,
			&rev.Hourly_rate,
			&rev.Daily_rate,
			&rev.Ads_id,
			&rev.Owner_id,
			&rev.Owner_name,
			&rev.Avatar_path,
			&rev.Owner_rating,
			&rev.Duration,
			&rev.Ads_count,
		)

		rev.Updated_at = int(Updated_at_int.Unix())

		if Review_id == nil || len(*Review_id) == 0 {
			rev.Review_count = 0
		}

		if err != nil {
			return err
		}

		if len(rev.File_path) > 0 {
			for i := 0; i < len(rev.File_path); i++ {
				rev.Images = append(rev.Images, ServeSpecificMediaBase64(rw, r, rev.File_path[i]))
				rev.File_path[i] = "/home/"
			}
		}
		if len(*Avatar_path_cost) > 0 {
			for i := 0; i < len(*Avatar_path_cost); i++ {
				Avatar_cost = append(Avatar_cost, ServeSpecificMediaBase64(rw, r, (*Avatar_path_cost)[i]))
				(*Avatar_path_cost)[i] = "/home/"
			}
		}

		if len(*Review_id) > 0 && len(*User_id) > 0 && len(*Name_cost) > 0 &&
			len(*Avatar_path_cost) > 0 && len(*Updated_at_cost) > 0 &&
			len(*Rating_cost) > 0 && len(*Comment_cost) > 0 && len(*State_cost) > 0 {

			rev.Reviews = append(rev.Reviews, Review_type{
				Review_id:     &(*Review_id)[0],
				User_id:       &(*User_id)[0],
				Review_name:   &(*Name_cost)[0],
				Review_avatar: &Avatar_cost[0], // Проверьте, что Avatar_cost инициализирован
				Updated_at_comment: func(t time.Time) *int {
					i := int(t.Unix())
					return &i
				}((*Updated_at_cost)[0]),
				Rating:  &(*Rating_cost)[0],
				Comment: &(*Comment_cost)[0],
				State:   &(*State_cost)[0],
			})

			if len(*Review_id) > 1 && len(*User_id) > 1 && len(*Name_cost) > 1 &&
				len(*Avatar_path_cost) > 1 && len(*Updated_at_cost) > 1 &&
				len(*Rating_cost) > 1 && len(*Comment_cost) > 1 && len(*State_cost) > 1 {

				rev.Reviews = append(rev.Reviews, Review_type{
					Review_id:     &(*Review_id)[1],
					User_id:       &(*User_id)[1],
					Review_name:   &(*Name_cost)[1],
					Review_avatar: &Avatar_cost[1], // Проверьте, что Avatar_cost инициализирован
					Updated_at_comment: func(t time.Time) *int {
						i := int(t.Unix())
						return &i
					}((*Updated_at_cost)[1]),
					Rating:  &(*Rating_cost)[1],
					Comment: &(*Comment_cost)[1],
					State:   &(*State_cost)[1],
				})
			}
		}

		rev.Avatar = ServeSpecificMediaBase64(rw, r, rev.Avatar_path)
		rev.Avatar_path = "/home/"

		prod = append(prod, rev)
	}

	type Response struct {
		Status  string      `json:"status"`
		Data    ForPrintAds `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	if len(prod) == 0 || err != nil || request == nil {
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
		Data:    prod[0],
		Message: "Объявление показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SortProductListDailyRateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, category []int, lowNum, higNum int, lowDate, higDate time.Time, position []float64, distance, rating int) (err error) {
	type Product struct {
		Id *int

		Ads_path  *string
		Ads_photo string

		Daily_rate         *int
		Title              *string
		Category_id        *int
		Name               *string
		Surname_or_ind_num *string
		Owner_id           *int
		Rating             *float64

		Avatar_path  *string
		Avatar_photo string
	}
	products := []Product{}

	errorr(err)

	request, err := rep.Query(
		ctx,
		`
		SELECT DISTINCT ON (t2.Id)
			t2.Id,
			t1.File_path::TEXT,
			t2.Daily_rate,
			t2.Title::TEXT,
			t2.Category_id,
			COALESCE(t3.name::TEXT, t5.name_of_company::TEXT) AS Name,
			COALESCE(t3.surname::TEXT, t5.ind_num_taxp::TEXT) AS Surname_or_ind_num,
			t2.Owner_id,
			t4.Rating,
			t4.Avatar_path

		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1 ON t2.id = t1.ad_id
			AND t2.status = true
		LEFT JOIN
			users.individual_user t3 ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5 ON t5.user_id = t2.owner_id
		INNER JOIN
			users.users t4 ON t2.owner_id = t4.id

		WHERE
			(cardinality($1::INT[]) = 0 OR t2.category_id = ANY($1::INT[]))
		-- сортировка по категориям
		
			AND $2 <= t2.Daily_rate AND t2.Daily_rate <= $3
		-- сортировка по цене

			AND EXISTS (
		WITH booking_array AS (
			SELECT 
				array_agg(bookings.starts_at) AS starts_at_list,
				array_agg(bookings.ends_at) AS ends_at_list,
				array_agg(bookings.id) AS bookings_id_list,
				bookings.ads_id AS ads_id_list
			FROM 
				orders.bookings AS bookings
			LEFT JOIN 
				orders.orders AS orders 
			ON 
				bookings.ads_id = t2.id AND orders.id = bookings.order_id
			GROUP BY 
				bookings.ads_id
		)
		SELECT 
			CASE 
				WHEN booking_array.ads_id_list IS NULL -- Проверка, если записей нет
					THEN TRUE
					WHEN ($4 >= ALL(booking_array.starts_at_list) 
					AND $5 >= ALL(booking_array.ends_at_list))
			OR
			($4 <= ALL(booking_array.starts_at_list) 
					AND $5 <= ALL(booking_array.ends_at_list))
				THEN TRUE 
				ELSE FALSE 
			END AS date_flag
		FROM booking_array
			)
		-- сортировка по дате

		AND ST_Distance(
					ST_SetSRID(ST_MakePoint($6::float8, $7::float8), 4326)::geography,
					ST_SetSRID(ST_MakePoint(t2.Position[0], t2.Position[1]), 4326)::geography
				) < $8
		-- сортировка по радиусу
		
		AND COALESCE(t4.rating = $9, TRUE)
		-- сортировка по рейтингу
		ORDER BY
			t2.Id;
		`,

		category,
		lowNum,
		higNum,
		lowDate,
		higDate,
		position[0],
		position[1],
		distance,
		rating,
	)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ads_path,
			&p.Daily_rate,
			&p.Title,
			&p.Category_id,
			&p.Name,
			&p.Surname_or_ind_num,
			&p.Owner_id,
			&p.Rating,
			&p.Avatar_path,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		if products[i].Ads_path != nil {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, ConvertToString(products[i].Ads_path))
		}
		products[i].Ads_path = nil

		if products[i].Avatar_path != nil {
			products[i].Avatar_photo = ServeSpecificMediaBase64(rw, r, ConvertToString(products[i].Avatar_path))
		}
		products[i].Avatar_path = nil
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

func (repo *MyRepository) SortProductListHourlyRateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, category []int, lowNum, higNum int, lowDate, higDate time.Time, position []float64, distance, rating int) (err error) {
	type Product struct {
		Id *int

		Ads_path  *string
		Ads_photo string

		Hourly_rate        *int
		Title              *string
		Category_id        *int
		Name               *string
		Surname_or_ind_num *string
		Owner_id           *int
		Rating             *float32

		Avatar_path  *string
		Avatar_photo string

		Review_count int
		Duration     string
	}
	products := []Product{}

	errorr(err)

	request, err := rep.Query(
		ctx,
		`
		WITH duration AS (
			SELECT 
				ads.id,
				ads.owner_id,
				ARRAY[
					MAX(bookings.starts_at),
					MAX(bookings.ends_at)
				] AS date_range
			FROM ads.ads, orders.bookings
			WHERE bookings.ads_id = ads.id
			GROUP BY ads.id, ads.owner_id
		)
		SELECT DISTINCT ON (t2.Id)
			t2.Id,
			t1.File_path::TEXT,
			t2.Hourly_rate,
			t2.Title::TEXT,
			t2.Category_id,
			COALESCE(t3.name::TEXT, t5.name_of_company::TEXT) AS Name,
			COALESCE(t3.surname::TEXT, t5.ind_num_taxp::TEXT) AS Surname_or_ind_num,
			t2.Owner_id,
			t4.Rating,
			t4.Avatar_path,
			(Review_count(ARRAY[t2.id]::INT[]))[1],
			COALESCE((Duration(
				(SELECT d.date_range::date[] FROM duration d WHERE d.id = t2.id)
			))[1], 'Нет точной информации')
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1 ON t2.id = t1.ad_id
			AND t2.status = true
		LEFT JOIN
			users.individual_user t3 ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5 ON t5.user_id = t2.owner_id
		INNER JOIN
			users.users t4 ON t2.owner_id = t4.id
		WHERE
			(cardinality($1::INT[]) = 0 OR t2.category_id = ANY($1::INT[]))
			AND $2 <= t2.Hourly_rate AND t2.Hourly_rate <= $3
			AND EXISTS (
				WITH booking_array AS (
				SELECT 
					COALESCE(array_agg(bookings.starts_at::date), ARRAY[NULL::date]) AS starts_at_list,
					COALESCE(array_agg(bookings.ends_at::date), ARRAY[NULL::date]) AS ends_at_list,
					COALESCE(array_agg(bookings.id), ARRAY[NULL::int]) AS bookings_id_list,
					COALESCE(bookings.ads_id, 0) AS ads_id_list
				FROM 
					orders.bookings AS bookings
				RIGHT JOIN
					orders.orders AS orders
					ON orders.id = bookings.order_id AND bookings.ads_id = t2.Id
				GROUP BY 
					bookings.ads_id
				)
				SELECT
				CASE
					WHEN booking_array.ads_id_list IS NULL THEN TRUE
					WHEN cardinality(array_remove(booking_array.starts_at_list, NULL)) = 0 THEN TRUE -- Проверка, содержит ли массив только NULL
					WHEN ($4::date >= ALL(booking_array.starts_at_list) 
					AND $5::date >= ALL(booking_array.ends_at_list))
					OR ($4::date <= ALL(booking_array.starts_at_list) 
					AND $5::date <= ALL(booking_array.ends_at_list))
					THEN TRUE 
					ELSE FALSE 
				END AS date_flag
				FROM booking_array
			)
			AND ST_Distance(
					ST_SetSRID(ST_MakePoint($6::float8, $7::float8), 4326)::geography,
					ST_SetSRID(ST_MakePoint(t2.Position[0], t2.Position[1]), 4326)::geography
				) < $8
		-- сортировка по радиусу
		
		AND (COALESCE(t4.rating = $9, TRUE) OR COALESCE($9 = 0, TRUE))
		-- сортировка по рейтингу
		ORDER BY
			t2.Id;
		`,

		category,
		lowNum,
		higNum,
		lowDate,
		higDate,
		position[0],
		position[1],
		distance,
		rating,
	)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ads_path,
			&p.Hourly_rate,
			&p.Title,
			&p.Category_id,
			&p.Name,
			&p.Surname_or_ind_num,
			&p.Owner_id,
			&p.Rating,
			&p.Avatar_path,
			&p.Review_count,
			&p.Duration,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		if products[i].Ads_path != nil {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, ConvertToString(products[i].Ads_path))
		}
		products[i].Ads_path = nil

		if products[i].Avatar_path != nil {
			products[i].Avatar_photo = ServeSpecificMediaBase64(rw, r, ConvertToString(products[i].Avatar_path))
		}
		products[i].Avatar_path = nil
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
	PositionX,
	PositionY float64,
	updated_at time.Time,
	images []string,
	pwd_mass []string,
	pwd string) (err error) {
	request, err := rep.Query(ctx, `
			WITH i AS (
				INSERT INTO Ads.ads (title, description, hourly_rate, daily_rate, owner_id, category_id, position, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, POINT($7, $8), $9) 
				RETURNING id
			),
			p AS (
				INSERT INTO Ads.Ad_photos (ad_id, file_path, removed_at, status)
				SELECT i.id, $10, $11, $12
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
		PositionX,
		PositionY,
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
		flag, file_err := DeleteImageMass(rw, pwd, images)
		if flag {
			response := Response{
				Status:  "fatal",
				Message: "Проблемы с удалением фото: " + file_err.Error(),
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)

			return err
		}

		response := Response{
			Status:  "fatal",
			Message: "Объявление не зарегистирровано",
		}

		// rw.WriteHeader(http.StatusOK)
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

func (repo *MyRepository) EditAdsListSQL(
	ctx context.Context,
	rw http.ResponseWriter,
	rep *pgxpool.Pool,
	ad_id,
	owner_id int,
) (err error) {
	request, err := rep.Query(ctx,
		`
		WITH i AS (
			SELECT id, title, description, hourly_rate, daily_rate, category_id, location, position
			FROM Ads.ads
			WHERE id = $1 AND owner_id = $2
		),
		j AS (
			SELECT ad_id, file_path, uploaded_at
			FROM ads.ad_photos
			WHERE ad_id IN (SELECT id FROM i)-- Используем подзапрос, чтобы получить id ads
		)
		SELECT i.title, i.description, i.hourly_rate, i.daily_rate, i.category_id, i.location, i.position, j.ad_id, j.file_path, j.uploaded_at
		FROM i
		LEFT JOIN j ON j.ad_id = i.id;
		`,

		ad_id,
		owner_id,
	)
	errorr(err)

	type Ads struct {
		Images      []string     `json:"Images"`
		Title       string       `json:"Title"`
		Description string       `json:"Description"`
		Hourly_rate int          `json:"Hourly_rate"`
		Daily_rate  int          `json:"Daily_rate"`
		Category_id int          `json:"Category_id"`
		Location    string       `json:"Location"`
		Point       pgtype.Point `json:"Point"`

		Ads_id      int       `json:"Ads_id"`
		File_path   []string  `json:"File_path"`
		Uploaded_at time.Time `json:"Uploaded_at"`
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
			&p.Point,

			&p.Ads_id,
			&p.File_path,
			&p.Uploaded_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		mass = append(mass, p)
	}

	type Response struct {
		Status  string `json:"status"`
		Data    Ads    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	for i := range len(mass[0].File_path) {
		image, err := DownloadFile(mass[0].File_path[i])
		if err != nil {
			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(Response{
				Status:  "fatal",
				Message: "Фото не найденно",
			})
			return err
		}

		mass[0].Images = append(mass[0].Images, image)
	}

	if err != nil && len(mass) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Объявление не изменено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	response := Response{
		Status: "success",
		Data: Ads{
			Images:      mass[0].Images,
			Title:       mass[0].Title,
			Description: mass[0].Description,
			Hourly_rate: mass[0].Hourly_rate,
			Daily_rate:  mass[0].Daily_rate,
			Category_id: mass[0].Category_id,
			Location:    mass[0].Location,
			Point:       mass[0].Point,
			Ads_id:      mass[0].Ads_id,
			Uploaded_at: mass[0].Uploaded_at,
		},
		Message: "Объявление показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return err
}

func (repo *MyRepository) UpdAdsAddImgSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, file_path string, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
			INSERT INTO ads.ad_photos (ad_id, file_path)
			VALUES ($1, $2)
			RETURNING ad_id;
			`,
		ad_id,
		file_path,
	)

	var req_ad_id int
	for request.Next() {
		err := request.Scan(
			&req_ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && req_ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_ad_id,
			Message: "Фото изменено",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsDelImgSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, img_id int) (err error) {
	request, err := rep.Query(ctx,
		`
			UPDATE Ads.ad_photos
				SET status = false,
				removed_at = NOW()
			WHERE id = $2
			RETURNING id;
			`,
		img_id,
	)

	var req_ad_id int
	for request.Next() {
		err := request.Scan(
			&req_ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_ad_id,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsTitleSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, title string, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
			UPDATE Ads.ads
				SET title = $1,
				updated_at = NOW()
			WHERE id = $2 AND owner_id = $3
			RETURNING title;
			`,
		title,

		ad_id,
		owner_id,
	)

	var req_title string
	for request.Next() {
		err := request.Scan(
			&req_title,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_title,
			Message: "Объявление показано",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsDescriptionSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, description string, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
		UPDATE Ads.ads
			SET description = $1,
			updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
		RETURNING description;
		`,
		description,

		ad_id,
		owner_id,
	)

	var req_description string
	for request.Next() {
		err := request.Scan(
			&req_description,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_description,
			Message: "Объявление показано",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsHourly_rateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, hourly_rate, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
		UPDATE Ads.ads
			SET hourly_rate = $1,
			updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
		RETURNING id, title;
		`,
		hourly_rate,

		ad_id,
		owner_id,
	)

	var req_hourly_rate int
	for request.Next() {
		err := request.Scan(
			&req_hourly_rate,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_hourly_rate,
			Message: "Объявление показано",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsDaily_rateSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, daily_rate, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
		UPDATE Ads.ads
			SET daily_rate = $1,
			updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
		RETURNING daily_rate;
		`,
		daily_rate,

		ad_id,
		owner_id,
	)

	var req_daily_rate string
	for request.Next() {
		err := request.Scan(
			&req_daily_rate,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_daily_rate,
			Message: "Объявление показано",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsCategory_idSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, category_id, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
		UPDATE Ads.ads
			SET category_id = $1,
			updated_at = NOW()
		WHERE id = $2 AND owner_id = $3
		RETURNING id, title;
		`,
		category_id,

		ad_id,
		owner_id,
	)

	var req_category_id int
	for request.Next() {
		err := request.Scan(
			&req_category_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    int    `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_category_id,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsLocationSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, location string, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
			UPDATE Ads.ads
				SET location = $1,
				updated_at = NOW()
			WHERE id = $2 AND owner_id = $3
			RETURNING id, title;
			`,
		location,

		ad_id,
		owner_id,
	)

	var req_location string
	for request.Next() {
		err := request.Scan(
			&req_location,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err != nil && ad_id == 0 {
			response := Response{
				Status:  "fatal",
				Message: "Объявление не изменено",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_location,
			Message: "Объявление показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) UpdAdsPositionSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, position pgtype.Point, ad_id, owner_id int) (err error) {
	request, err := rep.Query(ctx,
		`
			UPDATE Ads.ads
				SET position = $1,
				updated_at = NOW()
			WHERE id = $2 AND owner_id = $3
			RETURNING position;
			`,
		position,

		ad_id,
		owner_id,
	)

	var req_position pgtype.Point
	for request.Next() {
		err := request.Scan(
			&req_position,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		type Response struct {
			Status  string       `json:"status"`
			Data    pgtype.Point `json:"data,omitempty"`
			Message string       `json:"message"`
		}

		if err != nil && req_position.Status == pgtype.Null {
			response := Response{
				Status:  "fatal",
				Message: "Позиция не изменена",
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    req_position,
			Message: "Позиция изменена",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	return
}

func (repo *MyRepository) DelAdsSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, ads_id int, owner_id int) (err error) {
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

func (repo *MyRepository) SortProductListCategoriezSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, category []int) (err error) {
	type Product struct {
		Id          *int
		File_path   *string
		Hourly_rate *int
		Daily_rate  *int
		Title       *string
		Category_id *int
		Name        interface{}
		Owner_id    *int
		Ads_rating  *int
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
			t2.Id,
			MIN(t1.File_path) AS File_path,  -- или используйте другую агрегатную функцию
			MIN(t2.Hourly_rate) AS Hourly_rate,
			MIN(t2.Daily_rate) AS Daily_rate,
			MIN(t2.Title) AS Title,
			MIN(t2.Category_id) AS Category_id,
			COALESCE(MIN(t3.Name), MIN(CAST(comp_user.ind_num_taxp AS text))) AS Name,
			MIN(t2.Owner_id) AS Owner_id,
			MIN(t4.rating) AS rating
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1 ON t2.id = t1.ad_id
		LEFT JOIN
			users.individual_user t3 ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user comp_user ON comp_user.user_id = t2.owner_id
		INNER JOIN
			users.users t4 ON COALESCE(t3.user_id, comp_user.user_id) = t4.id
		WHERE
			t2.status = true
			AND (COALESCE($1, ARRAY[]::int[]) = ARRAY[]::int[]
			OR t2.category_id = ANY($1::int[]))
		GROUP BY t2.id
		ORDER BY t2.id;
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
			&p.Id,
			&p.File_path,
			&p.Hourly_rate,
			&p.Daily_rate,
			&p.Title,
			&p.Category_id,
			&p.Name,
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

func (repo *MyRepository) SigReviewSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, order_id, rating int, comment string, state int) (err error) {
	request, err := rep.Query(
		ctx,
		`
		WITH booking AS (
			SELECT booking_id FROM orders.orders WHERE id = $1
		),
		transact AS (
			SELECT transaction_id FROM orders.bookings WHERE id = (SELECT booking_id FROM booking)
		),
		wallet AS (
			SELECT wallet_id FROM finance.transactions WHERE id = (SELECT transaction_id FROM transact)
		),
		users AS (
			SELECT user_id FROM finance.wallets WHERE id = (SELECT wallet_id FROM wallet)
		)
		INSERT INTO ads.reviews(order_id, rating, comment, state)
		SELECT $1, $2, $3, $5
		WHERE $4 = (SELECT user_id FROM users) AND $5 >= 1 AND $5 <= 4
		RETURNING id;
		`,

		order_id,
		rating,
		comment,
		user_id,
		state)

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

func (repo *MyRepository) UpdReviewSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, reviewer_id, review_id, rating int, comment string) (err error) {
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

func (repo *MyRepository) SigFavAdsSQL(ctx context.Context, rep *pgxpool.Pool, rw http.ResponseWriter, user_id int, ads_id int) (err error) {
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

func (repo *MyRepository) DelFavAdsSQL(ctx context.Context, rep *pgxpool.Pool, rw http.ResponseWriter, user_id, ads_id int) (err error) {
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

func (repo *MyRepository) GroupFavByRecentSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, user_id int) (err error) {
	type Product struct {
		Ads_path    string //это фотки объявления
		Avatar_path string //это аватарка юзера

		Ads_photo string
		Avatar    string

		Title         string
		Hourly_rate   int
		Description   string
		Duration      string
		Created_at    time.Time
		Favorite_flag bool
		User_name     string
		Rating        float64
		Review_count  int

		Ads_id      int
		Owner_id    int
		Category_id int
	}

	var Duration_mass []string
	var Favorite_flag_mass []bool
	var Review_count_mass []int

	products := []Product{}
	request, err := rep.Query(
		ctx,
		`
		WITH duration AS (
			SELECT
				ARRAY_AGG(ads.id) AS ad_ids, -- Собираем все id в массив
				ARRAY_AGG(ARRAY[bookings.starts_at, bookings.ends_at]) AS date_range -- Собираем массив пар дат
			FROM ads.ads
			LEFT JOIN orders.bookings 
				ON bookings.ads_id = ads.id
			LEFT JOIN orders.orders
				ON orders.id = bookings.order_id
			INNER JOIN ads.favorite_ads 
				ON ads.id = favorite_ads.ad_id AND favorite_ads.user_id = $1
		)
		SELECT
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			Duration(
				(SELECT d.date_range::date[] FROM duration d WHERE t2.id = ANY(d.ad_ids))
			) AS duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			t4.Avatar_path::TEXT as User_avatar,
			COALESCE(t3.Name::TEXT, t5.name_of_company::TEXT) as User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			t2.Id as Ads_id,
			t2.Owner_id,
			t2.Category_id
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1
			ON t2.id = t1.ad_id  -- Соединение на уровне объявления
		LEFT JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5
			ON t5.user_id = t2.owner_id
		LEFT JOIN
			users.users t4
			ON t4.id = t2.owner_id
		WHERE
			t2.status = true
			AND t2.id = ANY((SELECT ad_ids FROM duration)::INT[])
		GROUP BY 
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			User_avatar,
			User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			Ads_id,
			t2.Owner_id,
			t2.Category_id
		ORDER BY t2.Created_at desc
		`,

		29)
	errorr(err)
	fmt.Println(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_path, //кладем сюда множество, длиною в три ӕлемента, с путями фоток
			&p.Title,
			&p.Hourly_rate,
			&p.Description,
			&Duration_mass,
			&p.Created_at,
			&Favorite_flag_mass,
			&p.Avatar_path,
			&p.User_name,
			&p.Rating,
			&Review_count_mass,

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

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		// products[i].Duration = Duration_mass[i]
		products[i].Favorite_flag = Favorite_flag_mass[i]
		products[i].Review_count = Review_count_mass[i]

		for j := 0; j < len(products[i].Ads_path); j++ {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, products[i].Ads_path)
		}

		products[i].Avatar = ServeSpecificMediaBase64(rw, r, products[i].Avatar_path)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) GroupFavByCheaperSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, user_id int) (err error) {
	type Product struct {
		Ads_path    string //это фотки объявления
		Avatar_path string //это аватарка юзера

		Ads_photo string
		Avatar    string

		Title         string
		Hourly_rate   int
		Description   string
		Duration      string
		Created_at    time.Time
		Favorite_flag bool
		User_name     string
		Rating        float64
		Review_count  int

		Ads_id      int
		Owner_id    int
		Category_id int
	}

	var Duration_mass []string
	var Favorite_flag_mass []bool
	var Review_count_mass []int

	products := []Product{}
	request, err := rep.Query(
		ctx,
		`
		WITH duration AS (
			SELECT
				ARRAY_AGG(ads.id) AS ad_ids, -- Собираем все id в массив
				ARRAY_AGG(ARRAY[bookings.starts_at, bookings.ends_at]) AS date_range -- Собираем массив пар дат
			FROM ads.ads
			LEFT JOIN orders.bookings 
				ON bookings.ads_id = ads.id
			LEFT JOIN orders.orders
				ON orders.id = bookings.order_id
			INNER JOIN ads.favorite_ads 
				ON ads.id = favorite_ads.ad_id AND favorite_ads.user_id = $1
		)
		SELECT
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			Duration(
				(SELECT d.date_range::date[] FROM duration d WHERE t2.id = ANY(d.ad_ids))
			) AS duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			t4.Avatar_path::TEXT as User_avatar,
			COALESCE(t3.Name::TEXT, t5.name_of_company::TEXT) as User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			t2.Id as Ads_id,
			t2.Owner_id,
			t2.Category_id
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1
			ON t2.id = t1.ad_id  -- Соединение на уровне объявления
		LEFT JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5
			ON t5.user_id = t2.owner_id
		LEFT JOIN
			users.users t4
			ON t4.id = t2.owner_id
		WHERE
			t2.status = true
			AND t2.id = ANY((SELECT ad_ids FROM duration)::INT[])
		GROUP BY 
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			User_avatar,
			User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			Ads_id,
			t2.Owner_id,
			t2.Category_id
		ORDER BY t2.Hourly_rate ASC;
		`,
		29)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_path, //кладем сюда множество, длиною в три ӕлемента, с путями фоток
			&p.Title,
			&p.Hourly_rate,
			&p.Description,
			&Duration_mass,
			&p.Created_at,
			&Favorite_flag_mass,
			&p.Avatar_path,
			&p.User_name,
			&p.Rating,
			&Review_count_mass,

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

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		// products[i].Duration = Duration_mass[i]
		products[i].Favorite_flag = Favorite_flag_mass[i]
		products[i].Review_count = Review_count_mass[i]

		for j := 0; j < len(products[i].Ads_path); j++ {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, products[i].Ads_path)
		}

		products[i].Avatar = ServeSpecificMediaBase64(rw, r, products[i].Avatar_path)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) GroupFavByDearlySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, r *http.Request, user_id int) (err error) {
	type Product struct {
		Ads_path    string //это фотки объявления
		Avatar_path string //это аватарка юзера

		Ads_photo string
		Avatar    string

		Title         string
		Hourly_rate   int
		Description   string
		Duration      string
		Created_at    time.Time
		Favorite_flag bool
		User_name     string
		Rating        float64
		Review_count  int

		Ads_id      int
		Owner_id    int
		Category_id int
	}

	var Duration_mass []string
	var Favorite_flag_mass []bool
	var Review_count_mass []int

	products := []Product{}
	request, err := rep.Query(
		ctx,
		`
		WITH duration AS (
			SELECT
				ARRAY_AGG(ads.id) AS ad_ids, -- Собираем все id в массив
				ARRAY_AGG(ARRAY[bookings.starts_at, bookings.ends_at]) AS date_range -- Собираем массив пар дат
			FROM ads.ads
			LEFT JOIN orders.bookings 
				ON bookings.ads_id = ads.id
			LEFT JOIN orders.orders
				ON orders.id = bookings.order_id
			INNER JOIN ads.favorite_ads 
				ON ads.id = favorite_ads.ad_id AND favorite_ads.user_id = $1
		)
		SELECT
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			Duration(
				(SELECT d.date_range::date[] FROM duration d WHERE t2.id = ANY(d.ad_ids))
			) AS duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			t4.Avatar_path::TEXT as User_avatar,
			COALESCE(t3.Name::TEXT, t5.name_of_company::TEXT) as User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			t2.Id as Ads_id,
			t2.Owner_id,
			t2.Category_id
		FROM
			ads.ads t2
		LEFT JOIN
			ads.ad_photos t1
			ON t2.id = t1.ad_id  -- Соединение на уровне объявления
		LEFT JOIN
			users.individual_user t3
			ON t3.user_id = t2.owner_id
		LEFT JOIN
			users.company_user t5
			ON t5.user_id = t2.owner_id
		LEFT JOIN
			users.users t4
			ON t4.id = t2.owner_id
		WHERE
			t2.status = true
			AND t2.id = ANY((SELECT ad_ids FROM duration)::INT[])
		GROUP BY 
			COALESCE(t1.File_path::TEXT, '/root/'),
			t2.Title::TEXT,
			t2.Hourly_rate,
			t2.Description::TEXT,
			duration_result, -- Функция принимает массив
			t2.Created_at,
			Favorite_flag($1, (SELECT ad_ids FROM duration)::INT[]),
			User_avatar,
			User_name,
			t4.Rating,
			Review_count((SELECT ad_ids FROM duration)::INT[]),
			Ads_id,
			t2.Owner_id,
			t2.Category_id
		ORDER BY t2.Hourly_rate DESC;
		`,
		29)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Ads_path, //кладем сюда множество, длиною в три ӕлемента, с путями фоток
			&p.Title,
			&p.Hourly_rate,
			&p.Description,
			&Duration_mass,
			&p.Created_at,
			&Favorite_flag_mass,
			&p.Avatar_path,
			&p.User_name,
			&p.Rating,
			&Review_count_mass,

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

	for i := 0; i < len(products); i++ { //пока что у нас три объявления
		// products[i].Duration = Duration_mass[i]
		products[i].Favorite_flag = Favorite_flag_mass[i]
		products[i].Review_count = Review_count_mass[i]

		for j := 0; j < len(products[i].Ads_path); j++ {
			products[i].Ads_photo = ServeSpecificMediaBase64(rw, r, products[i].Ads_path)
		}

		products[i].Avatar = ServeSpecificMediaBase64(rw, r, products[i].Avatar_path)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) GroupAdsByRentedSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int) (err error) {
	type Product struct {
		Id          int
		Title       string
		Description string
		Hourly_rate int
		Daily_rate  int
		Owner_id    int
		Category_id int
		Position    pgtype.Point
		Created_at  time.Time
		Updated_at  time.Time
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, position, created_at, updated_at FROM Ads.ads WHERE status = true AND owner_id = $1;",

		user_id,
	)
	errorr(err)

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
			&p.Position,
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
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Сгруппировано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
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
		Position    pgtype.Point
		Created_at  time.Time
		Updated_at  time.Time
		Bool_state  bool
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, title, description, hourly_rate, daily_rate, owner_id, category_id, position, created_at, updated_at, false FROM Ads.ads WHERE status = false AND owner_id = $1;",

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
			&p.Position,
			&p.Created_at,
			&p.Updated_at,
			&p.Bool_state,
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
