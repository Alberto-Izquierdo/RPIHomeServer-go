package grpc_client

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/stretchr/testify/assert"
)

func createServer(t *testing.T) (chan bool, chan types.Action, chan types.TelegramMessage) {
	var serverConfig configuration_loader.InitialConfiguration
	serverConfig.ServerConfiguration = &configuration_loader.ServerConfiguration{GRPCServerPort: 8080}
	serverConfig.PinsActive = append(serverConfig.PinsActive, types.PairNamePin{"pin1", 90})
	serverExitChannel := make(chan bool)
	outputChannel := make(chan types.Action)
	responsesChannel := make(chan types.TelegramMessage)
	err := grpc_server.SetupAndRun(serverConfig, outputChannel, responsesChannel, nil, serverExitChannel)
	assert.Nil(t, err)
	return serverExitChannel, outputChannel, responsesChannel
}

func TestWrongConfig(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	_, _, err := ConnectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Empty config should return an error")
	config.GRPCServerIp = "asdf"
	_, _, err = ConnectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Wrong config should return an error")
	config.GRPCServerIp = "localhost:8080"
	_, _, err = ConnectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Connecting to a non existing server should return an error")
}

func TestConnectionToServer(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	clientConfig := configuration_loader.InitialConfiguration{GRPCServerIp: "localhost:8080"}
	{
		client, connection, err := ConnectToGrpcServer(clientConfig)
		assert.Equal(t, err, nil, "Valid config should not return an error")
		assert.NotEqual(t, client, nil, "Valid config should create a client")
		assert.NotEqual(t, connection, nil, "Valid config should create a connection")
	}
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestRegisterPinsToGRPCServer(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin1", 90})
	client1, _, _ := ConnectToGrpcServer(clientConfig)
	err := RegisterPinsToGRPCServer(client1, clientConfig, []types.ProgrammedAction{})
	assert.Equal(t, err, nil, "Correct register repeated should not return an error")
	client2, _, _ := ConnectToGrpcServer(clientConfig)
	err = RegisterPinsToGRPCServer(client2, clientConfig, []types.ProgrammedAction{})
	assert.NotEqual(t, err, nil, "Register with repeated pins should return an error")
	clientConfig.PinsActive = []types.PairNamePin{types.PairNamePin{"pin2", 90}}
	err = RegisterPinsToGRPCServer(client2, clientConfig, []types.ProgrammedAction{})
	assert.Equal(t, err, nil, "Register with valid pins should not return an error")
	err = UnregisterPins(client1)
	assert.Nil(t, err)
	err = UnregisterPins(client2)
	assert.Nil(t, err)
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestCheckForActions(t *testing.T) {
	serverExitChannel, serverInputChannel, serverOutputChannel := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin2", 90})
	client, _, err := ConnectToGrpcServer(clientConfig)
	assert.Nil(t, err)
	RegisterPinsToGRPCServer(client, clientConfig, []types.ProgrammedAction{})
	go func() {
		serverInputChannel <- types.Action{"pin2", true, 0}
		<-serverOutputChannel
	}()
	actions, _, err := CheckForActions(client)
	assert.Equal(t, len(actions), 1, "Actions received should only contain one element, instead it contains %d", len(actions))
	assert.Equal(t, actions[0].Pin, "pin2", "Action received should be \"pin2\", instead it is %s", actions[0].Pin)
	assert.Equal(t, actions[0].State, true, "Action state received should be \"true\"")
	err = UnregisterPins(client)
	assert.Nil(t, err)
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestSendMessageToTelegram(t *testing.T) {
	serverExitChannel, _, serverOutputChannel := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin2", 90})
	client, _, err := ConnectToGrpcServer(clientConfig)
	assert.Nil(t, err)
	RegisterPinsToGRPCServer(client, clientConfig, []types.ProgrammedAction{})
	SendMessageToTelegram(client, types.TelegramMessage{"Hello", 1})
	msg := <-serverOutputChannel
	assert.Equal(t, msg.Message, "Hello")
	assert.Equal(t, msg.ChatId, int64(1))
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestRun(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin2", 90})
	client, connection, err := ConnectToGrpcServer(clientConfig)
	assert.Nil(t, err)
	clientExitChannel := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		Run(clientExitChannel, client, connection)
	}()
	clientExitChannel <- true
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}
