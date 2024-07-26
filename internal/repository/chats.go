package repository

import (
	"context"
	"fmt"
	"net/http"
)

func (r *Repository) RegNewChatSQL(ctx context.Context, rw http.ResponseWriter, id_user int, id_buddy int) (err error) {
	type product struct {
		id_user  int
		id_buddy int
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT Chat.add_chat($1, $2);",

		id_user,
		id_buddy,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, renter_id, total_price, created_at")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id_user,
			&p.id_buddy,
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
