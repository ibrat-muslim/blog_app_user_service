package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ibrat-muslim/blog_app_user_service/config"
	pbn "github.com/ibrat-muslim/blog_app_user_service/genproto/notification_service"
	pbu "github.com/ibrat-muslim/blog_app_user_service/genproto/user_service"
	grpcPkg "github.com/ibrat-muslim/blog_app_user_service/pkg/grpc_client"
	"github.com/ibrat-muslim/blog_app_user_service/pkg/utils"
	"github.com/ibrat-muslim/blog_app_user_service/storage"
	"github.com/ibrat-muslim/blog_app_user_service/storage/repo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	RegisterCodeKey   = "register_code_"
	ForgotPasswordKey = "forgot_password_code_"
)

type AuthService struct {
	pbu.UnimplementedAuthServiceServer
	storage    storage.StorageI
	inMemory   storage.InMemoryStorageI
	grpcClient grpcPkg.GrpcClientI
	cfg        *config.Config
	logger     *logrus.Logger
}

func NewAuthService(strg storage.StorageI, inMemory storage.InMemoryStorageI, grpcClient grpcPkg.GrpcClientI, cfg *config.Config, logger *logrus.Logger) *AuthService {
	return &AuthService{
		storage:    strg,
		inMemory:   inMemory,
		grpcClient: grpcClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *pbu.RegisterRequest) (*emptypb.Empty, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user := repo.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Type:      repo.UserTypeUser,
		Password:  hashedPassword,
	}

	userData, err := json.Marshal(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal: %v", err)
	}

	err = s.inMemory.Set("user_"+user.Email, string(userData), 10*time.Minute)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set to redis: %v", err)
	}

	go func() {
		err := s.sendVerificationCode(RegisterCodeKey, req.Email)
		if err != nil {
			fmt.Printf("failed to send verification code: %v", err)
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *AuthService) sendVerificationCode(key, email string) error {
	code, err := utils.GenerateRandomCode(6)
	if err != nil {
		return err
	}

	err = s.inMemory.Set(key+email, code, time.Minute)
	if err != nil {
		return err
	}

	_, err = s.grpcClient.NotificationService().SendEmail(context.Background(), &pbn.SendEmailRequest{
		To:      email,
		Type:    "verification_email",
		Subject: "Verification email",
		Body: map[string]string{
			"code": code,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Verify(ctx context.Context, req *pbu.VerifyRequest) (*pbu.AuthResponse, error) {

	userData, err := s.inMemory.Get("user_" + req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	var user repo.User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal: %v", err)
	}

	code, err := s.inMemory.Get(RegisterCodeKey + user.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "code_expired")
	}

	if req.Code != code {
		return nil, status.Error(codes.Internal, "incorrect_code")
	}

	result, err := s.storage.User().Create(&user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	token, _, err := utils.CreateToken(s.cfg, &utils.TokenParams{
		UserID:   result.ID,
		UserType: result.Type,
		Email:    result.Email,
		Duration: time.Hour * 24,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create token: %v", err)
	}

	return &pbu.AuthResponse{
		Id:          result.ID,
		FirstName:   result.FirstName,
		LastName:    result.LastName,
		Email:       result.Email,
		Username:    result.Username,
		Type:        result.Type,
		CreatedAt:   result.CreatedAt.Format(time.RFC3339),
		AccessToken: token,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *pbu.LoginRequest) (*pbu.AuthResponse, error) {
	result, err := s.storage.User().GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.Internal, "wrong email or password: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	err = utils.CheckPassword(req.Password, result.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "wrong email or password: %v", err)
	}

	token, _, err := utils.CreateToken(s.cfg, &utils.TokenParams{
		UserID:   result.ID,
		UserType: result.Type,
		Email:    result.Email,
		Duration: time.Hour * 24,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	return &pbu.AuthResponse{
		Id:          result.ID,
		FirstName:   result.FirstName,
		LastName:    result.LastName,
		Username:    result.Username,
		Email:       result.Email,
		Type:        result.Type,
		CreatedAt:   result.CreatedAt.Format(time.RFC3339),
		AccessToken: token,
	}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *pbu.ForgotPasswordRequest) (*emptypb.Empty, error) {
	go func() {
		err := s.sendVerificationCode(ForgotPasswordKey, req.Email)
		if err != nil {
			fmt.Printf("failed to send verification code: %v", err)
		}
	}()

	return &emptypb.Empty{}, nil
}

func (s *AuthService) VerifyForgotPassword(ctx context.Context, req *pbu.VerifyRequest) (*pbu.AuthResponse, error) {
	code, err := s.inMemory.Get(ForgotPasswordKey + req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "code expired: %v", err)
	}

	if req.Code != code {
		return nil, status.Errorf(codes.Internal, "incorrect code: %v", err)
	}

	result, err := s.storage.User().GetByEmail(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	token, _, err := utils.CreateToken(s.cfg, &utils.TokenParams{
		UserID:   result.ID,
		UserType: result.Type,
		Email:    result.Email,
		Duration: time.Minute * 30,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	return &pbu.AuthResponse{
		Id:          result.ID,
		FirstName:   result.FirstName,
		LastName:    result.LastName,
		Username:    result.Username,
		Email:       result.Email,
		Type:        result.Type,
		CreatedAt:   result.CreatedAt.Format(time.RFC3339),
		AccessToken: token,
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *pbu.VerifyTokenRequest) (*pbu.AuthPayload, error) {
	accessToken := req.AccessToken

	payload, err := utils.VerifyToken(s.cfg, accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	hasPermission, err := s.storage.Permission().CheckPermission(payload.UserType, req.Resource, req.Action)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "internal error: %v", err)
	}

	return &pbu.AuthPayload{
		Id:            payload.ID.String(),
		UserId:        payload.UserID,
		Email:         payload.Email,
		UserType:      payload.UserType,
		IssuedAt:      payload.IssuedAt.Format(time.RFC3339),
		ExpiredAt:     payload.ExpiredAt.Format(time.RFC3339),
		HasPermission: hasPermission,
	}, nil
}
