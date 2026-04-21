package repository

import (
	"context"

	"github.com/informalinx/blog/internal/lib"
)

func (queries *Queries) FindUserByEmail(email, secret string) (FindByEmailRow, error) {
	hashed, err := lib.HashEmail(email, secret)
	if err != nil {
		return FindByEmailRow{}, err
	}

	return queries.FindByEmail(context.Background(), hashed)
}
