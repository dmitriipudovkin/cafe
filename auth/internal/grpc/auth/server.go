package auth

import (
	"auth/internal/domain/models"
	"auth/internal/services/auth"
	"context"
	"errors"
	"strings"

	authv1 "github.com/dmitriipudovkin/cafe/protos/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	accessToken, err := s.GetAccessToken(ctx)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	err = s.auth.Logout(ctx, accessToken)

	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &emptypb.Empty{}, nil
}

func (s *serverAPI) RefreshToken(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid refresh token")
	}

	token, err := s.auth.RefreshToken(ctx, req.GetRefreshToken())

	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return nil, status.Error(codes.InvalidArgument, "invalid token")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RefreshResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
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

func (s *serverAPI) GetAccessToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	if !strings.HasPrefix(accessToken, "Bearer ") {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")

	return accessToken, nil
}
