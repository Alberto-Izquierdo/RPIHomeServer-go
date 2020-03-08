package rpi_client

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_client"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"google.golang.org/grpc"
)

const timeBetweenReconnectionAttempts time.Duration = 2 * time.Second
const numberOfReconnectingAttemptsUntilShutdown int = 2

func connectToGrpcServer(config configuration_loader.InitialConfiguration) (client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, err error) {
	client, connection, err = grpc_client.ConnectToGrpcServer(config)
	if err != nil {
		err = errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}
	return client, connection, err
}

func SetupAndRun(config configuration_loader.InitialConfiguration, exitChannel chan bool) error {
	// gpio manager config
	err := gpio_manager.Setup(config.PinsActive)
	if err != nil {
		return err
	}

	// gRPC client config
	client, connection, err := connectToGrpcServer(config)
	if err != nil {
		return err
	}
	var programmedActions []types.ProgrammedAction
	for _, programmedMessage := range config.AutomaticMessages {
		programmedActions = append(programmedActions, types.ProgrammedAction{
			Action: programmedMessage.Action,
			Repeat: programmedMessage.Repeat,
			Time:   programmedMessage.Time,
		})
	}
	err = grpc_client.RegisterPinsToGRPCServer(client, config, programmedActions)
	if err != nil {
		return errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}

	go run(exitChannel, client, connection)

	return nil
}

func run(exitChannel chan bool, client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn) {
	grpcClientExitChannel := make(chan bool)
	go runGRPCClientLoop(grpcClientExitChannel, client, connection)
	<-exitChannel
	fmt.Println("Exit signal received in RPI client")
	grpcClientExitChannel <- true
	exitChannel <- true
	gpio_manager.ClearAllPins()
}

func runGRPCClientLoop(grpcClientExitChannel chan bool, client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn) {
	defer connection.Close()
	for {
		select {
		case <-grpcClientExitChannel:
			err := grpc_client.UnregisterPins(client)
			if err != nil {
				fmt.Println("There was an error unregistering in gRPC client: ", err.Error())
			}
			fmt.Println("Exit signal received in gRPC client")
			return
		default:
			actions, _, err := grpc_client.CheckForActions(client)
			if err != nil {
				fmt.Println("There was an error checking actions in gRPC client: ", err.Error())
				fmt.Println("Trying to reconnect to server...")
				time.Sleep(1 * time.Second)
				attempts := 0
				for err != nil {
					select {
					case <-grpcClientExitChannel:
						fmt.Println("Exit signal received in gRPC client")
						return
					default:
						if attempts < numberOfReconnectingAttemptsUntilShutdown {
							var programmedActions []types.ProgrammedAction
							// TODO: add programmed actions from message_generator
							err = grpc_client.RegisterPinsToGRPCServer(client, configuration_loader.InitialConfiguration{}, programmedActions)
							if err != nil {
								fmt.Println("There was an error connecting to the gRPC server: " + err.Error())
								fmt.Println("Trying again in " + timeBetweenReconnectionAttempts.String() + "...")
								attempts += 1
							} else {
								fmt.Println("Reconnected!")
							}
						} else {
							fmt.Println("Could not reconnect to the server, closing the application...")
							time.Sleep(time.Second * 1)
							return
						}
					}
				}
			} else {
				for _, action := range actions {
					gpio_manager.HandleAction(action)
				}
			}
		}
	}
}
