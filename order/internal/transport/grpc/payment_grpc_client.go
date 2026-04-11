package grpc

import (
	"context"
	"order/internal/usecase"

	pb "github.com/Sp206907/ap2-generated/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPaymentClient struct {
	client pb.PaymentServiceClient
}

func NewGRPCPaymentClient(addr string) (usecase.PaymentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCPaymentClient{client: pb.NewPaymentServiceClient(conn)}, nil
}

func (c *GRPCPaymentClient) Authorize(req usecase.PaymentRequest) (*usecase.PaymentResponse, error) {
	resp, err := c.client.ProcessPayment(context.Background(), &pb.PaymentRequest{
		OrderId: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &usecase.PaymentResponse{
		Status:        resp.Status,
		TransactionID: resp.TransactionId,
	}, nil
}
