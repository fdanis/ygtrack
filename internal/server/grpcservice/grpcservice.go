package grpcservice

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/fdanis/ygtrack/internal/server/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "net/http/pprof"

	ms "github.com/fdanis/ygtrack/internal/server/metricsservice"
	pb "github.com/fdanis/ygtrack/proto"

	_ "github.com/golang-migrate/migrate/source/file"
)

type GrpcMetricServer struct {
	pb.UnimplementedMetricServiceServer
	service *ms.MetricsService
}

func NewGrpcMetricServer(service *ms.MetricsService) *GrpcMetricServer {
	return &GrpcMetricServer{service: service}
}

func (s *GrpcMetricServer) SendList(stream pb.MetricService_SendListServer) error {
	for {
		model, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Response{})
		}
		if err != nil {
			return err
		}
		item := models.Metrics{ID: model.Id, MType: strings.ToLower(model.MType.String()), Delta: &model.Delta, Value: &model.Value, Hash: model.Hash}
		err = s.service.AddMetric(item)
		if err != nil {
			var merr *ms.MetricsError
			if errors.As(err, &merr) {
				return status.Errorf(codes.InvalidArgument, merr.Error())
			}
			return status.Errorf(codes.Internal, "server error")
		}
		log.Printf("type = %s metric was added %v", model.MType.String(), model)

	}
}

func (s *GrpcMetricServer) Send(ctx context.Context, model *pb.Metrics) (*pb.Response, error) {
	err := s.service.AddMetric(models.Metrics{ID: model.Id, MType: model.MType.String(), Delta: &model.Delta, Value: &model.Value, Hash: model.Hash})
	if err != nil {
		var merr *ms.MetricsError
		if errors.As(err, &merr) {
			return nil, status.Errorf(codes.InvalidArgument, merr.Error())
		}
		return nil, status.Errorf(codes.Internal, "server error")
	}
	return &pb.Response{}, nil
}
