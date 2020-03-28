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
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const timeWaitingForNewActions time.Duration = 2 * time.Second
const timeWaitingForClientConnection time.Duration = timeWaitingForNewActions * 5

func SetupAndRun(config configuration_loader.InitialConfiguration, inputChannel chan types.Action, programmedActionsChannel chan types.ProgrammedActionOperation, responsesChannel chan types.TelegramMessage, exitChannel chan bool) error {
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
		clientsRegistered: make(map[net.Addr]*clientRegisteredData),
		actionsToPerform:  make(map[net.Addr]chan types.Action),
		programmedActions: make(map[net.Addr]chan types.ProgrammedActionOperation),
		responsesChannel:  responsesChannel,
	}
	messages_protocol.RegisterRPIHomeServerServiceServer(server, &rpiServer)
	go run(server, &rpiServer, &lis, exitChannel, inputChannel, responsesChannel, programmedActionsChannel)
	return nil
}

func run(server *grpc.Server, rpiServer *rpiHomeServer, listener *net.Listener, exitChannel chan bool, inputChannel chan types.Action, responsesChannel chan types.TelegramMessage, programmedActionsChannel chan types.ProgrammedActionOperation) {
	go server.Serve(*listener)
	for {
		select {
		case <-exitChannel:
			server.GracefulStop()
			fmt.Println("Exit signal received in gRPC server")
			exitChannel <- true
			return
		case action := <-inputChannel:
			if action.Pin == "start" {
				activePins := "start " + rpiServer.getPinsAndUpdateMap()
				responsesChannel <- types.TelegramMessage{activePins, action.ChatId}
			} else {
				rpiServer.mutex.Lock()
				client, err := getClientAssociatedWithPin(action.Pin, rpiServer)
				if err != nil {
					responsesChannel <- types.TelegramMessage{err.Error(), action.ChatId}
				} else {
					rpiServer.actionsToPerform[client] <- action
				}
				rpiServer.mutex.Unlock()
			}
		case action := <-programmedActionsChannel:
			// Replace with a function
			rpiServer.mutex.Lock()
			if action.Operation == types.GET_ACTIONS {
				// Return the cached programmed actions
				response := "ProgrammedActions"
				for _, client := range rpiServer.clientsRegistered {
					for _, v := range *client.ProgrammedActions {
						response += " " + types.ProgrammedActionToString(v)
					}
				}
				responsesChannel <- types.TelegramMessage{response, action.ProgrammedAction.Action.ChatId}
			} else {
				client, err := getClientAssociatedWithPin(action.ProgrammedAction.Action.Pin, rpiServer)
				if err != nil {
					responsesChannel <- types.TelegramMessage{err.Error(), action.ProgrammedAction.Action.ChatId}
				} else {
					// Update the cache
					slice := rpiServer.clientsRegistered[client].ProgrammedActions
					if action.Operation == types.REMOVE {
						found := false
						for index, v := range *slice {
							if action.ProgrammedAction.Equals(v) {
								(*slice)[index] = (*slice)[len(*slice)-1]
								*slice = (*slice)[:len(*slice)-1]
								found = true
								break
							}
						}
						if found {
							// Send the operation
							rpiServer.programmedActions[client] <- action
						} else {
							responsesChannel <- types.TelegramMessage{"This programmed action did not exist", action.ProgrammedAction.Action.ChatId}
						}
					} else if action.Operation == types.CREATE {
						found := false
						for _, v := range *slice {
							if action.ProgrammedAction.Equals(v) {
								found = true
								break
							}
						}
						if !found {
							*slice = append(*slice, action.ProgrammedAction)
							// Send the operation
							rpiServer.programmedActions[client] <- action
						} else {
							responsesChannel <- types.TelegramMessage{"This programmed action already existed", action.ProgrammedAction.Action.ChatId}
						}
					}
					// Return the cached programmed actions
					response := "ProgrammedActions"
					for _, client := range rpiServer.clientsRegistered {
						for _, v := range *client.ProgrammedActions {
							response += " " + types.ProgrammedActionToString(v)
						}
					}
					responsesChannel <- types.TelegramMessage{response, action.ProgrammedAction.Action.ChatId}
				}
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
	clientsRegistered map[net.Addr]*clientRegisteredData
	actionsToPerform  map[net.Addr]chan types.Action
	programmedActions map[net.Addr]chan types.ProgrammedActionOperation
	responsesChannel  chan types.TelegramMessage
	mutex             sync.Mutex
}

type clientRegisteredData struct {
	LastTimeConnected time.Time
	Pins              []string
	ProgrammedActions *[]types.ProgrammedAction
}

func (s *rpiHomeServer) getPinsAndUpdateMap() string {
	s.mutex.Lock()
	response := ""
	var clientsToRemove []net.Addr
	for key, pins := range s.clientsRegistered {
		if time.Now().Sub(pins.LastTimeConnected) > timeWaitingForClientConnection {
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
			myTime := types.MyTime(time.Now())
			err := myTime.UnmarshalJSON([]byte(programmedAction.Time))
			if err != nil {
				continue
			}
			programmedActions = append(programmedActions, types.ProgrammedAction{
				Action: types.Action{
					Pin:   programmedAction.Action.Pin,
					State: programmedAction.Action.State,
				},
				Time:   myTime,
				Repeat: true,
			})
		}
		s.clientsRegistered[p.Addr] = &clientRegisteredData{
			LastTimeConnected: time.Now(),
			Pins:              message.PinsToHandle,
			ProgrammedActions: &programmedActions,
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
	s.mutex.Lock()
	s.clientsRegistered[p.Addr].LastTimeConnected = time.Now()
	s.mutex.Unlock()
	actions := messages_protocol.ActionsToPerform{}
	select {
	case action := <-s.actionsToPerform[p.Addr]:
		protoAction := messages_protocol.PinStatePair{
			Pin:    action.Pin,
			State:  action.State,
			ChatId: action.ChatId,
		}
		actions.Actions = []*messages_protocol.PinStatePair{&protoAction}
	case action := <-s.programmedActions[p.Addr]:
		programmedAction := messages_protocol.ProgrammedActionOperation{
			Operation: action.Operation,
			ProgrammedAction: &messages_protocol.ProgrammedAction{
				Action: &messages_protocol.PinStatePair{
					Pin:    action.ProgrammedAction.Action.Pin,
					State:  action.ProgrammedAction.Action.State,
					ChatId: action.ProgrammedAction.Action.ChatId,
				},
				Time:   action.ProgrammedAction.Time.Format("15:04:05"),
				Repeat: action.ProgrammedAction.Repeat,
			},
		}
		actions.ProgrammedActionOperations = []*messages_protocol.ProgrammedActionOperation{&programmedAction}
	case <-time.After(timeWaitingForNewActions):
		break
	}
	s.mutex.Lock()
	s.clientsRegistered[p.Addr].LastTimeConnected = time.Now()
	s.mutex.Unlock()
	return &actions, nil
}

func (s *rpiHomeServer) SendMessageToTelegram(ctx context.Context, message *messages_protocol.TelegramMessage) (*messages_protocol.Empty, error) {
	s.responsesChannel <- types.TelegramMessage{message.Message, message.ChatId}
	return &messages_protocol.Empty{}, nil
}
