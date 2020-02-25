package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_client"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/message_generator"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/telegram_bot"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
)

var mainExitChannel = make(chan bool)

func main() {
	setupKeyboardSignal()
	config, err := loadConfiguration()
	if err != nil {
		fmt.Println("There was an error parsing the configuration file: ", err.Error())
		return
	}
	err = gpio_manager.Setup(config.PinsActive)
	defer gpio_manager.ClearAllPins()
	if err != nil {
		if err.Error() == gpio_manager.EmptyPins {
			fmt.Println("There are not any pins active, the program will act just as server")
		} else {
			fmt.Println("There was an error setting up the GPIO manager: ", err.Error())
			return
		}
	}
	// general variables
	var exitChannels []chan bool

	// gRPC server and telegram bot
	if config.ServerConfiguration != nil {
		tgGrpcActionsChannel := make(chan types.Action)
		tgGrpcResponsesChannel := make(chan string)
		exitChannels = append(exitChannels, make(chan bool))
		err = telegram_bot.LaunchTelegramBot(config, tgGrpcActionsChannel, tgGrpcResponsesChannel, exitChannels[len(exitChannels)-1])
		if err != nil {
			fmt.Println("Error while setting up telegram bot: " + err.Error())
			return
		}
		exitChannels = append(exitChannels, make(chan bool))
		err = grpc_server.SetupAndRun(config, tgGrpcActionsChannel, tgGrpcResponsesChannel, nil, exitChannels[len(exitChannels)-1])
		if err != nil {
			fmt.Println("Error while setting up gRPC server: " + err.Error())
			return
		}
	}

	// gRPC client
	gRPCClientActionsChannel := make(chan types.Action)
	exitChannels = append(exitChannels, make(chan bool))
	err = grpc_client.Run(config, exitChannels[len(exitChannels)-1], gRPCClientActionsChannel, nil, mainExitChannel)
	if err != nil {
		exitChannels = exitChannels[:len(exitChannels)-1]
		if err.Error() == grpc_client.EmptyPinsMessage || config.ServerConfiguration != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Error while setting up gRPC client: " + err.Error())
			return
		}
	} else {
	}
	//TODO: gRPCClientMessagesChannel := make(chan string)

	if len(config.AutomaticMessages) > 0 {
		exitChannels = append(exitChannels, make(chan bool))
		err = message_generator.Run(config.AutomaticMessages, gRPCClientActionsChannel, exitChannels[len(exitChannels)-1])
		if err != nil {
			fmt.Println("Error while setting up message generator: " + err.Error())
			return
		}
	}
	fmt.Println("Waiting for messages")
	var exit = false
	for !exit {
		select {
		case action := <-gRPCClientActionsChannel:
			gpio_manager.SetPinState(action.Pin, action.State)
		case exit = <-mainExitChannel:
			for index := range exitChannels {
				exitChannels[len(exitChannels)-1-index] <- true
			}
			for _, exitChannel := range exitChannels {
				<-exitChannel
			}
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
		mainExitChannel <- true
	}()
}
