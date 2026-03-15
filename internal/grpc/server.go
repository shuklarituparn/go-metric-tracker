// Package grpc provides gRPC server implementation for metrics service.
package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	models "github.com/shuklarituparn/go-metric-tracker/internal/model"
	pb "github.com/shuklarituparn/go-metric-tracker/internal/proto"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
)

// MetricsServer implements the gRPC MetricsServer interface.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	storage repository.Storage
	logger  *zap.Logger
}

// NewMetricsServer creates a new gRPC metrics server.
func NewMetricsServer(storage repository.Storage, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		storage: storage,
		logger:  logger,
	}
}

// UpdateMetrics implements the MetricsServer.UpdateMetrics method.
// It accepts a batch of metrics and stores them in the storage.
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	for _, metric := range req.Metrics {
		if metric == nil {
			continue
		}

		switch metric.Type {
		case pb.Metric_GAUGE:
			if err := s.storage.UpdateGauge(metric.Id, metric.Value); err != nil {
				s.logger.Error("failed to update gauge metric",
					zap.String("id", metric.Id),
					zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to update gauge metric %s: %v", metric.Id, err)
			}
			s.logger.Debug("updated gauge metric",
				zap.String("id", metric.Id),
				zap.Float64("value", metric.Value))

		case pb.Metric_COUNTER:
			if err := s.storage.UpdateCounter(metric.Id, metric.Delta); err != nil {
				s.logger.Error("failed to update counter metric",
					zap.String("id", metric.Id),
					zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to update counter metric %s: %v", metric.Id, err)
			}
			s.logger.Debug("updated counter metric",
				zap.String("id", metric.Id),
				zap.Int64("delta", metric.Delta))

		default:
			s.logger.Warn("unknown metric type",
				zap.String("id", metric.Id),
				zap.Int32("type", int32(metric.Type)))
		}
	}

	s.logger.Info("processed metrics batch",
		zap.Int("count", len(req.Metrics)))

	return &pb.UpdateMetricsResponse{}, nil
}

// TrustedSubnetInterceptor creates a unary interceptor that validates client IP
// against a trusted subnet. The client must pass its IP address in the
// "x-real-ip" metadata key.
func TrustedSubnetInterceptor(trustedSubnet string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	// If trusted subnet is empty, allow all requests
	if trustedSubnet == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Error("invalid trusted subnet CIDR, all requests will be rejected",
			zap.String("subnet", trustedSubnet),
			zap.Error(err))
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, status.Error(codes.PermissionDenied, "server misconfigured: invalid trusted subnet")
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("request rejected: no metadata in context")
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		realIPs := md.Get("x-real-ip")
		if len(realIPs) == 0 {
			logger.Warn("request rejected: missing x-real-ip metadata")
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		realIP := realIPs[0]
		ip := net.ParseIP(realIP)
		if ip == nil {
			logger.Warn("request rejected: invalid IP address in x-real-ip metadata",
				zap.String("ip", realIP))
			return nil, status.Error(codes.PermissionDenied, "invalid IP address")
		}

		if !subnet.Contains(ip) {
			logger.Warn("request rejected: IP not in trusted subnet",
				zap.String("ip", realIP),
				zap.String("subnet", trustedSubnet))
			return nil, status.Error(codes.PermissionDenied, "IP not in trusted subnet")
		}

		logger.Debug("request allowed from trusted IP",
			zap.String("ip", realIP))

		return handler(ctx, req)
	}
}

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
}

func NewServer(storage repository.Storage, trustedSubnet string, logger *zap.Logger) (*Server, error) {
	var opts []grpc.ServerOption

	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	grpcServer := grpc.NewServer(opts...)
	metricsServer := NewMetricsServer(storage, logger)
	pb.RegisterMetricsServer(grpcServer, metricsServer)

	return &Server{
		grpcServer: grpcServer,
		logger:     logger,
	}, nil
}

func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	s.listener = listener

	s.logger.Info("starting gRPC server",
		zap.String("address", address))

	return s.grpcServer.Serve(listener)
}

func (s *Server) Stop() {
	s.logger.Info("stopping gRPC server")
	s.grpcServer.GracefulStop()
}

func ConvertModelToProto(m models.Metrics) *pb.Metric {
	metric := &pb.Metric{
		Id: m.ID,
	}

	switch m.MType {
	case models.Gauge:
		metric.Type = pb.Metric_GAUGE
		if m.Value != nil {
			metric.Value = *m.Value
		}
	case models.Counter:
		metric.Type = pb.Metric_COUNTER
		if m.Delta != nil {
			metric.Delta = *m.Delta
		}
	}

	return metric
}

func ConvertProtoToModel(m *pb.Metric) models.Metrics {
	metric := models.Metrics{
		ID: m.Id,
	}

	switch m.Type {
	case pb.Metric_GAUGE:
		metric.MType = models.Gauge
		value := m.Value
		metric.Value = &value
	case pb.Metric_COUNTER:
		metric.MType = models.Counter
		delta := m.Delta
		metric.Delta = &delta
	}

	return metric
}