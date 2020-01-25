package grpc_server

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func SetupAndRun(config configuration_loader.InitialConfiguration, outputChannel chan configuration_loader.Action, exitChannel chan bool) error {
	if config.GRPCServerConfiguration == nil {
		return errors.New("Server parameters not set in the configuration file")
	}
	if config.GRPCServerConfiguration.Port == 0 {
		return errors.New("Server port not set in the configuration file")
	}
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(config.GRPCServerConfiguration.Port))
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	server := grpc.NewServer()
	pinsRegistered := make([]string, len(config.PinsActive))
	for _, pin := range config.PinsActive {
		pinsRegistered = append(pinsRegistered, pin.Name)
	}
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiHomeServer{nil, outputChannel, make(map[net.Addr][]string)})
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
	outputChannel     chan configuration_loader.Action
	clientsRegistered map[net.Addr][]string
}

func (s *rpiHomeServer) RegisterToServer(ctx context.Context, message *messages_protocol.RegistrationMessage) (*messages_protocol.RegistrationResult, error) {
	result := new(messages_protocol.RegistrationResult)
	pinsRepeated := []string{}
	for _, messagePin := range message.PinsToHandle {
		for _, clientPins := range s.clientsRegistered {
			for _, configPin := range clientPins {
				if messagePin == configPin {
					pinsRepeated = append(pinsRepeated, messagePin)
				}
			}
		}
	}
	if len(pinsRepeated) == 0 {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("Error while extracting the peer from context")
		}
		s.clientsRegistered[p.Addr] = message.PinsToHandle
		code := messages_protocol.RegistrationStatusCodes_Ok
		result.Result = &code
	} else {
		code := messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered
		result.Result = &code
		result.PinsRepeated = pinsRepeated
	}
	return result, nil
}

func (s *rpiHomeServer) UnregisterToServer(ctx context.Context, empty *messages_protocol.Empty) (*messages_protocol.Empty, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Error while extracting the peer from context")
	}
	delete(s.clientsRegistered, p.Addr)
	return nil, nil
}
