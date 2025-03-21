package userService

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	bizErrors "service-propusk-backend/internal/pkg/errors/biz"
	"service-propusk-backend/internal/pkg/logger/sl"
	"unicode/utf8"
)

type UserService struct {
	log          *slog.Logger
	userProvider UserProvider
	userSaver    UserSaver
	authService  AuthService
}

type AuthService interface {
	Claims(token string) (string, bool, error)
}

func New(
	userProvider UserProvider,
	userSaver UserSaver,
	authService AuthService,
) *UserService {
	return &UserService{
		userProvider: userProvider,
		userSaver:    userSaver,
		authService:  authService,
	}
}

type UserProvider interface {
	User(ctx context.Context, login string) (
		string,
		bool,
		string,
		error,
	)
}

type UserSaver interface {
	NewUser(
		ctx context.Context,
		uid string,
		login string,
		passHash string,
		isAdmin bool,
	) error
	EditUser(
		ctx context.Context,
		uid string,
		login string,
		passHash string,
	) error
}

func (u *UserService) NewUser(
	ctx context.Context,
	login string,
	password string,
) error {

	_, _, _, err := u.userProvider.User(ctx, login)
	if err == nil {
		return bizErrors.ErrorUserAlreadyExists
	}

	if utf8.RuneCountInString(password) < 8 {
		return bizErrors.ErrorShortPassword
	}

	passHash, err := HashPassword(password)
	if err != nil {
		u.log.Error("Error generating password hash: ", sl.Err(err))
		return errors.New("internal server error")
	}

	uid := uuid.NewString()

	if err := u.userSaver.NewUser(
		ctx,
		uid,
		login,
		passHash,
		false,
	); err != nil {
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func (u *UserService) Logout(ctx context.Context, token string) error {
	//if err := u.tokenStorage.DeleteSession(token); err != nil {
	//	return err
	//}

	return nil
}

func (u *UserService) Edit(
	ctx context.Context,
	accessToken, newLogin, newPassword string,
) error {
	uid, _, err := u.authService.Claims(accessToken)
	if err != nil {
		return err
	}

	newPassHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := u.userSaver.EditUser(
		ctx,
		uid,
		newLogin,
		newPassHash,
	); err != nil {
		return bizErrors.ErrorInternalServer
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
