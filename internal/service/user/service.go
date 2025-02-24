package userService

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
	"unicode/utf8"
)

type UserService struct {
	tokenStorage TokenStorage
	userProvider UserProvider
	userSaver    UserSaver
}

func New(
	tokenStorage TokenStorage,
	userProvider UserProvider,
	userSaver UserSaver,
) *UserService {
	return &UserService{
		tokenStorage: tokenStorage,
		userProvider: userProvider,
		userSaver:    userSaver,
	}
}

type TokenStorage interface {
	Get(token string) (string, error)
	NewSession(uid string, sessionToken string, expiresAt time.Time) error
	DeleteSession(uid string) error
}

type UserProvider interface {
	User(ctx context.Context, login string) (string, string, error)
}

type UserSaver interface {
	NewUser(ctx context.Context, login string, passHash string) error
	EditUser(
		ctx context.Context,
		uid string,
		login string,
		passHash string,
	) error
}

func (u *UserService) Auth(
	ctx context.Context,
	login string,
	password string,
) (string, time.Time, error) {
	uid, passHash, err := u.userProvider.User(ctx, login)
	if err != nil {
		return "", time.Time{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(password))
	if err != nil {
		return "", time.Time{}, err
	}

	sessionToken := uuid.NewString()

	expiresAt := time.Now().Add(time.Hour * 24 * 7)

	if err := u.tokenStorage.NewSession(
		uid,
		sessionToken,
		expiresAt,
	); err != nil {
		return "", time.Time{}, err
	}

	return sessionToken, expiresAt, nil
}

func (u *UserService) NewUser(
	ctx context.Context,
	login string,
	password string,
) error {
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	passHash, err := HashPassword(password)
	if err != nil {
		return errors.New("internal server error")
	}

	if err := u.userSaver.NewUser(ctx, login, passHash); err != nil {
		return errors.New("internal server error")
	}

	return nil
}

func (u *UserService) Logout(ctx context.Context, token string) error {
	if err := u.tokenStorage.DeleteSession(token); err != nil {
		return err
	}

	return nil
}

func (u *UserService) Edit(
	ctx context.Context,
	token, login, password string,
) error {
	uid, err := u.tokenStorage.Get(token)
	if err != nil {
		return err
	}

	passHash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if err := u.userSaver.EditUser(ctx, uid, login, passHash); err != nil {
		return err
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
