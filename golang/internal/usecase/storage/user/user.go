package user

import (
	"context"

	"github.com/nikitaSstepanov/templates/golang/internal/entity"
	"github.com/nikitaSstepanov/tools/client/pg"
	rs "github.com/nikitaSstepanov/tools/client/redis"
	e "github.com/nikitaSstepanov/tools/error"
)

type User struct {
	postgres pg.Client
	redis    rs.Client
}

func New(postgres pg.Client, redis rs.Client) *User {
	return &User{
		postgres,
		redis,
	}
}

func (u *User) GetById(ctx context.Context, id uint64) (*entity.User, e.Error) {
	var user entity.User

	err := u.redis.Get(ctx, redisKey(id)).Scan(&user)
	if err != nil && err != rs.Nil {
		return nil, internalErr
	}

	if user.Id != 0 {
		return &user, nil
	}

	query := idQuery(id)

	row := u.postgres.QueryRow(ctx, query)

	if err := user.Scan(row); err != nil {
		if err == pg.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	err = u.redis.Set(ctx, redisKey(id), &user, redisExpires).Err()
	if err != nil {
		return nil, internalErr
	}

	return &user, nil
}

func (u *User) GetByEmail(ctx context.Context, email string) (*entity.User, e.Error) {
	var user entity.User

	query := emailQuery(email)

	row := u.postgres.QueryRow(ctx, query)

	if err := user.Scan(row); err != nil {
		if err == pg.ErrNoRows {
			return nil, notFoundErr
		} else {
			return nil, internalErr
		}
	}

	return &user, nil
}

func (u *User) Create(ctx context.Context, user *entity.User) e.Error {
	query := createQuery(user)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, query)

	err = row.Scan(&user.Id)
	if err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = u.redis.Set(ctx, redisKey(user.Id), user, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (u *User) Update(ctx context.Context, user *entity.User) e.Error {
	query := updateQuery(user)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, query); err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = u.redis.Del(ctx, redisKey(user.Id)).Err()
	if err != nil {
		return internalErr
	}

	user, getErr := u.GetById(ctx, user.Id)
	if getErr != nil {
		return getErr
	}

	err = u.redis.Set(ctx, redisKey(user.Id), user, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (u *User) Verify(ctx context.Context, user *entity.User) e.Error {
	query := verifyQuery(user.Verified, user.Id)

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, query); err != nil {
		return internalErr
	}

	if err := tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = u.redis.Del(ctx, redisKey(user.Id)).Err()
	if err != nil {
		return internalErr
	}

	user, getErr := u.GetById(ctx, user.Id)
	if getErr != nil {
		return getErr
	}

	err = u.redis.Set(ctx, redisKey(user.Id), user, redisExpires).Err()
	if err != nil {
		return internalErr
	}

	return nil
}

func (u *User) Delete(ctx context.Context, user *entity.User) e.Error {
	query := deleteQuery()

	tx, err := u.postgres.Begin(ctx)
	if err != nil {
		return internalErr
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, user.Id)
	if err != nil {
		return internalErr
	}

	if err = tx.Commit(ctx); err != nil {
		return internalErr
	}

	err = u.redis.Del(ctx, redisKey(user.Id)).Err()
	if err != nil {
		return internalErr
	}

	return nil
}
