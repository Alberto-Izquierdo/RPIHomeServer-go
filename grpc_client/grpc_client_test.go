package grpc_client

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
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

func createServer(t *testing.T) (chan bool, chan configuration_loader.Action, chan string) {
	var serverConfig configuration_loader.InitialConfiguration
	serverConfig.ServerConfiguration = &configuration_loader.ServerConfiguration{GRPCServerPort: 8080}
	serverConfig.PinsActive = append(serverConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	serverExitChannel := make(chan bool)
	outputChannel := make(chan configuration_loader.Action)
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
	<-serverExitChannel
	time.Sleep(100 * time.Millisecond)
}

func TestRegisterPinsToGRPCServer(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	client1, _, _ := connectToGrpcServer(clientConfig)
	err := registerPinsToGRPCServer(client1, clientConfig)
	assert.Equal(t, err, nil, "Correct register repeated should not return an error")
	client2, _, _ := connectToGrpcServer(clientConfig)
	err = registerPinsToGRPCServer(client2, clientConfig)
	assert.NotEqual(t, err, nil, "Register with repeated pins should return an error")
	clientConfig.PinsActive = []gpio_manager.PairNamePin{gpio_manager.PairNamePin{"pin2", 90}}
	err = registerPinsToGRPCServer(client2, clientConfig)
	assert.Equal(t, err, nil, "Register with valid pins should not return an error")
	err = unregisterPins(client1)
	assert.Nil(t, err)
	err = unregisterPins(client2)
	assert.Nil(t, err)
	serverExitChannel <- true
	<-serverExitChannel
	time.Sleep(100 * time.Millisecond)
}

func TestCheckForActions(t *testing.T) {
	serverExitChannel, serverInputChannel, serverOutputChannel := createServer(t)
	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, gpio_manager.PairNamePin{"pin2", 90})
	client, _, _ := connectToGrpcServer(clientConfig)
	registerPinsToGRPCServer(client, clientConfig)
	clientOutputChannel := make(chan configuration_loader.Action)
	go checkForActions(client, clientOutputChannel, nil)
	serverInputChannel <- configuration_loader.Action{"pin2", true}
	<-serverOutputChannel
	action := <-clientOutputChannel
	assert.Equal(t, action.Pin, "pin2", "Action received should be \"pin2\", instead it is %s", action.Pin)
	assert.Equal(t, action.State, true, "Action state received should be \"true\"")
	err := unregisterPins(client)
	assert.Nil(t, err)
	serverExitChannel <- true
	<-serverExitChannel
	time.Sleep(100 * time.Millisecond)
}

func TestRun(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)
	clientExitChannel := make(chan bool)
	clientOutputChannel := make(chan configuration_loader.Action)
	clientConfig := configuration_loader.InitialConfiguration{GRPCServerIp: "localhost:8080"}
	err := Run(clientConfig, clientExitChannel, clientOutputChannel, nil, nil)
	assert.NotEqual(t, err, nil, "Config without pins should return an error")
	clientConfig.PinsActive = append(clientConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	err = Run(clientConfig, clientExitChannel, clientOutputChannel, nil, nil)
	assert.Equal(t, err, nil, "Correct config should not return an error")
	time.Sleep(1 * time.Second)
	clientExitChannel <- true
	<-clientExitChannel
	serverExitChannel <- true
	<-serverExitChannel
}
