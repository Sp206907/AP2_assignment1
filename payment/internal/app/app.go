package app

import (
	"database/sql"

	"payment/internal/repository"
	transporthttp "payment/internal/transport/http"
	"payment/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Run(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := transporthttp.NewPaymentHandler(paymentUC)

	r := gin.Default()
	paymentHandler.RegisterRoutes(r)

	return r.Run(":8081")
}
