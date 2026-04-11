package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"payment/internal/repository"
	transportgrpc "payment/internal/transport/grpc"
	transporthttp "payment/internal/transport/http"
	"payment/internal/usecase"

	pb "github.com/Sp206907/ap2-generated/payment"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func Run() error {
	godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	grpcPort := os.Getenv("GRPC_PORT")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	// gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(transportgrpc.LoggingInterceptor),
		)
		pb.RegisterPaymentServiceServer(grpcServer, transportgrpc.NewPaymentGRPCServer(paymentUC))
		log.Printf("gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// REST server
	paymentHandler := transporthttp.NewPaymentHandler(paymentUC)
	r := gin.Default()
	paymentHandler.RegisterRoutes(r)
	return r.Run(":8081")
}
