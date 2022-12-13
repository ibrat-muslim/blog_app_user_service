package service

import (
	"context"
	"time"

	pb "github.com/ibrat-muslim/blog_app_user_service/genproto/user_service"
	"github.com/ibrat-muslim/blog_app_user_service/storage"
	"github.com/ibrat-muslim/blog_app_user_service/storage/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	storage  storage.StorageI
	inMemory storage.InMemoryStorageI
}

func NewUserService(strg storage.StorageI, inMemory storage.InMemoryStorageI) *UserService {
	return &UserService{
		storage:  strg,
		inMemory: inMemory,
	}
}

func (s *UserService) Create(ctx context.Context, req *pb.User) (*pb.User, error) {
	user, err := s.storage.User().Create(&repo.User{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		PhoneNumber:     req.PhoneNumber,
		Email:           req.Email,
		Gender:          req.Gender,
		Password:        req.Password,
		Username:        req.Username,
		ProfileImageUrl: req.ProfileImageUrl,
		Type:            req.Type,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseUserModel(user), nil
}

func (s *UserService) Get(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := s.storage.User().Get(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseUserModel(user), nil
}

func (s *UserService) GetByEmail(ctx context.Context, req *pb.EmailRequest) (*pb.User, error) {
	user, err := s.storage.User().GetByEmail(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseUserModel(user), nil
}

func (s *UserService) GetAll(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	result, err := s.storage.User().GetAll(&repo.GetUsersParams{
		Limit:  req.Limit,
		Page:   req.Page,
		Search: req.Search,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	response := pb.GetAllUsersResponse{
		Users: make([]*pb.User, 0),
		Count: result.Count,
	}

	for _, user := range result.Users {
		response.Users = append(response.Users, parseUserModel(user))
	}

	return &response, nil
}

func (s *UserService) Update(ctx context.Context, req *pb.User) (*pb.User, error) {
	user, err := s.storage.User().Update(&repo.User{
		ID:              req.Id,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		PhoneNumber:     req.PhoneNumber,
		Gender:          req.Gender,
		Username:        req.Username,
		ProfileImageUrl: req.ProfileImageUrl,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return parseUserModel(user), nil
}

func (s *UserService) Delete(ctx context.Context, req *pb.GetUserRequest) (*emptypb.Empty, error) {
	err := s.storage.User().Delete(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal server error: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func parseUserModel(user *repo.User) *pb.User {
	return &pb.User{
		Id:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		PhoneNumber:     user.PhoneNumber,
		Email:           user.Email,
		Gender:          user.Gender,
		Password:        user.Password,
		Username:        user.Username,
		ProfileImageUrl: user.ProfileImageUrl,
		Type:            user.Type,
		CreatedAt:       user.CreatedAt.Format(time.RFC3339),
	}
}
