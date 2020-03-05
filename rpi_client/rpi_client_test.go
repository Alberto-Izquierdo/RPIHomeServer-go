package rpi_client

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_client"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/stretchr/testify/assert"
)

func TestWrongConfig(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	_, _, err := connectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Empty config should return an error")
	config.GRPCServerIp = "asdf"
	_, _, err = connectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Wrong config should return an error")
	config.GRPCServerIp = "localhost:8080"
	_, _, err = connectToGrpcServer(config)
	assert.NotEqual(t, err, nil, "Connecting to a non existing server should return an error")
}

func createServer(t *testing.T) (chan bool, chan types.Action, chan string) {
	var serverConfig configuration_loader.InitialConfiguration
	serverConfig.ServerConfiguration = &configuration_loader.ServerConfiguration{GRPCServerPort: 8080}
	serverConfig.PinsActive = append(serverConfig.PinsActive, types.PairNamePin{"pin1", 90})
	serverExitChannel := make(chan bool)
	outputChannel := make(chan types.Action)
	responsesChannel := make(chan string)
	err := grpc_server.SetupAndRun(serverConfig, outputChannel, responsesChannel, nil, serverExitChannel)
	assert.Nil(t, err)
	return serverExitChannel, outputChannel, responsesChannel
}

func TestConnectionToServer(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	clientConfig := configuration_loader.InitialConfiguration{GRPCServerIp: "localhost:8080"}
	{
		client, connection, err := connectToGrpcServer(clientConfig)
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
	client1, _, _ := connectToGrpcServer(clientConfig)
	err := grpc_client.RegisterPinsToGRPCServer(client1, clientConfig, []types.ProgrammedAction{})
	assert.Equal(t, err, nil, "Correct register repeated should not return an error")
	client2, _, _ := connectToGrpcServer(clientConfig)
	err = grpc_client.RegisterPinsToGRPCServer(client2, clientConfig, []types.ProgrammedAction{})
	assert.NotEqual(t, err, nil, "Register with repeated pins should return an error")
	clientConfig.PinsActive = []types.PairNamePin{types.PairNamePin{"pin2", 90}}
	err = grpc_client.RegisterPinsToGRPCServer(client2, clientConfig, []types.ProgrammedAction{})
	assert.Equal(t, err, nil, "Register with valid pins should not return an error")
	err = grpc_client.UnregisterPins(client1)
	assert.Nil(t, err)
	err = grpc_client.UnregisterPins(client2)
	assert.Nil(t, err)
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestCheckForActions(t *testing.T) {
	serverExitChannel, serverInputChannel, serverOutputChannel := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin2", 90})
	client, _, err := connectToGrpcServer(clientConfig)
	assert.Nil(t, err)
	grpc_client.RegisterPinsToGRPCServer(client, clientConfig, []types.ProgrammedAction{})
	go func() {
		serverInputChannel <- types.Action{"pin2", true}
		<-serverOutputChannel
	}()
	actions, _, err := grpc_client.CheckForActions(client)
	assert.Equal(t, len(actions), 1, "Actions received should only contain one element, instead it contains %d", len(actions))
	assert.Equal(t, actions[0].Pin, "pin2", "Action received should be \"pin2\", instead it is %s", actions[0].Pin)
	assert.Equal(t, actions[0].State, true, "Action state received should be \"true\"")
	err = grpc_client.UnregisterPins(client)
	assert.Nil(t, err)
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestRun(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	clientExitChannel := make(chan bool)
	clientConfig := configuration_loader.InitialConfiguration{GRPCServerIp: "localhost:8080"}
	err := SetupAndRun(clientConfig, clientExitChannel)
	assert.NotEqual(t, err, nil, "Config without pins should return an error")
	clientConfig.PinsActive = append(clientConfig.PinsActive, types.PairNamePin{"pin1", 90})
	err = SetupAndRun(clientConfig, clientExitChannel)
	assert.Equal(t, err, nil, "Correct config should not return an error")
	time.Sleep(1 * time.Second)
	clientExitChannel <- true
	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}
