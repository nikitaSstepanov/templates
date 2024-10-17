package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/templates/golang/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/templates/golang/internal/entity"
	e "github.com/nikitaSstepanov/tools/error"
)

const (
	ok      = http.StatusOK
	created = http.StatusCreated
	badReq  = http.StatusBadRequest
	unauth  = http.StatusUnauthorized

	cookieName = "refreshToken"
	cookieAge  = 259200
	cookiePath = "/"
	cookieHost = "localhost"
)

var (
	badReqErr = e.New("Incorrect data.", e.BadInput)
	unauthErr = e.New("You are unauth.", e.Unauthorize)

	logoutMsg = dto.NewMessage("Logout success.")
)

type AuthUseCase interface {
	Login(ctx context.Context, user *entity.User) (*entity.Tokens, e.Error)
	Logout(ctx context.Context, userId uint64) e.Error
	Refresh(ctx context.Context, refresh string) (*entity.Tokens, e.Error)
}

type Middleware interface {
	CheckAccess(roles ...string) gin.HandlerFunc
}
