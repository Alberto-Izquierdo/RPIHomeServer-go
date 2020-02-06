package grpc_client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
)

const timeBetweenReconnectionAttempts time.Duration = 10 * time.Second
const numberOfReconnectingAttemptsUntilShutdown int = 30

func Run(config configuration_loader.InitialConfiguration, exitChannel chan bool, outputChannel chan configuration_loader.Action, mainExitChannel chan bool) error {
	client, connection, err := connectToGrpcServer(config)
	if err != nil {
		return errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}
	err = registerPinsToGRPCServer(client, config)
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
				err = checkForActions(client, outputChannel)
				if err != nil {
					fmt.Println("There was an error checking actions in gRPC client: ", err.Error())
					fmt.Println("Trying to reconnect to server...")
					time.Sleep(1 * time.Second)
					attempts := 0
					for err != nil {
						if attempts < numberOfReconnectingAttemptsUntilShutdown {
							err = registerPinsToGRPCServer(client, config)
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

func registerPinsToGRPCServer(client messages_protocol.RPIHomeServerServiceClient, config configuration_loader.InitialConfiguration) (err error) {
	var pins []string
	for _, pin := range config.PinsActive {
		pins = append(pins, pin.Name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := client.RegisterToServer(ctx, &messages_protocol.RegistrationMessage{PinsToHandle: pins})
	if err == nil && *result.Result != messages_protocol.RegistrationStatusCodes_Ok {
		errorMessage := result.Result.String()
		if *result.Result == messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered {
			errorMessage += "Pins repeated:"
			for _, v := range result.PinsRepeated {
				errorMessage += " " + v
			}
		}
		err = errors.New(errorMessage)
	}
	return err
}

func checkForActions(client messages_protocol.RPIHomeServerServiceClient, outputChannel chan configuration_loader.Action) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	actions, err := client.CheckForActions(ctx, &messages_protocol.Empty{})
	if err != nil {
		return err
	}
	for _, action := range actions.Actions {
		outputChannel <- configuration_loader.Action{*action.Pin, *action.State}
	}
	return nil
}

func unregisterPins(client messages_protocol.RPIHomeServerServiceClient) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.UnregisterToServer(ctx, &messages_protocol.Empty{})
	return err
}
