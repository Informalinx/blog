package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/informalinx/blog/internal/lib"
)

func (queries *Queries) FindUserByEmail(email, secret string) (FindByEmailRow, error) {
	hashed, err := lib.HashEmail(email, secret)
	if err != nil {
		return FindByEmailRow{}, err
	}

	return queries.FindByEmail(context.Background(), hashed)
}

func (queries *Queries) EmailExists(ctx context.Context, emailHash string) (bool, error) {
	_, err := queries.FindByEmail(ctx, emailHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return !errors.Is(err, sql.ErrNoRows), nil
}
