package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	kratosgrpc "github.com/go-kratos/kratos/v3/transport/grpc"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	pb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	v1 "cyber-ecosystem/gen/go/cyber/transfer/v1"
)

// Struct ----------------------------------------------------------------------------------------------------------------

type TransferService struct {
	pb.UnimplementedMobileTransferServiceServer

	log *slog.Logger
}

func NewTransferService(logger *slog.Logger) *TransferService {
	return &TransferService{
		log: logger.With("module", "service/transfer"),
	}
}

func (s *TransferService) RegisterGRPC(srv *kratosgrpc.Server) {
	pb.RegisterMobileTransferServiceServer(srv, s)
}

func (s *TransferService) RegisterHTTP(srv *kratoshttp.Server) {
	pb.RegisterMobileTransferServiceHTTPServer(srv, s)
}

func (s *TransferService) RegisterConnect(srv *connecttransport.Server) {
	pb.RegisterMobileTransferServiceConnectServer(srv, s)
}

// Handler ---------------------------------------------------------------------------------------------------------------

func (s *TransferService) Subscribe(req *v1.SubscribeRequest, stream grpc.ServerStreamingServer[v1.SubscribeResponse]) error {
	s.log.Info("Subscribe", "topic", req.GetTopic(), "last_event_id", req.GetLastEventId())

	for i := 0; i < 5; i++ {
		msg := &v1.SubscribeResponse{
			EventId:   fmt.Sprintf("%s-%d", req.GetTopic(), i+1),
			EventType: "message",
			Data:      []byte(fmt.Sprintf("event %d on topic %q", i+1, req.GetTopic())),
			Timestamp: timestamppb.Now(),
		}
		if err := stream.Send(msg); err != nil {
			return err
		}

		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil
}

func (s *TransferService) Echo(stream grpc.ClientStreamingServer[v1.EchoRequest, v1.EchoResponse]) error {
	var (
		totalMessages int32
		totalBytes    int64
		lastSeq       int64
	)
	start := time.Now()

	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}
		totalMessages++
		totalBytes += int64(len(req.GetData()))
		lastSeq = req.GetSequence()
	}

	return stream.SendAndClose(&v1.EchoResponse{
		TotalMessages: totalMessages,
		TotalBytes:    totalBytes,
		LastSequence:  lastSeq,
		DurationNs:    time.Since(start).Nanoseconds(),
	})
}

func (s *TransferService) Pipe(stream grpc.BidiStreamingServer[v1.PipeRequest, v1.PipeResponse]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return nil // EOF or client closed
		}
		if err := stream.Send(&v1.PipeResponse{
			Data:     req.GetData(),
			Type:     req.GetType(),
			Sequence: req.GetSequence(),
		}); err != nil {
			return err
		}
	}
}

func (s *TransferService) Raw(ctx context.Context, req *v1.RawRequest) (*httpbody.HttpBody, error) {
	s.log.Info("Raw", "content_type", req.GetContentType(), "data_len", len(req.GetData()))

	ct := req.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{
		ContentType: ct,
		Data:        req.GetData(),
	}, nil
}
