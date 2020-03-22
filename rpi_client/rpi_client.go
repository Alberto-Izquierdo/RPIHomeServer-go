package rpi_client

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_client"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/message_generator"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"google.golang.org/grpc"
)

const timeBetweenReconnectionAttempts time.Duration = 10 * time.Second

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
	err = grpc_client.RegisterPinsToGRPCServer(client, config, config.AutomaticMessages)
	if err != nil {
		return errors.New("There was an error connecting to the gRPC server: " + err.Error())
	}

	go run(exitChannel, client, connection, config)

	return nil
}

func run(exitChannel chan bool, client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, config configuration_loader.InitialConfiguration) {
	telegramResponsesChannel := make(chan types.TelegramMessage)
	programmedActionOperationsChannel := make(chan types.ProgrammedActionOperation)
	messageGeneratorExitChannel := make(chan bool)
	message_generator.Run(config.AutomaticMessages, programmedActionOperationsChannel, telegramResponsesChannel, messageGeneratorExitChannel)
	grpcClientExitChannel := make(chan bool)
	go grpc_client.Run(programmedActionOperationsChannel, telegramResponsesChannel, grpcClientExitChannel, client, connection, config)
	<-exitChannel
	fmt.Println("Exit signal received in RPI client")
	grpcClientExitChannel <- true
	messageGeneratorExitChannel <- true
	exitChannel <- true
	gpio_manager.ClearAllPins()
}
