package auth

import (
	"auth/internal/domain/models"
	"auth/internal/services/auth"
	"context"
	"errors"

	authv1 "github.com/dmitriipudovkin/cafe/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverAPI struct {
	authv1.AuthServer
	auth Auth
}

type Token struct {
	AccessToken  string
	RefreshToken string
}

type Auth interface {
	Login(ctx context.Context, login string, password string) (*models.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.Token, error)
	Logout(ctx context.Context, refreshToken string) error
}

func Register(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &serverAPI{
		auth: auth,
	})
}

func (s *serverAPI) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	panic("implement me")
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	panic("implement me")
}

func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req.GetLogin() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid credentials")
	}

	token, err := s.auth.Login(ctx, req.GetLogin(), req.GetPassword())

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid login or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
