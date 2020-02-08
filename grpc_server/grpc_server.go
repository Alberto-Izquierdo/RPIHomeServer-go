package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const timeWaitingForNewActions time.Duration = 2 * time.Second

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
	rpiServer := rpiHomeServer{clientsRegistered: make(map[net.Addr]dateStringsPair), actionsToPerform: make(map[net.Addr]chan configuration_loader.Action)}
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
				response = rpiServer.getPinsAndUpdateMap()
			} else {
				response = "Action does not exist"
				for client, pins := range rpiServer.clientsRegistered {
					for _, pin := range pins.Pins {
						if pin == action.Pin {
							rpiServer.actionsToPerform[client] <- action
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
	clientsRegistered map[net.Addr]dateStringsPair
	actionsToPerform  map[net.Addr]chan configuration_loader.Action
	mutex             sync.Mutex
}

type dateStringsPair struct {
	LastTimeConnected time.Time
	Pins              []string
}

func (s *rpiHomeServer) getPinsAndUpdateMap() string {
	s.mutex.Lock()
	response := ""
	var clientsToRemove []net.Addr
	for key, pins := range s.clientsRegistered {
		if time.Now().Sub(pins.LastTimeConnected) > time.Second*6 {
			clientsToRemove = append(clientsToRemove, key)
		} else {
			for _, pin := range pins.Pins {
				response += pin + " "
			}
		}
	}
	s.mutex.Unlock()
	for _, client := range clientsToRemove {
		s.removeClient(client)
	}
	return response
}

func (s *rpiHomeServer) RegisterToServer(ctx context.Context, message *messages_protocol.RegistrationMessage) (*messages_protocol.RegistrationResult, error) {
	result := new(messages_protocol.RegistrationResult)
	pinsRepeated := []string{}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, messagePin := range message.PinsToHandle {
		for _, clientPins := range s.clientsRegistered {
			for _, configPin := range clientPins.Pins {
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
		s.clientsRegistered[p.Addr] = dateStringsPair{time.Now(), message.PinsToHandle}
		s.actionsToPerform[p.Addr] = make(chan configuration_loader.Action)
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
	s.removeClient(p.Addr)
	return &messages_protocol.Empty{}, nil
}

func (s *rpiHomeServer) removeClient(client net.Addr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.clientsRegistered, client)
	delete(s.actionsToPerform, client)
}

func (s *rpiHomeServer) CheckForActions(ctx context.Context, empty *messages_protocol.Empty) (*messages_protocol.ActionsToPerform, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Error while extracting the peer from context")
	}
	actions := messages_protocol.ActionsToPerform{}
	select {
	case action := <-s.actionsToPerform[p.Addr]:
		protoAction := messages_protocol.PinStatePair{}
		protoAction.Pin = &action.Pin
		protoAction.State = &action.State
		actions.Actions = []*messages_protocol.PinStatePair{&protoAction}
	case <-time.After(timeWaitingForNewActions):
		break
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	previousClientData := s.clientsRegistered[p.Addr]
	s.clientsRegistered[p.Addr] = dateStringsPair{time.Now(), previousClientData.Pins}
	return &actions, nil
}
