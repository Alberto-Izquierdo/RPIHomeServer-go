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

func RunClient(config configuration_loader.InitialConfiguration, exitChannel chan bool, outputChannel chan configuration_loader.Action) {
	client, connection, err := connectToGrpcServer(config)
	defer connection.Close()
	if err != nil {
		fmt.Println("There was an error connecting to the gRPC server: ", err)
		return
	}
	err = registerPinsToGRPCServer(client, config)
	if err != nil {
		fmt.Println("There was an error registering pins in gRPC client: ", err)
		return
	}
	for {
		select {
		case <-exitChannel:
			err = unregisterPins(client)
			if err != nil {
				fmt.Println("There was an error unregistering in gRPC client: ", err)
			}
			return
		default:
			err = checkForActions(client, outputChannel)
			if err != nil {
				fmt.Println("There was an error checking actions in gRPC client: ", err)
			}
		}
	}
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
