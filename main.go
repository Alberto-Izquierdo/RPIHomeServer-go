package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc"
	messages_protocol "github.com/Alberto-Izquierdo/RPIHomeServer-go/messages"
)

var exit = false

func main() {
	setupKeyboardSignal()
	config, err := loadConfiguration()
	if err != nil {
		fmt.Println("There was an error parsing the configuration file: ", err)
		return
	}
	err = gpio_manager.Setup(config.PinsActive)
	if err != nil {
		fmt.Println("There was an error setting up the GPIO manager: ", err)
		return
	}
	defer gpio_manager.ClearAllPins()
	fmt.Println("Configuration loaded, connecting to gRPC server")
	client, connection, err := grpc.ConnectToGrpcServer(config)
	if err != nil {
		fmt.Println("There was an error connecting to the gRPC server: ", err)
		return
	}
	defer connection.Close()
	err = grpc.RegisterPinsToGRPCServer(client, config)
	if err != nil {
		fmt.Println("There was an error with the gRPC registration: ", err)
		return
	}
	fmt.Println("Waiting for messages")
	for !exit {
		ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
		defer cancel()
		actions, err := client.CheckForActions(ctx, &messages_protocol.Empty{})
		if err != nil {
			fmt.Println("There was an error receiving actions to complete: ", err)
			return
		}
		for _, action := range actions.Actions {
			gpio_manager.SetPinState(*action.Pin, *action.State)
		}
	}
	fmt.Println("Done!")
}

func loadConfiguration() (config configuration_loader.InitialConfiguration, err error) {
	filepath := flag.String("path", "", "path to the configuration file")
	flag.Parse()
	if *filepath == "" {
		return config, errors.New("\"path\" argument can not be empty")
	}
	fmt.Println("Reading configuration from file")
	return configuration_loader.LoadConfigurationFromPath(*filepath)
}

func setupKeyboardSignal() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		fmt.Println("Ctrl+C captured, cleaning up")
		exit = true
	}()
}
