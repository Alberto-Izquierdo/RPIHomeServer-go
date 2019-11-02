package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc"
)

func ConnectToGrpcServer(config configuration_loader.InitialConfiguration) (client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, err error) {
	connection, err = grpc.Dial(config.GRPCServerIp, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	if err != nil {
		return nil, nil, err
	}
	client = messages_protocol.NewRPIHomeServerServiceClient(connection)
	return client, connection, err
}

func RegisterPinsToGRPCServer(client messages_protocol.RPIHomeServerServiceClient, config configuration_loader.InitialConfiguration) (err error) {
	var pins []string
	for _, pin := range config.PinsActive {
		pins = append(pins, pin.Name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := client.RegisterToServer(ctx, &messages_protocol.RegistrationMessage{PinsToHandle: pins})
	if err == nil && *result.Result != messages_protocol.RegistrationStatusCodes_Ok {
		errorMessage := result.Result.String()
		if *result.Result != messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered {
			errorMessage += "Pins repeated:"
			for _, v := range result.PinsRepeated {
				errorMessage += " " + v
			}
		}
		err = errors.New(errorMessage)
	}
	return err
}
