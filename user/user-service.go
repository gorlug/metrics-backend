package user

import (
	"context"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	connPool *pgxpool.Pool
}

func NewUserService(pool *pgxpool.Pool) (*UserService, error) {
	return &UserService{connPool: pool}, nil
}

func (u *UserService) DoesUserEmailExist(email string) (bool, error) {
	dialect := goqu.Dialect("postgres")
	sql, args, _ := dialect.From("users").
		Prepared(true).
		Select(goqu.COUNT("email")).
		Where(goqu.Ex{"email": email}).
		ToSQL()

	rows, err := u.connPool.Query(context.Background(), sql, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return false, err
		}
		return count == 1, nil
	}
	return false, nil
}
