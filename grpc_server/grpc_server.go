package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func SetupAndRun(config configuration_loader.InitialConfiguration, inputChannel chan configuration_loader.Action, responsesChannel chan string, exitChannel chan bool) error {
	if config.ServerConfiguration == nil {
		return errors.New("Server parameters not set in the configuration file")
	}
	if config.ServerConfiguration.GRPCServerPort == 0 {
		return errors.New("Server port not set in the configuration file")
	}
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(config.ServerConfiguration.GRPCServerPort))
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	server := grpc.NewServer()
	pinsRegistered := make([]string, len(config.PinsActive))
	for _, pin := range config.PinsActive {
		pinsRegistered = append(pinsRegistered, pin.Name)
	}
	rpiServer := rpiHomeServer{nil, make(map[net.Addr][]string), make(map[net.Addr][]configuration_loader.Action)}
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiServer)
	go run(server, &rpiServer, &lis, exitChannel, inputChannel, responsesChannel)
	return nil
}

func run(server *grpc.Server, rpiServer *rpiHomeServer, listener *net.Listener, exitChannel chan bool, inputChannel chan configuration_loader.Action, responsesChannel chan string) {
	go server.Serve(*listener)
	for {
		select {
		case <-exitChannel:
			server.GracefulStop()
			fmt.Println("Exit signal received in gRPC server")
			exitChannel <- true
			return
		case action := <-inputChannel:
			response := ""
			if action.Pin == "start" {
				for _, pins := range rpiServer.clientsRegistered {
					for _, pin := range pins {
						response += pin + " "
					}
				}
			} else {
				response = "Action does not exist"
				for client, pins := range rpiServer.clientsRegistered {
					for _, pin := range pins {
						if pin == action.Pin {
							rpiServer.actionsToPerform[client] = append(rpiServer.actionsToPerform[client], action)
							response = "Action received!"
							continue
						}
					}
				}
				if response == "" {
					response = "Action does not exist"
				}
			}
			responsesChannel <- response
		}
	}
}

type rpiHomeServer struct {
	messages_protocol.RPIHomeServerServiceServer
	clientsRegistered map[net.Addr][]string
	actionsToPerform  map[net.Addr][]configuration_loader.Action
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
		s.actionsToPerform[p.Addr] = []configuration_loader.Action{}
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
	delete(s.actionsToPerform, p.Addr)
	return &messages_protocol.Empty{}, nil
}

func (s *rpiHomeServer) CheckForActions(ctx context.Context, empty *messages_protocol.Empty) (*messages_protocol.ActionsToPerform, error) {
	// TODO: wait for actions for this client (make use of a channel for each client) instead of looping asking for actions
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Error while extracting the peer from context")
	}
	actions := messages_protocol.ActionsToPerform{}
	for _, v := range s.actionsToPerform[p.Addr] {
		action := messages_protocol.PinStatePair{}
		action.Pin = &v.Pin
		action.State = &v.State
		actions.Actions = append(actions.Actions, &action)
	}
	delete(s.actionsToPerform, p.Addr)
	return &actions, nil
}
