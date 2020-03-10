package grpc_client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const timeBetweenReconnectionAttempts time.Duration = 10 * time.Second
const numberOfReconnectingAttemptsUntilShutdown int = 30

const EmptyPinsMessage string = "There are not any pins active, gRPC client will not be run"

func ConnectToGrpcServer(config configuration_loader.InitialConfiguration) (client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, err error) {
	connection, err = grpc.Dial(config.GRPCServerIp, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	if err == nil {
		client = messages_protocol.NewRPIHomeServerServiceClient(connection)
	}
	return client, connection, err
}

func Run(grpcClientExitChannel chan bool, client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn) {
	defer connection.Close()
	for {
		select {
		case <-grpcClientExitChannel:
			err := UnregisterPins(client)
			if err != nil {
				fmt.Println("There was an error unregistering in gRPC client: ", err.Error())
			}
			fmt.Println("Exit signal received in gRPC client")
			return
		default:
			actions, _, err := CheckForActions(client)
			if err != nil {
				fmt.Println("There was an error checking actions in gRPC client: ", err.Error())
				fmt.Println("Trying to reconnect to server...")
				time.Sleep(1 * time.Second)
				for err != nil {
					select {
					case <-grpcClientExitChannel:
						fmt.Println("Exit signal received in gRPC client")
						return
					default:
						var programmedActions []types.ProgrammedAction
						// TODO: add programmed actions from message_generator
						err = RegisterPinsToGRPCServer(client, configuration_loader.InitialConfiguration{}, programmedActions)
						if err != nil {
							fmt.Println("There was an error connecting to the gRPC server: " + err.Error())
							fmt.Println("Trying again in " + timeBetweenReconnectionAttempts.String() + "...")
							time.Sleep(timeBetweenReconnectionAttempts)
						} else {
							fmt.Println("Reconnected!")
						}
					}
				}
			} else {
				for _, action := range actions {
					success, _ := gpio_manager.HandleAction(action)
					message := ""
					if success == true {
						message = "Action " + action.Pin + " successful"
					} else {
						message = "Action " + action.Pin + " not successful"
					}
					SendMessageToTelegram(client, types.TelegramMessage{message, action.ChatId})
				}
			}
		}
	}
}

func RegisterPinsToGRPCServer(client messages_protocol.RPIHomeServerServiceClient,
	config configuration_loader.InitialConfiguration,
	programmedActions []types.ProgrammedAction) (err error) {
	var pins []string
	for _, pin := range config.PinsActive {
		pins = append(pins, pin.Name)
	}
	// TODO: add programmed actions
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := client.RegisterToServer(ctx, &messages_protocol.RegistrationMessage{PinsToHandle: pins})
	if err == nil && result.Result != messages_protocol.RegistrationStatusCodes_Ok {
		errorMessage := result.Result.String()
		if result.Result == messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered {
			errorMessage += "Pins repeated:"
			for _, v := range result.PinsRepeated {
				errorMessage += " " + v
			}
		}
		err = errors.New(errorMessage)
	}
	return err
}

func CheckForActions(client messages_protocol.RPIHomeServerServiceClient) ([]types.Action, []types.ProgrammedActionOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	protoActions, err := client.CheckForActions(ctx, &messages_protocol.Empty{})
	if err != nil {
		return nil, nil, err
	}
	var actions []types.Action
	for _, action := range protoActions.Actions {
		actions = append(actions, types.Action{action.Pin, action.State, action.ChatId})
	}
	var programmedActionOperations []types.ProgrammedActionOperation
	for _, programmedAction := range protoActions.ProgrammedActionOperations {
		timestamp, err := ptypes.Timestamp(programmedAction.ProgrammedAction.Time)
		if err != nil {
			continue
		}
		for timestamp.Before(time.Now()) {
			timestamp = timestamp.Add(time.Hour * 24)
		}
		action := types.ProgrammedActionOperation{
			Operation: programmedAction.Operation,
			ProgrammedAction: types.ProgrammedAction{
				Action: types.Action{
					programmedAction.ProgrammedAction.Action.Pin,
					programmedAction.ProgrammedAction.Action.State,
					programmedAction.ProgrammedAction.Action.ChatId,
				},
				Time:   types.MyTime(timestamp),
				Repeat: programmedAction.ProgrammedAction.Repeat,
			},
		}
		programmedActionOperations = append(programmedActionOperations, action)
	}
	return actions, programmedActionOperations, nil
}

func UnregisterPins(client messages_protocol.RPIHomeServerServiceClient) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.UnregisterToServer(ctx, &messages_protocol.Empty{})
	return err
}

func SendMessageToTelegram(client messages_protocol.RPIHomeServerServiceClient, message types.TelegramMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.SendMessageToTelegram(ctx, &messages_protocol.TelegramMessage{Message: message.Message, ChatId: message.ChatId})
	return err
}
