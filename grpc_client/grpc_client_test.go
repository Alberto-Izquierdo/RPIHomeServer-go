package grpc_client

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
)

func TestWrongConfig(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	_, _, err := connectToGrpcServer(config)
	if err == nil {
		t.Errorf("Empty config should return an error")
	}
	config.GRPCServerIp = "asdf"
	_, _, err = connectToGrpcServer(config)
	if err == nil {
		t.Errorf("Wrong config should return an error")
	}
	config.GRPCServerIp = "localhost:8088"
	_, _, err = connectToGrpcServer(config)
	if err == nil {
		t.Errorf("Connecting to a non existing server should return an error")
	}
}

func createServer(t *testing.T) (chan bool, chan configuration_loader.Action, chan string) {
	var serverConfig configuration_loader.InitialConfiguration
	serverConfig.ServerConfiguration = &configuration_loader.ServerConfiguration{}
	serverConfig.ServerConfiguration.GRPCServerPort = 8080
	serverConfig.PinsActive = append(serverConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	serverExitChannel := make(chan bool)
	outputChannel := make(chan configuration_loader.Action)
	responsesChannel := make(chan string)
	err := grpc_server.SetupAndRun(serverConfig, outputChannel, responsesChannel, serverExitChannel)
	if err != nil {
		t.Errorf("Setting up the server should not return an error, instead we got %s", err.Error())
	}
	return serverExitChannel, outputChannel, responsesChannel
}

func TestConnectionToServer(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)

	var clientConfig configuration_loader.InitialConfiguration
	//TODO: Assert pins is not empty
	clientConfig.GRPCServerIp = "localhost:8080"
	{
		client, connection, err := connectToGrpcServer(clientConfig)
		if err != nil {
			t.Errorf("Valid config should not return an error")
		} else if client == nil {
			t.Errorf("Valid config should create a client")
		} else if connection == nil {
			t.Errorf("Valid config should create a connection")
		}
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
	if err != nil {
		t.Error("Correct register repeated should not return an error")
	}

	client2, _, _ := connectToGrpcServer(clientConfig)
	err = registerPinsToGRPCServer(client2, clientConfig)
	if err == nil {
		t.Error("Register with repeated pins should return an error")
	}

	clientConfig.PinsActive = []gpio_manager.PairNamePin{gpio_manager.PairNamePin{"pin2", 90}}
	err = registerPinsToGRPCServer(client2, clientConfig)
	if err != nil {
		t.Error("Register with valid pins should not return an error")
	}

	err = unregisterPins(client1)
	if err != nil {
		t.Errorf("Valid unregister should not return an error: %s", err.Error())
	}

	err = unregisterPins(client2)
	if err != nil {
		t.Errorf("Valid unregister should not return an error: %s", err.Error())
	}

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

	go checkForActions(client, clientOutputChannel)
	serverInputChannel <- configuration_loader.Action{"pin2", true}
	<-serverOutputChannel

	action := <-clientOutputChannel
	if action.Pin != "pin2" {
		t.Errorf("Action received should be \"pin2\", instead it is %s", action.Pin)
	} else if action.State != true {
		t.Error("Action state received should be \"true\"")
	}

	err := unregisterPins(client)
	if err != nil {
		t.Errorf("Valid unregister should not return an error: %s", err.Error())
	}

	serverExitChannel <- true
	<-serverExitChannel
	time.Sleep(100 * time.Millisecond)
}

func TestRun(t *testing.T) {
	serverExitChannel, _, _ := createServer(t)

	clientExitChannel := make(chan bool)
	clientOutputChannel := make(chan configuration_loader.Action)

	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	go Run(clientConfig, clientExitChannel, clientOutputChannel, nil)

	time.Sleep(1 * time.Second)

	clientExitChannel <- true
	<-clientExitChannel
	serverExitChannel <- true
	<-serverExitChannel
}
