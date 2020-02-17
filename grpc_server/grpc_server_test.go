package grpc_server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/message_generator"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/peer"
)

func TestWrongConfig(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	err := SetupAndRun(config, nil, nil, nil, nil)
	exitChannel := make(chan bool)
	assert.NotEqual(t, err, nil, "Empty config should return an error")
	config.ServerConfiguration = &configuration_loader.ServerConfiguration{}
	err = SetupAndRun(config, nil, nil, nil, exitChannel)
	assert.NotEqual(t, err, nil, "Empty server port config should return an error")
	config.ServerConfiguration.GRPCServerPort = -8080
	err = SetupAndRun(config, nil, nil, nil, exitChannel)
	assert.NotEqual(t, err, nil, "Negative server port config should return an error")
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	config.ServerConfiguration.GRPCServerPort = 8080
	err = SetupAndRun(config, nil, nil, nil, exitChannel)
	assert.Equal(t, err, nil, "Correct server config should not return an error")
	exitChannel <- true
	<-exitChannel
}

func TestRegisterToServer(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{
		clientsRegistered: make(map[net.Addr]clientRegisteredData),
		actionsToPerform:  make(map[net.Addr]chan configuration_loader.Action),
		programmedActions: make(map[net.Addr]chan message_generator.ProgrammedActionOperation),
	}
	message0 := messages_protocol.RegistrationMessage{}
	message0.PinsToHandle = []string{"pin1"}
	result, _ := server.RegisterToServer(ctx, &message0)
	assert.Equal(t, result.Result, messages_protocol.RegistrationStatusCodes_Ok, "Message with no repeated pins should return an ok code, instead it is %d", result.Result)
	assert.Equal(t, len(result.PinsRepeated), 0, "Message with no repeated pins should return an empty list of pins repeated")
	assert.Equal(t, len(server.clientsRegistered), 1, "After a successfull registration, the clients registered should contain one element, instead it contains %d", len(server.clientsRegistered))
	pins := server.clientsRegistered[conn.LocalAddr()]
	assert.Equal(t, len(pins.Pins), 1, "After a successfull registration, the pins registered for the client should contain one element, instead it contains %d", len(pins.Pins))

	assert.Equal(t, pins.Pins[0], "pin1", "After a successfull registration, the first pin registered should be \"pin1\", instead it is %s", pins.Pins[0])
	message1 := messages_protocol.RegistrationMessage{}
	message1.PinsToHandle = []string{"pin1"}
	result, _ = server.RegisterToServer(ctx, &message1)
	assert.Equal(t, result.Result, messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered, "Message with repeated pins should return an error code, instead it is %d", result.Result)
	assert.Equal(t, len(result.PinsRepeated), 1, "Message with repeated pins should return the name of the pins")
	assert.Equal(t, result.PinsRepeated[0], "pin1", "Message with repeated pins should return the name of the pins")
	assert.Equal(t, len(server.clientsRegistered), 1, "After a filed registration, the clients registered should contain one element, instead it contains %d", len(server.clientsRegistered))
}

func TestUnregisterToServer(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{clientsRegistered: map[net.Addr]clientRegisteredData{conn.LocalAddr(): clientRegisteredData{LastTimeConnected: time.Now(), Pins: []string{"pin1"}}}}
	assert.Equal(t, len(server.clientsRegistered), 1, "The server should contain an element initially, instead it contains %d", len(server.clientsRegistered))
	server.actionsToPerform = map[net.Addr]chan configuration_loader.Action{conn.LocalAddr(): make(chan configuration_loader.Action)}
	server.UnregisterToServer(ctx, &messages_protocol.Empty{})
	assert.Equal(t, len(server.clientsRegistered), 0, "After a successfull unregistering, the clients registered should contain zero elements, instead it contains %d", len(server.clientsRegistered))
	assert.Equal(t, len(server.actionsToPerform), 0, "After a successfull unregistering, the actions to perform from that client should contain zero elements, instead it contains %d", len(server.actionsToPerform))
	assert.Equal(t, len(server.programmedActions), 0, "After a successfull unregistering, the programmed actions in the client should contain zero elements, instead it contains %d", len(server.actionsToPerform))
}

func TestCheckForActions(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{clientsRegistered: make(map[net.Addr]clientRegisteredData), actionsToPerform: make(map[net.Addr]chan configuration_loader.Action)}
	server.actionsToPerform[conn.LocalAddr()] = make(chan configuration_loader.Action)
	go func() {
		server.actionsToPerform[conn.LocalAddr()] <- configuration_loader.Action{"pin1", false}
	}()
	actions, err := server.CheckForActions(ctx, &messages_protocol.Empty{})
	assert.Equal(t, err, nil, "Check for actions should not return an error")
	assert.Equal(t, len(actions.Actions), 1, "Check for actions should return 1 action")
	assert.Equal(t, len(actions.ProgrammedActionOperations), 0, "Check for actions should return 0 programmed actions")
	assert.Equal(t, actions.Actions[0].Pin, "pin1", "Element 0 from check for actions should be \"pin1\", instead it is %s", actions.Actions[0].Pin)
	assert.Equal(t, actions.Actions[0].State, false, "Element 0 state from check for actions should be false")
	assert.Equal(t, len(server.actionsToPerform[conn.LocalAddr()]), 0, "After receiving the actions to perform, they should be removed")
}

func TestClientDisconnection(t *testing.T) {
	conn := net.TCPConn{}
	server := rpiHomeServer{clientsRegistered: make(map[net.Addr]clientRegisteredData), actionsToPerform: make(map[net.Addr]chan configuration_loader.Action)}
	server.clientsRegistered[conn.LocalAddr()] = clientRegisteredData{
		LastTimeConnected: time.Now(),
		Pins:              []string{"pin1"},
	}
	pins := server.getPinsAndUpdateMap()
	assert.Equal(t, pins, "pin1 ", "When there is a client registered with a pin, it should be returned")
	time.Sleep(time.Second * 8)
	pins = server.getPinsAndUpdateMap()
	assert.Equal(t, pins, "", "After the timeout has passed, the string returned should be empty")
}

func TestGetProgrammedActions(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{
		clientsRegistered: make(map[net.Addr]clientRegisteredData),
		actionsToPerform:  make(map[net.Addr]chan configuration_loader.Action),
		programmedActions: make(map[net.Addr]chan message_generator.ProgrammedActionOperation),
	}
	server.actionsToPerform[conn.LocalAddr()] = make(chan configuration_loader.Action)
	server.programmedActions[conn.LocalAddr()] = make(chan message_generator.ProgrammedActionOperation)
	go func() {
		server.programmedActions[conn.LocalAddr()] <- message_generator.ProgrammedActionOperation{
			message_generator.ProgrammedAction{
				configuration_loader.ActionTime{
					configuration_loader.Action{"pin1", false},
					configuration_loader.MyTime(time.Now())},
				false,
			},
			message_generator.CREATE,
		}
	}()
	actions, err := server.CheckForActions(ctx, &messages_protocol.Empty{})
	assert.Equal(t, err, nil, "Check for actions should not return an error")
	assert.Equal(t, len(actions.Actions), 0, "Check for actions should return 0 actions")
	assert.Equal(t, len(actions.ProgrammedActionOperations), 1, "Check for actions should return 1 programmed action")
	assert.Equal(t, len(server.programmedActions[conn.LocalAddr()]), 0, "After receiving the actions to perform, they should be removed")
}
