package auth

import (
	"context"
	"errors"
	"fmt"
	"gRPCauth/internal/domain/models"
	jwtSrv "gRPCauth/internal/lib/jwt"
	"gRPCauth/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (app models.App, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// New returns a new instance of the Auth service
func New(log *slog.Logger, saver UserSaver, provider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {

	return &Auth{
		usrSaver:    saver,
		usrProvider: provider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int64) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if errors.Is(err, storage.ErrUserNotFound) {
		a.log.Warn("user not found")
		log.Error(err.Error())
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials")
		log.Error(err.Error())
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		a.log.Error("no such app")
		log.Error(err.Error())
		return "", err
	}
	log.Info("successfully logged in")
	token, err := jwtSrv.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create token")
		log.Error(err.Error())
		return "", err
	}
	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string) (int64, error) {
	const op = "auth.registerNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if errors.Is(err, storage.ErrUserExists) {
		return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
	}
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}
	log.Info("register user success")
	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64) (bool, error) {
	const op = "auth.isAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.String("userID", fmt.Sprint(userID)))

	log.Info("checking user")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if errors.Is(err, storage.ErrAppNotFound) {
		return false, ErrInvalidAppId
	}
	if err != nil {
		log.Error(err.Error())
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin", slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
