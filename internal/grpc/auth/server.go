package auth

import (
	"context"
	"errors"
	"gRPCauth/internal/services/auth"
	"github.com/lovesupergames/authProto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email, password string, appId int64) (token string, err error)
	RegisterNewUser(ctx context.Context, email, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

const emptyValue = 0

type serverApi struct {
	sso.UnimplementedAuthServer
	auth Auth
}

func RegisterServerAPI(s *grpc.Server, auth Auth) {
	sso.RegisterAuthServer(s, &serverApi{auth: auth})
}

func (s *serverApi) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	err := validateLogin(req)
	if err != nil {
		return nil, err
	}

	//TODO: implement login via auth service
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int64(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &sso.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	err := validateRegister(req)
	if err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &sso.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *sso.IsAdminRequest) (*sso.IsAdminResponse, error) {
	err := validateIsAdmin(req)
	if err != nil {
		return nil, err
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &sso.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(req *sso.LoginRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "missing required fields")
	}
	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}
	return nil
}

func validateRegister(req *sso.RegisterRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "missing required fields")
	}
	return nil
}

func validateIsAdmin(req *sso.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}
	return nil
}
