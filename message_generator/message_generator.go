package message_generator

import (
	"errors"
	"fmt"
	"time"

	ordered_queue "github.com/Alberto-Izquierdo/GoOrderedQueue"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
)

const (
	CREATE = iota
	REMOVE
	GET_ACTIONS
)

func Run(actions []types.ProgrammedAction, exitChannel chan bool) error {
	queue := ordered_queue.OrderedQueue{}
	err := initQueue(actions, &queue)
	if err != nil {
		fmt.Println(err.Error())
		return err
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
	if nextAction.Repeat {
		newTime := time.Now().Add(time.Hour * 24)
		newAction := types.ProgrammedAction{
			Action: nextAction.Action,
			Repeat: true,
			Time:   types.MyTime(newTime),
		}
		err := queue.Push(newAction)
		if err != nil {
			fmt.Println("[message_generator]: Could not push elements into the queue: ", err.Error())
			exitChannel <- true
			return
		}
	}
}
