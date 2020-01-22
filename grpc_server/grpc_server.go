package grpc_server

import (
	"context"
	"log"
	"net"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
)

func RunServer(config configuration_loader.InitialConfiguration, exitChannel chan bool, outputChannel chan configuration_loader.Action) {
	lis, err := net.Listen("tcp", "9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiHomeServer{})
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type rpiHomeServer struct {
	messages_protocol.RPIHomeServerServiceServer
}

func (s *rpiHomeServer) RegisterToServer(ctx context.Context, message *messages_protocol.RegistrationMessage) (*messages_protocol.RegistrationResult, error) {
	result := new(messages_protocol.RegistrationResult)
	code := messages_protocol.RegistrationStatusCodes_Ok
	result.Result = &code
	return result, nil
}
