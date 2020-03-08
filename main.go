package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/grpc_server"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/rpi_client"
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

	// RPI client (gRPC, message_generator and GPIO manager)
	exitChannels = append(exitChannels, make(chan bool))
	err = rpi_client.SetupAndRun(config, exitChannels[len(exitChannels)-1])
	if err != nil {
		fmt.Println("RPI client configuration failed: %s", err.Error())
		exitChannels = exitChannels[:len(exitChannels)-1]
	}

	fmt.Println("Waiting for messages")
	<-mainExitChannel
	for index := range exitChannels {
		exitChannels[len(exitChannels)-1-index] <- true
		<-exitChannels[len(exitChannels)-1-index]
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
