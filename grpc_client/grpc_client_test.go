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

func createServer(t *testing.T) (chan bool, chan configuration_loader.Action) {
	var serverConfig configuration_loader.InitialConfiguration
	serverConfig.ServerConfiguration = &configuration_loader.ServerConfiguration{}
	serverConfig.ServerConfiguration.GRPCServerPort = 8080
	serverConfig.PinsActive = append(serverConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	serverExitChannel := make(chan bool)
	outputChannel := make(chan configuration_loader.Action)
	err := grpc_server.SetupAndRun(serverConfig, outputChannel, serverExitChannel)
	if err != nil {
		t.Errorf("Setting up the server should not return an error, instead we got %s", err.Error())
	}
	return serverExitChannel, outputChannel
}

func TestConnectionToServer(t *testing.T) {
	serverExitChannel, _ := createServer(t)

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
	time.Sleep(100 * time.Millisecond)
}

func TestRegisterPinsToGRPCServer(t *testing.T) {
	serverExitChannel, _ := createServer(t)

	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	client, _, _ := connectToGrpcServer(clientConfig)
	err := registerPinsToGRPCServer(client, clientConfig)
	if err == nil {
		t.Error("Register with repeated pins should return an error")
	}

	clientConfig.PinsActive = []gpio_manager.PairNamePin{gpio_manager.PairNamePin{"pin2", 90}}
	err = registerPinsToGRPCServer(client, clientConfig)
	if err != nil {
		t.Error("Register with valid pins should not return an error")
	}

	err = unregisterPins(client)
	if err != nil {
		t.Errorf("Valid unregister should not return an error: %s", err.Error())
	}

	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestCheckForActions(t *testing.T) {
	serverExitChannel, serverInputChannel := createServer(t)

	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	clientConfig.PinsActive = append(clientConfig.PinsActive, gpio_manager.PairNamePin{"pin2", 90})
	client, _, _ := connectToGrpcServer(clientConfig)
	registerPinsToGRPCServer(client, clientConfig)

	clientOutputChannel := make(chan configuration_loader.Action)
	err := checkForActions(client, clientOutputChannel)
	if err != nil {
		t.Errorf("Check for actions should not return an error even if there are no actions to perform, instead it go %s", err.Error())
	}
	serverInputChannel <- configuration_loader.Action{"pin2", true}

	go checkForActions(client, clientOutputChannel)

	action := <-clientOutputChannel
	if action.Pin != "pin2" {
		t.Errorf("Action received should be \"pin2\", instead it is %s", action.Pin)
	} else if action.State != true {
		t.Error("Action state received should be \"true\"")
	}

	err = unregisterPins(client)
	if err != nil {
		t.Errorf("Valid unregister should not return an error: %s", err.Error())
	}

	serverExitChannel <- true
	time.Sleep(100 * time.Millisecond)
}

func TestRunClient(t *testing.T) {
	serverExitChannel, _ := createServer(t)

	clientExitChannel := make(chan bool)
	clientOutputChannel := make(chan configuration_loader.Action)

	var clientConfig configuration_loader.InitialConfiguration
	clientConfig.GRPCServerIp = "localhost:8080"
	go RunClient(clientConfig, clientExitChannel, clientOutputChannel)

	time.Sleep(1 * time.Second)

	serverExitChannel <- true
	clientExitChannel <- true
}
