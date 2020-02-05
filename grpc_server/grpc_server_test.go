package grpc_server

import (
	"context"
	"net"
	"testing"
	"time"

	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
	"google.golang.org/grpc/peer"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
)

func TestWrongConfig(t *testing.T) {
	var config configuration_loader.InitialConfiguration
	err := SetupAndRun(config, nil, nil, nil)
	exitChannel := make(chan bool)
	if err == nil {
		t.Errorf("Empty config should return an error")
	}
	config.ServerConfiguration = &configuration_loader.ServerConfiguration{}
	err = SetupAndRun(config, nil, nil, exitChannel)
	if err == nil {
		t.Errorf("Empty server port config should return an error")
	}
	config.ServerConfiguration.GRPCServerPort = -8080
	err = SetupAndRun(config, nil, nil, exitChannel)
	if err == nil {
		t.Errorf("Negative server port config should return an error")
	}
	config.PinsActive = append(config.PinsActive, gpio_manager.PairNamePin{"pin1", 90})
	config.ServerConfiguration.GRPCServerPort = 8080
	err = SetupAndRun(config, nil, nil, exitChannel)
	if err != nil {
		t.Errorf("Correct server config should not return an error")
	}
	exitChannel <- true
}

func TestRegisterToServer(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{nil, make(map[net.Addr]dateStringsPair), make(map[net.Addr]chan configuration_loader.Action)}
	message0 := messages_protocol.RegistrationMessage{}
	message0.PinsToHandle = []string{"pin1"}
	result, _ := server.RegisterToServer(ctx, &message0)
	if *result.Result != messages_protocol.RegistrationStatusCodes_Ok {
		t.Errorf("Message with no repeated pins should return an ok code, instead it is %d", *result.Result)
	} else if len(result.PinsRepeated) != 0 {
		t.Errorf("Message with no repeated pins should return an empty list of pins repeated")
	} else if len(server.clientsRegistered) != 1 {
		t.Errorf("After a successfull registration, the clients registered should contain one element, instead it contains %d", len(server.clientsRegistered))
	} else {
		pins := server.clientsRegistered[conn.LocalAddr()]
		if len(pins.Pins) != 1 {
			t.Errorf("After a successfull registration, the pins registered for the client should contain one element, instead it contains %d", len(pins.Pins))
		} else if pins.Pins[0] != "pin1" {
			t.Errorf("After a successfull registration, the first pin registered should be \"pin1\", instead it is %s", pins.Pins[0])
		}
	}
	message1 := messages_protocol.RegistrationMessage{}
	message1.PinsToHandle = []string{"pin1"}
	result, _ = server.RegisterToServer(ctx, &message1)
	if *result.Result != messages_protocol.RegistrationStatusCodes_PinNameAlreadyRegistered {
		t.Errorf("Message with repeated pins should return an error code, instead it is %d", *result.Result)
	} else if len(result.PinsRepeated) != 1 || result.PinsRepeated[0] != "pin1" {
		t.Errorf("Message with repeated pins should return the name of the pins")
	} else if len(server.clientsRegistered) != 1 {
		t.Errorf("After a filed registration, the clients registered should contain one element, instead it contains %d", len(server.clientsRegistered))
	}
}

func TestUnregisterToServer(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{nil, map[net.Addr]dateStringsPair{conn.LocalAddr(): dateStringsPair{time.Now(), []string{"pin1"}}}, nil}
	if len(server.clientsRegistered) != 1 {
		t.Errorf("The server should contain an element initially, instead it contains %d", len(server.clientsRegistered))
	}
	server.actionsToPerform = map[net.Addr]chan configuration_loader.Action{conn.LocalAddr(): make(chan configuration_loader.Action)}
	server.UnregisterToServer(ctx, &messages_protocol.Empty{})
	if len(server.clientsRegistered) != 0 {
		t.Errorf("After a successfull unregistering, the clients registered should contain zero elements, instead it contains %d", len(server.clientsRegistered))
	} else if len(server.actionsToPerform) != 0 {
		t.Errorf("After a successfull unregistering, the actions to perform from that client should contain zero elements, instead it contains %d", len(server.actionsToPerform))
	}
}

func TestCheckForActions(t *testing.T) {
	conn := net.TCPConn{}
	p := peer.Peer{conn.LocalAddr(), nil}
	ctx := peer.NewContext(context.TODO(), &p)
	server := rpiHomeServer{nil, make(map[net.Addr]dateStringsPair), make(map[net.Addr]chan configuration_loader.Action)}
	server.actionsToPerform[conn.LocalAddr()] = make(chan configuration_loader.Action)
	go func() {
		server.actionsToPerform[conn.LocalAddr()] <- configuration_loader.Action{"pin1", false}
	}()
	actions, err := server.CheckForActions(ctx, &messages_protocol.Empty{})
	if err != nil {
		t.Errorf("Check for actions should not return an error")
	} else if len(actions.Actions) != 1 {
		t.Errorf("Check for actions should return 1 action")
	} else if *actions.Actions[0].Pin != "pin1" {
		t.Errorf("Element 0 from check for actions should be \"pin1\", instead it is %s", *actions.Actions[0].Pin)
	} else if *actions.Actions[0].State != false {
		t.Errorf("Element 0 state from check for actions should be false")
	} else if len(server.actionsToPerform[conn.LocalAddr()]) != 0 {
		t.Errorf("After receiving the actions to perform, they should be removed")
	}
}

func TestClientDisconnection(t *testing.T) {
	conn := net.TCPConn{}
	server := rpiHomeServer{nil, make(map[net.Addr]dateStringsPair), make(map[net.Addr]chan configuration_loader.Action)}
	server.clientsRegistered[conn.LocalAddr()] = dateStringsPair{time.Now(), []string{"pin1"}}
	pins := server.getPinsAndUpdateMap()
	if pins != "pin1 " {
		t.Errorf("When there is a client registered with a pin, it should be returned")
	}
	time.Sleep(time.Second * 8)
	pins = server.getPinsAndUpdateMap()
	if pins != "" {
		t.Errorf("After the timeout has passed, the string returned should be empty")
	}
}
