package main

import (
	"fmt"
	"log"
	"net"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/ibrat-muslim/blog_app_user_service/genproto/user_service"
	grpcPkg "github.com/ibrat-muslim/blog_app_user_service/pkg/grpc_client"
	"github.com/ibrat-muslim/blog_app_user_service/pkg/logger"

	"github.com/ibrat-muslim/blog_app_user_service/config"
	"github.com/ibrat-muslim/blog_app_user_service/service"
	"github.com/ibrat-muslim/blog_app_user_service/storage"
)

func main() {
	cfg := config.Load(".")

	psqlUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	psqlConn, err := sqlx.Connect("postgres", psqlUrl)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})

	strg := storage.NewStoragePg(psqlConn)
	inMemory := storage.NewInMemoryStorage(rdb)

	grpcConn, err := grpcPkg.New(&cfg)
	if err != nil {
		log.Fatalf("failed to get grpc connections: %v", err)
	}

	logger := logger.New()

	userService := service.NewUserService(strg, inMemory, logger)
	authService := service.NewAuthService(strg, inMemory, grpcConn, &cfg, logger)

	lis, err := net.Listen("tcp", cfg.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	pb.RegisterUserServiceServer(s, userService)
	pb.RegisterAuthServiceServer(s, authService)

	log.Println("Grpc server started in port", cfg.GrpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error while listening: %v", err)
	}
}
