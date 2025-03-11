// Package grpcserver реализует gRPC-сервер
package grpcserver

import (
	"context"
	"fmt"

	pb "github.com/ivanmolchanov1988/shortener/api"
	"github.com/ivanmolchanov1988/shortener/internal/core"
	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	shortenerService core.Shortener
	config           *server.Config
}

// NewShortenerServer - конструктор для gRPC-сервера
func NewShortenerServer(svc core.Shortener, cfg *server.Config) *ShortenerServer {
	return &ShortenerServer{
		shortenerService: svc,
		config:           cfg,
	}
}

// МЕТОДЫ //

// Shorten gRPC метод для сокращения ссылки (аналог HTTP PostURL)
func (s *ShortenerServer) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	userID := utils.GenUUID()

	response, err := s.shortenerService.Shorten(core.ShortenRequest{OriginalURL: req.Url}, userID)
	if err != nil {
		return nil, err
	}

	return &pb.ShortenResponse{Result: fmt.Sprintf("%s/%s", s.config.BaseURL, response.ShortURL)}, nil
}

// GetURL gRPC метод для получения оригинального URL (аналог HTTP GetURL)
func (s *ShortenerServer) GetURL(ctx context.Context, req *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	response, err := s.shortenerService.GetURL(core.GetURLRequest{ShortURL: req.ShortUrl})
	if err != nil {
		return nil, err
	}

	return &pb.GetURLResponse{OriginalUrl: response.OriginalURL}, nil
}

// Batch gRPC метод для пакетного сокращения ссылок (аналог HTTP Batch)
func (s *ShortenerServer) Batch(ctx context.Context, req *pb.BatchRequest) (*pb.BatchResponse, error) {
	userID := utils.GenUUID()

	var batchItems []core.BatchRequestItem
	for _, item := range req.Urls {
		batchItems = append(batchItems, core.BatchRequestItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	results, err := s.shortenerService.BatchURL(batchItems, userID)
	if err != nil {
		return nil, err
	}

	var response pb.BatchResponse
	for _, resItem := range results {
		response.Results = append(response.Results, &pb.BatchURLResult{
			CorrelationId: resItem.CorrelationID,
			ShortUrl:      fmt.Sprintf("%s/%s", s.config.BaseURL, resItem.ShortURL),
		})
	}

	return &response, nil
}

// GetUserURLs gRPC метод для получения всех ссылок пользователя (аналог HTTP GetUserURLs)
func (s *ShortenerServer) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	userID := req.UserId

	response, err := s.shortenerService.GetUserURLs(core.GetUserURLsRequest{UserID: userID})
	if err != nil {
		return nil, err
	}

	var grpcResponse pb.GetUserURLsResponse
	for _, url := range response.URLs {
		grpcResponse.Urls = append(grpcResponse.Urls, &pb.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &grpcResponse, nil
}

// DeleteURLs gRPC метод для удаления ссылок (аналог HTTP DeleteURLS)
func (s *ShortenerServer) DeleteURLs(ctx context.Context, req *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	_, err := s.shortenerService.DeleteURLs(core.DeleteURLsRequest{
		UserID:    req.UserId,
		ShortURLs: req.ShortUrls,
	})
	if err != nil {
		return nil, err
	}

	return &pb.DeleteURLsResponse{Status: "accepted"}, nil
}

// GetStats gRPC метод для получения статистики (аналог HTTP GetStats)
func (s *ShortenerServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	response, err := s.shortenerService.GetStats(core.GetStatsRequest{
		ClientIP: req.ClientIp,
		Subnet:   s.config.TrustedSubnet,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetStatsResponse{
		Urls:  int32(response.URLs),
		Users: int32(response.Users),
	}, nil
}
