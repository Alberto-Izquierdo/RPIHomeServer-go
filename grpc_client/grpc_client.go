package grpc_client

import (
	"context"
	"errors"
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

func ConnectToGrpcServer(config configuration_loader.InitialConfiguration) (client messages_protocol.RPIHomeServerServiceClient, connection *grpc.ClientConn, err error) {
	connection, err = grpc.Dial(config.GRPCServerIp, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	if err == nil {
		client = messages_protocol.NewRPIHomeServerServiceClient(connection)
	}
	return client, connection, err
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
		actions = append(actions, types.Action{action.Pin, action.State})
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
