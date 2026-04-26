package cmd

import (
	"log"
	"net"

	"github.com/abdultalif/microservices-grocery-delivery/product-service/config"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/handler"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/adapter/repository"
	"github.com/abdultalif/microservices-grocery-delivery/product-service/internal/core/service"
	productPB "github.com/abdultalif/microservices-grocery-delivery/product-service/proto/product"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var startGrpcCmd = &cobra.Command{
	Use:   "start-grpc",
	Short: "Start gRPC server",
	Long:  "Run gRPC server for product-service",
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
	elasticseachInit, err := cfg.InitElasticsearch()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		return
	}

	productRepo := repository.NewProductRepository(db.DB, elasticseachInit)
	categoryRepo := repository.NewCategoryRepository(db.DB)

	productService := service.NewProductService(productRepo, categoryRepo)

	grpcServer := grpc.NewServer()
	productHandler := handler.NewGRPCProductHandler(productService)
	productPB.RegisterProductServiceServer(grpcServer, productHandler)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("✅ gRPC server listening on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
