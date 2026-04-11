package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"order/internal/repository"
	transportgrpc "order/internal/transport/grpc"
	transporthttp "order/internal/transport/http"
	"order/internal/usecase"

	pb "github.com/Sp206907/ap2-generated/order"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func Run() error {
	godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	paymentGRPCAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	restPort := os.Getenv("REST_PORT")
	orderGRPCPort := os.Getenv("ORDER_GRPC_PORT")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	paymentClient, err := transportgrpc.NewGRPCPaymentClient(paymentGRPCAddr)
	if err != nil {
		return err
	}

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := transporthttp.NewOrderHandler(orderUC)

	// Order gRPC server (streaming)
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", orderGRPCPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		pb.RegisterOrderServiceServer(grpcServer, transportgrpc.NewOrderGRPCServer(db))
		log.Printf("Order gRPC server listening on :%s", orderGRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	r := gin.Default()
	orderHandler.RegisterRoutes(r)
	return r.Run(":" + restPort)
}
