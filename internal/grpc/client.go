// Package grpc provides gRPC client implementation for metrics service.
package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/shuklarituparn/go-metric-tracker/internal/proto"
	"github.com/shuklarituparn/go-metric-tracker/internal/repository"
	"github.com/shuklarituparn/go-metric-tracker/internal/retry"
)

type GRPCServerSender interface {
	Start()
	Stop()
	SendMetricsBatch() error
	Close() error
}

type GRPCSender struct {
	conn           *grpc.ClientConn
	client         pb.MetricsClient
	serverAddress  string
	storage        repository.Storage
	ticker         *time.Ticker
	retryConfig    retry.RetryConfig
	reportInterval time.Duration
	localIP        string
}

func NewGRPCSender(serverAddress string, reportInterval time.Duration, storage repository.Storage) (*GRPCSender, error) {
	sender := &GRPCSender{
		serverAddress:  serverAddress,
		reportInterval: reportInterval,
		storage:        storage,
		retryConfig:    retry.DefaultConfig(),
		localIP:        getLocalIP(),
	}

	if err := sender.connect(); err != nil {
		return nil, err
	}

	return sender, nil
}

func (s *GRPCSender) connect() error {
	conn, err := grpc.NewClient(
		s.serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at %s: %w", s.serverAddress, err)
	}

	s.conn = conn
	s.client = pb.NewMetricsClient(conn)
	log.Printf("Connected to gRPC server at %s", s.serverAddress)

	return nil
}

func (s *GRPCSender) Start() {
	s.ticker = time.NewTicker(s.reportInterval)
	go func() {
		for range s.ticker.C {
			if err := s.SendMetricsBatch(); err != nil {
				log.Printf("Error sending metrics batch via gRPC: %v", err)
			}
		}
	}()
}

func (s *GRPCSender) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		log.Println("gRPC sender stopped")
	}
}

func (s *GRPCSender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *GRPCSender) SendMetricsBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := retry.ExecuteWithRetry(ctx, s.retryConfig, func() (struct{}, error) {
		return struct{}{}, s.sendMetricsBatchInternal(ctx)
	})
	return err
}

func (s *GRPCSender) sendMetricsBatchInternal(ctx context.Context) error {
	metrics := s.storage.GetAllMetrics()

	if len(metrics) == 0 {
		log.Println("No metrics to send via gRPC")
		return nil
	}

	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		pbMetrics = append(pbMetrics, ConvertModelToProto(m))
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	if s.localIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", s.localIP)
	}

	_, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics via gRPC: %w", err)
	}

	log.Printf("Successfully sent %d metrics via gRPC", len(metrics))
	return nil
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}