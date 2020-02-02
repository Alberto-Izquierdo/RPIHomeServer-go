package message_generator

import (
	"errors"
	"fmt"
	"time"

	ordered_queue "github.com/Alberto-Izquierdo/GoOrderedQueue"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
)

func Run(actions []configuration_loader.ActionTime, outputChannel chan configuration_loader.Action, exitChannel chan bool) error {
	queue := ordered_queue.OrderedQueue{}
	if len(actions) == 0 {
		return errors.New("No actions to launch")
	}
	for _, actionTime := range actions {
		currTime := time.Time(actionTime.Time)
		now := time.Now()
		date := time.Date(now.Year(), now.Month(), now.Day(), currTime.Hour(), currTime.Minute(), currTime.Second(), 0, now.Location())
		for date.Before(now) {
			date = date.Add(time.Hour * 24)
		}
		actionTime.Time = configuration_loader.MyTime(date)
		err := queue.Push(actionTime)
		if err != nil {
			return errors.New("[message_generator]: Could not push elements into the queue: " + err.Error())
		}
	}
	go func() {
		for {
			nextElement, err := queue.Pop()
			if err != nil {
				fmt.Println("[message_generator]: Error when trying to pop elements from actions queue: ", err.Error())
				exitChannel <- true
				return
			}
			nextAction := nextElement.(configuration_loader.ActionTime)
			t := time.Time(nextAction.Time)
			now := time.Now()

			select {
			case _ = <-exitChannel:
				fmt.Println("[message_generator] Exit signal received, exiting...")
				exitChannel <- true
				return
			case <-time.After(t.Sub(now)):
				// Enqueue the action to the gpio manager
				outputChannel <- nextAction.Action
				// Push the action again but with the time increased 24 hours
				newTime := now.Add(time.Hour * 24)
				newAction := configuration_loader.ActionTime{nextAction.Action, configuration_loader.MyTime(newTime)}
				err := queue.Push(newAction)
				if err != nil {
					fmt.Println("[message_generator]: Could not push elements into the queue: ", err.Error())
					exitChannel <- true
					return
				}
			}
		}
	}()
	return nil
}
