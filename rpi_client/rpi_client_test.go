package rpi_client

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
	err := grpc_server.SetupAndRun(serverConfig, outputChannel, nil, responsesChannel, serverExitChannel)
	assert.Nil(t, err)
	return serverExitChannel, outputChannel, responsesChannel
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
