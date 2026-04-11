package grpc

import (
	"database/sql"
	"time"

	pb "github.com/Sp206907/ap2-generated/order"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	db *sql.DB
}

func NewOrderGRPCServer(db *sql.DB) *OrderGRPCServer {
	return &OrderGRPCServer{db: db}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderService_SubscribeToOrderUpdatesServer) error {
	lastStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			var status string
			err := s.db.QueryRow(`SELECT status FROM orders WHERE id = $1`, req.OrderId).Scan(&status)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if status != lastStatus {
				lastStatus = status
				err = stream.Send(&pb.OrderStatusUpdate{
					OrderId:   req.OrderId,
					Status:    status,
					UpdatedAt: timestamppb.Now(),
				})
				if err != nil {
					return err
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}
