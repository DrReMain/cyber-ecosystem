package transfer

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

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type TransferService struct {
	systempb.UnimplementedTransferServiceServer

	log *slog.Logger
}

func NewTransferService(logger *slog.Logger) *TransferService {
	return &TransferService{
		log: logger.With("module", "module/transfer_service"),
	}
}

func (s *TransferService) RegisterGRPC(srv *kratosgrpc.Server) {
	systempb.RegisterTransferServiceServer(srv, s)
}

func (s *TransferService) RegisterHTTP(srv *kratoshttp.Server) {
	systempb.RegisterTransferServiceHTTPServer(srv, s)
}

func (s *TransferService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterTransferServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *TransferService) Subscribe(in *systempb.SubscribeRequest, stream grpc.ServerStreamingServer[systempb.SubscribeResponse]) error {
	s.log.Info("Subscribe", "topic", in.GetTopic(), "last_event_id", in.GetLastEventId())

	for i := range 5 {
		msg := &systempb.SubscribeResponse{
			EventId:   fmt.Sprintf("%s-%d", in.GetTopic(), i+1),
			EventType: "message",
			Data:      fmt.Appendf(nil, "event %d on topic %q", i+1, in.GetTopic()),
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

func (s *TransferService) Echo(stream grpc.ClientStreamingServer[systempb.EchoRequest, systempb.EchoResponse]) error {
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

	return stream.SendAndClose(&systempb.EchoResponse{
		TotalMessages: totalMessages,
		TotalBytes:    totalBytes,
		LastSequence:  lastSeq,
		DurationNs:    time.Since(start).Nanoseconds(),
	})
}

func (s *TransferService) Pipe(stream grpc.BidiStreamingServer[systempb.PipeRequest, systempb.PipeResponse]) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			//nolint:nilerr // any Recv error ends the echo stream; close is the success path
			return nil
		}
		if err := stream.Send(&systempb.PipeResponse{
			Data:     req.GetData(),
			Type:     req.GetType(),
			Sequence: req.GetSequence(),
		}); err != nil {
			return err
		}
	}
}

func (s *TransferService) Raw(ctx context.Context, in *systempb.RawRequest) (*httpbody.HttpBody, error) {
	s.log.Info("Raw", "content_type", in.GetContentType(), "data_len", len(in.GetData()))

	ct := in.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{
		ContentType: ct,
		Data:        in.GetData(),
	}, nil
}
