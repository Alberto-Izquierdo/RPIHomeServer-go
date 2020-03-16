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
			nextElement, err := queue.Pop()
			now := time.Now()
			t := now.Add(time.Hour * 10000)
			var nextAction types.ProgrammedAction

			if err != nil {
				fmt.Println("[message_generator]: Error when trying to pop elements from actions queue: ", err.Error(), "")
			} else {
				nextAction = nextElement.(types.ProgrammedAction)
				t = time.Time(nextAction.Time)
			}

			select {
			case _ = <-exitChannel:
				fmt.Println("[message_generator] Exit signal received, exiting...")
				return
			case operation := <-inputChannel:
				outputChannel <- handleOperation(operation, &queue, nextAction)
				if err == nil {
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

func handleOperation(operation types.ProgrammedActionOperation, queue *ordered_queue.OrderedQueue, nextAction types.ProgrammedAction) types.TelegramMessage {
	switch operation.Operation {
	case types.CREATE:
		err := queue.Push(operation.ProgrammedAction)
		if err == nil {
			return types.TelegramMessage{Message: "Programmed action added", ChatId: operation.ProgrammedAction.Action.ChatId}
		}
		return types.TelegramMessage{Message: "Error while trying to add the new programmed action" + err.Error(), ChatId: operation.ProgrammedAction.Action.ChatId}
	case types.REMOVE:
		if operation.ProgrammedAction.Equals(nextAction) {
			return types.TelegramMessage{Message: "Programmed action removed", ChatId: operation.ProgrammedAction.Action.ChatId}
		}
		removed, err := queue.RemoveElement(operation.ProgrammedAction)
		if removed {
			return types.TelegramMessage{Message: "Programmed action removed", ChatId: operation.ProgrammedAction.Action.ChatId}
		}
		return types.TelegramMessage{Message: "Error while trying to remove the new programmed action: " + err.Error(), ChatId: operation.ProgrammedAction.Action.ChatId}
	case types.GET_ACTIONS:
		// TODO:
	}
	return types.TelegramMessage{Message: "Operation not known", ChatId: operation.ProgrammedAction.Action.ChatId}
}
