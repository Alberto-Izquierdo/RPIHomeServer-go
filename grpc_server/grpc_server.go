package grpc_server

import (
	"context"
	"errors"
	"net"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
)

func SetupAndRun(config configuration_loader.InitialConfiguration, exitChannel chan bool, outputChannel chan configuration_loader.Action) error {
	if config.GRPCServerConfiguration == nil {
		return errors.New("Server parameters not set in the configuration file")
	}
	lis, err := net.Listen("tcp", config.GRPCServerConfiguration.Port)
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	server := grpc.NewServer()
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiHomeServer{nil, outputChannel})
	if err := server.Serve(lis); err != nil {
		return errors.New("failed to serve: " + err.Error())
	}
	go run(server, &lis, exitChannel)
	return nil
}

func run(server *grpc.Server, listener *net.Listener, exitChannel chan bool) {
	go server.Serve(*listener)
	<-exitChannel
	server.GracefulStop()
}

type rpiHomeServer struct {
	messages_protocol.RPIHomeServerServiceServer
	outputChannel chan configuration_loader.Action
}

func (s *rpiHomeServer) RegisterToServer(ctx context.Context, message *messages_protocol.RegistrationMessage) (*messages_protocol.RegistrationResult, error) {
	result := new(messages_protocol.RegistrationResult)
	code := messages_protocol.RegistrationStatusCodes_Ok
	result.Result = &code
	return result, nil
}
