package authService

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	bizErrors "rip/internal/pkg/errors/biz"
	jwtToken "rip/internal/pkg/jwt"
	"rip/internal/pkg/logger/sl"
	"time"
)

type AuthService struct {
	log            *slog.Logger
	tokenTTL       time.Duration
	userProvider   UserProvider
	secretProvider SecretProvider
}

func New(
	log *slog.Logger,
	tokenTTL time.Duration,
	userProvider UserProvider,
	secretP SecretProvider,
) *AuthService {
	return &AuthService{
		log:            log,
		tokenTTL:       tokenTTL,
		userProvider:   userProvider,
		secretProvider: secretP,
	}
}

type UserProvider interface {
	User(ctx context.Context, login string) (string, bool, string, error)
}

type SecretProvider interface {
	Secret() []byte
}

func (a *AuthService) Auth(
	ctx context.Context,
	login string,
	password string,
) (string, error) {
	uid, isAdmin, passHash, err := a.userProvider.User(ctx, login)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(password))
	if err != nil {
		return "", err
	}

	a.log.Info("создаём токен, исАдмин - ")
	fmt.Println(isAdmin)

	token, err := jwtToken.New(
		uid,
		isAdmin,
		a.tokenTTL,
		a.secretProvider.Secret(),
	)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		return "", err
	}

	return token, nil
}

func (a *AuthService) Claims(
	token string,
) (string, bool, error) {
	uid, isAdmin, err := jwtToken.VerifyToken(token, a.secretProvider.Secret())
	if err != nil {
		return "", false, bizErrors.ErrorAuthToken
	}

	return uid, isAdmin, nil
}
