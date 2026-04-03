package app

import (
	"database/sql"
	"net/http"
	"time"

	"order/internal/repository"
	transporthttp "order/internal/transport/http"
	"order/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Run(dsn, paymentServiceURL string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}

	orderRepo := repository.NewPostgresOrderRepository(db)
	paymentClient := transporthttp.NewHTTPPaymentClient(httpClient, paymentServiceURL)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := transporthttp.NewOrderHandler(orderUC)

	r := gin.Default()
	orderHandler.RegisterRoutes(r)

	return r.Run(":8082")
}
