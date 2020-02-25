package grpc_client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const timeBetweenReconnectionAttempts time.Duration = 10 * time.Second
const numberOfReconnectingAttemptsUntilShutdown int = 30

const EmptyPinsMessage string = "There are not any pins active, gRPC client will not be run"

func Run(config configuration_loader.InitialConfiguration,
	exitChannel chan bool, outputChannel chan types.Action,
	programmedActionsChannel chan types.ProgrammedActionOperation,
	mainExitChannel chan bool) error {

	if len(config.PinsActive) == 0 {
		return errors.New(EmptyPinsMessage)
	}
	client, connection, err := connectToGrpcServer(config)
	if err != nil {
		return errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}
	var programmedActions []types.ProgrammedAction
	for _, programmedMessage := range config.AutomaticMessages {
		programmedActions = append(programmedActions, types.ProgrammedAction{
			Action: programmedMessage.Action,
			Repeat: programmedMessage.Repeat,
			Time:   programmedMessage.Time,
		})
	}
	err = registerPinsToGRPCServer(client, config, programmedActions)
	if err != nil {
		return errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}
	go func() {
		for {
			select {
			case <-exitChannel:
				err = unregisterPins(client)
				if err != nil {
					fmt.Println("There was an error unregistering in gRPC client: ", err.Error())
				}
				fmt.Println("Exit signal received in gRPC client")
				exitChannel <- true
				return
			default:
				err = checkForActions(client, outputChannel, programmedActionsChannel)
				if err != nil {
					fmt.Println("There was an error checking actions in gRPC client: ", err.Error())
					fmt.Println("Trying to reconnect to server...")
					time.Sleep(1 * time.Second)
					attempts := 0
					for err != nil {
						if attempts < numberOfReconnectingAttemptsUntilShutdown {
							var programmedActions []types.ProgrammedAction
							// TODO: add programmed actions from message_generator
							err = registerPinsToGRPCServer(client, config, programmedActions)
							if err != nil {
								select {
								case <-exitChannel:
									fmt.Println("Exit signal received in gRPC client")
									exitChannel <- true
									return
								case <-time.After(timeBetweenReconnectionAttempts):
									fmt.Println("There was an error connecting to the gRPC server: " + err.Error())
									fmt.Println("Trying again in " + timeBetweenReconnectionAttempts.String() + "...")
									attempts += 1
								}
							} else {
								fmt.Println("Reconnected!")
								break
							}
						} else {
							fmt.Println("Could not reconnect to the server, closing the application...")
							mainExitChannel <- true
							time.Sleep(time.Second * 1)
							break
						}
					}
				}
			}
		}
		defer connection.Close()
	}()
	return nil
}

func connectToGrpcServer(config configuration_loader.InitialConfiguration) (client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, err error) {
	connection, err = grpc.Dial(config.GRPCServerIp, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	if err == nil {
		client = messages_protocol.NewRPIHomeServerServiceClient(connection)
	}
	return client, connection, err
}

func registerPinsToGRPCServer(client messages_protocol.RPIHomeServerServiceClient,
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

func checkForActions(client messages_protocol.RPIHomeServerServiceClient,
	outputChannel chan types.Action,
	programmedActionsChannel chan types.ProgrammedActionOperation) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	actions, err := client.CheckForActions(ctx, &messages_protocol.Empty{})
	if err != nil {
		return err
	}
	for _, action := range actions.Actions {
		outputChannel <- types.Action{action.Pin, action.State}
	}
	for _, programmedAction := range actions.ProgrammedActionOperations {
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
				},
				Time:   types.MyTime(timestamp),
				Repeat: programmedAction.ProgrammedAction.Repeat,
			},
		}
		programmedActionsChannel <- action
	}
	return nil
}

func unregisterPins(client messages_protocol.RPIHomeServerServiceClient) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.UnregisterToServer(ctx, &messages_protocol.Empty{})
	return err
}
