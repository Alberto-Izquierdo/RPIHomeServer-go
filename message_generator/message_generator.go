package message_generator

import (
	"errors"
	"fmt"
	"time"

	ordered_queue "github.com/Alberto-Izquierdo/GoOrderedQueue"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
)

func Run(actions []types.ProgrammedAction, inputChannel chan types.ProgrammedActionOperation, outputChannel chan types.TelegramMessage, exitChannel chan bool) error {
	queue := ordered_queue.OrderedQueue{}
	err := initQueue(actions, &queue)
	if err != nil {
		fmt.Println("Error while creating the module: " + err.Error())
	}
	go func() {
		for {
			nextActionValid := true
			nextElement, err := queue.Pop()
			now := time.Now()
			t := now.Add(time.Hour * 10000)
			var nextAction types.ProgrammedAction

			if err != nil {
				fmt.Println("[message_generator]: Error when trying to pop elements from actions queue: ", err.Error(), "")
				nextActionValid = false
			} else {
				nextAction = nextElement.(types.ProgrammedAction)
				t = time.Time(nextAction.Time)
			}

			select {
			case _ = <-exitChannel:
				fmt.Println("[message_generator] Exit signal received, exiting...")
				return
			case operation := <-inputChannel:
				response, addPreviousAction := handleOperation(operation, &queue, nextAction, nextActionValid)
				outputChannel <- response
				if nextActionValid == true && addPreviousAction == true {
					queue.Push(nextAction)
				}
			case <-time.After(t.Sub(now)):
				handleNextAction(&nextAction, &queue, exitChannel)
			}
		}
	}()
	return nil
}

func initQueue(actions []types.ProgrammedAction, queue *ordered_queue.OrderedQueue) error {
	if len(actions) == 0 {
		return errors.New("No actions to launch")
	}
	for _, programmedAction := range actions {
		currTime := time.Time(programmedAction.Time)
		now := time.Now()
		date := time.Date(now.Year(), now.Month(), now.Day(), currTime.Hour(), currTime.Minute(), currTime.Second(), 0, now.Location())
		for date.Before(now) {
			date = date.Add(time.Hour * 24)
		}
		programmedAction.Time = types.MyTime(date)
		err := queue.Push(types.ProgrammedAction{Action: programmedAction.Action, Repeat: programmedAction.Repeat, Time: types.MyTime(date)})
		if err != nil {
			return errors.New("[message_generator]: Could not push elements into the queue: " + err.Error())
		}
	}
	return nil
}

func handleNextAction(nextAction *types.ProgrammedAction, queue *ordered_queue.OrderedQueue, exitChannel chan bool) {
	// Enqueue the action to the gpio manager
	gpio_manager.HandleAction(nextAction.Action)
	// Push the action again but with the time increased 24 hours
	if nextAction.Repeat == true {
		newAction := types.ProgrammedAction{
			Action: nextAction.Action,
			Repeat: true,
			Time:   types.MyTime(time.Time(nextAction.Time).Add(time.Hour * 24)),
		}
		err := queue.Push(newAction)
		if err != nil {
			fmt.Println("[message_generator]: Could not push elements into the queue: ", err.Error())
			exitChannel <- true
			return
		}
	}
}

func handleOperation(operation types.ProgrammedActionOperation, queue *ordered_queue.OrderedQueue, nextAction types.ProgrammedAction, nextActionValid bool) (response types.TelegramMessage, addPreviousAction bool) {
	addPreviousAction = true
	programmedAction := operation.ProgrammedAction
	currTime := time.Time(programmedAction.Time)
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), currTime.Hour(), currTime.Minute(), currTime.Second(), 0, now.Location())
	programmedAction.Time = types.MyTime(date)
	for time.Time(programmedAction.Time).Before(time.Now()) {
		programmedAction.Time = types.MyTime(time.Time(programmedAction.Time).Add(24 * time.Hour))
	}
	switch operation.Operation {
	case types.CREATE:
		err := queue.Push(programmedAction)
		if err == nil {
			response = types.TelegramMessage{Message: "Programmed action added", ChatId: programmedAction.Action.ChatId}
		} else {
			response = types.TelegramMessage{Message: "Error while trying to add the new programmed action" + err.Error(), ChatId: programmedAction.Action.ChatId}
		}
	case types.REMOVE:
		fmt.Println("action received")
		fmt.Println(types.ProgrammedActionToString(programmedAction))
		fmt.Println("first action")
		fmt.Println(types.ProgrammedActionToString(nextAction))

		if nextActionValid == true && programmedAction.Equals(nextAction) {
			response = types.TelegramMessage{Message: "Programmed action removed", ChatId: programmedAction.Action.ChatId}
			addPreviousAction = false
		} else {
			removed, err := queue.RemoveElement(programmedAction)
			if removed {
				response = types.TelegramMessage{Message: "Programmed action removed", ChatId: programmedAction.Action.ChatId}
			} else {
				response = types.TelegramMessage{Message: "Error while trying to remove the new programmed action: " + err.Error(), ChatId: programmedAction.Action.ChatId}
			}
		}
	default:
		response = types.TelegramMessage{Message: "Operation not known", ChatId: programmedAction.Action.ChatId}
	}
	return response, addPreviousAction
}
