package echo

import (
	"context"
	"fmt"
	"github.com/lnashier/goarc/grpc"
	pb "grpcservice/internal/proto/echo"
	"io"
	"strings"
)

type Service struct {
	pb.UnsafeEchoServer
}

func Register(srv *grpc.Service) {
	pb.RegisterEchoServer(srv, &Service{})
}

func (s *Service) Single(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	return &pb.Response{Message: req.Message}, nil
}

func (s *Service) ServiceStream(req *pb.Request, stream pb.Echo_ServiceStreamServer) error {
	for i := 0; i < 5; i++ {
		err := stream.Send(&pb.Response{
			Message: fmt.Sprintf("%s (%d)", req.Message, i+1),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ClientStream(stream pb.Echo_ClientStreamServer) error {
	var messages []string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Response{
				Message: strings.Join(messages, ","),
			})
		}
		if err != nil {
			return err
		}

		messages = append(messages, req.Message)
	}
}

func (s *Service) BothStream(stream pb.Echo_BothStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		err = stream.Send(&pb.Response{Message: req.Message})
		if err != nil {
			return err
		}
	}
}
