package cmd

import (
	"net"

	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/abdultalif/microservices-grocery-delivery/user-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/message"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/user-service/internal/core/service"
	userPB "github.com/abdultalif/microservices-grocery-delivery/user-service/proto/user"
)

var startGrpcCmd = &cobra.Command{
	Use:   "start-grpc",
	Short: "Start gRPC server",
	Long:  "Run gRPC server for user-service",
	Run: func(cmd *cobra.Command, args []string) {
		startGrpcServer()
	},
}

func init() {
	rootCmd.AddCommand(startGrpcCmd)
}

func startGrpcServer() {
	cfg := config.NewConfig()

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-2] failed to connect postgres: %v", err)
		return
	}

	publisher, err := message.NewUserEventPublisher(*cfg)
	if err != nil {
		log.Fatalf("[RunServer-4] failed to init RabbitMQ publisher: %v", err)
	}
	defer publisher.Close()

	authRepo := repository.NewAuthRepository(db.DB)
	customerRepo := repository.NewCustomerRepository(db.DB)
	tokenRepo := repository.NewVerficationTokenRepository(db.DB)

	jwtService := service.NewJwtService(cfg)
	customerService := service.NewCustomerService(customerRepo, authRepo, cfg, jwtService, tokenRepo, publisher)

	grpcServer := grpc.NewServer()
	userHandler := handler.NewGRPCUserHandler(customerService, jwtService)
	userPB.RegisterUserServiceServer(grpcServer, userHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Info("✅ gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
