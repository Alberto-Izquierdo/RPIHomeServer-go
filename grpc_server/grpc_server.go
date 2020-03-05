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
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const timeWaitingForNewActions time.Duration = 2 * time.Second

func SetupAndRun(config configuration_loader.InitialConfiguration, inputChannel chan types.Action, responsesChannel chan string, programmedActionsChannel chan types.ProgrammedActionOperation, exitChannel chan bool) error {
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
	rpiServer := rpiHomeServer{
		clientsRegistered: make(map[net.Addr]clientRegisteredData),
		actionsToPerform:  make(map[net.Addr]chan types.Action),
		programmedActions: make(map[net.Addr]chan types.ProgrammedActionOperation),
	}
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiServer)
	go run(server, &rpiServer, &lis, exitChannel, inputChannel, responsesChannel, programmedActionsChannel)
	return nil
}

func run(server *grpc.Server, rpiServer *rpiHomeServer, listener *net.Listener, exitChannel chan bool, inputChannel chan types.Action, responsesChannel chan string, programmedActionsChannel chan types.ProgrammedActionOperation) {
	go server.Serve(*listener)
	for {
		select {
		case <-exitChannel:
			server.GracefulStop()
			fmt.Println("Exit signal received in gRPC server")
			return
		case action := <-inputChannel:
			if action.Pin == "start" {
				responsesChannel <- rpiServer.getPinsAndUpdateMap()
			} else {
				rpiServer.mutex.Lock()
				client, err := getClientAssociatedWithPin(action.Pin, rpiServer)
				if err != nil {
					responsesChannel <- err.Error()
				} else {
					rpiServer.actionsToPerform[client] <- action
					responsesChannel <- "Action received!"
				}
				rpiServer.mutex.Unlock()
			}
		case action := <-programmedActionsChannel:
			rpiServer.mutex.Lock()
			client, err := getClientAssociatedWithPin(action.ProgrammedAction.Action.Pin, rpiServer)
			if err != nil {
				responsesChannel <- err.Error()
			} else {
				rpiServer.programmedActions[client] <- action
				responsesChannel <- "Action received!"
			}
			rpiServer.mutex.Unlock()

		}
	}
}

func getClientAssociatedWithPin(pinName string, rpiServer *rpiHomeServer) (net.Addr, error) {
	for client, pins := range rpiServer.clientsRegistered {
		for _, pin := range pins.Pins {
			if pin == pinName {
				return client, nil
			}
		}
	}
	return nil, errors.New("Pin does not exist: " + pinName)
}

type rpiHomeServer struct {
	messages_protocol.RPIHomeServerServiceServer
	clientsRegistered map[net.Addr]clientRegisteredData
	actionsToPerform  map[net.Addr]chan types.Action
	programmedActions map[net.Addr]chan types.ProgrammedActionOperation
	mutex             sync.Mutex
}

type clientRegisteredData struct {
	LastTimeConnected time.Time
	Pins              []string
	ProgrammedActions []types.ProgrammedAction
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
		_, err := getClientAssociatedWithPin(messagePin, s)
		if err == nil {
			pinsRepeated = append(pinsRepeated, messagePin)
		}
	}
	if len(pinsRepeated) == 0 {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("Error while extracting the peer from context")
		}
		var programmedActions []types.ProgrammedAction
		for _, programmedAction := range message.ProgrammedActions {
			timestamp, err := ptypes.Timestamp(programmedAction.Time)
			if err != nil {
				continue
			}
			programmedActions = append(programmedActions, types.ProgrammedAction{
				Action: types.Action{
					Pin:   programmedAction.Action.Pin,
					State: programmedAction.Action.State,
				},
				Time:   types.MyTime(timestamp),
				Repeat: true,
			})
		}
		s.clientsRegistered[p.Addr] = clientRegisteredData{
			LastTimeConnected: time.Now(),
			Pins:              message.PinsToHandle,
			ProgrammedActions: programmedActions,
		}
		s.actionsToPerform[p.Addr] = make(chan types.Action)
		s.programmedActions[p.Addr] = make(chan types.ProgrammedActionOperation)
		code := messages_protocol.RegistrationStatusCodes_Ok
		result.Result = code
	} else {
		code := messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered
		result.Result = code
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
	delete(s.programmedActions, client)
}

func (s *rpiHomeServer) CheckForActions(ctx context.Context, empty *messages_protocol.Empty) (*messages_protocol.ActionsToPerform, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Error while extracting the peer from context")
	}
	actions := messages_protocol.ActionsToPerform{}
	select {
	case action := <-s.actionsToPerform[p.Addr]:
		protoAction := messages_protocol.PinStatePair{Pin: action.Pin, State: action.State}
		actions.Actions = []*messages_protocol.PinStatePair{&protoAction}
	case action := <-s.programmedActions[p.Addr]:
		time, err := ptypes.TimestampProto(time.Time(action.ProgrammedAction.Time))
		if err != nil {
			return nil, err
		}
		programmedAction := messages_protocol.ProgrammedActionOperation{
			Operation: action.Operation,
			ProgrammedAction: &messages_protocol.ProgrammedAction{
				Action: &messages_protocol.PinStatePair{
					Pin:   action.ProgrammedAction.Action.Pin,
					State: action.ProgrammedAction.Action.State,
				},
				Time:   time,
				Repeat: action.ProgrammedAction.Repeat,
			},
		}
		actions.ProgrammedActionOperations = []*messages_protocol.ProgrammedActionOperation{&programmedAction}
	case <-time.After(timeWaitingForNewActions):
		break
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	previousClientData := s.clientsRegistered[p.Addr]
	s.clientsRegistered[p.Addr] = clientRegisteredData{time.Now(), previousClientData.Pins, previousClientData.ProgrammedActions}
	return &actions, nil
}
