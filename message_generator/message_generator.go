package message_generator

import (
	"errors"
	"fmt"
	"time"

	ordered_queue "github.com/Alberto-Izquierdo/GoOrderedQueue"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
)

const (
	CREATE = iota
	REMOVE
)

type ProgrammedActionOperation struct {
	ProgrammedAction ProgrammedAction
	Operation        int32
}

type ProgrammedAction struct {
	Action configuration_loader.ActionTime
	Repeat bool
}

func (this ProgrammedAction) LessThan(other interface{}) bool {
	return time.Time(this.Action.Time).Before(time.Time(other.(ProgrammedAction).Action.Time))
}

func Run(actions []configuration_loader.ActionTime, outputChannel chan configuration_loader.Action, exitChannel chan bool) error {
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
			var nextAction ProgrammedAction

			if err != nil {
				fmt.Println("[message_generator]: Error when trying to pop elements from actions queue: ", err.Error(), "")
			} else {
				nextAction = nextElement.(ProgrammedAction)
				t = time.Time(nextAction.Action.Time)
			}

			select {
			case _ = <-exitChannel:
				fmt.Println("[message_generator] Exit signal received, exiting...")
				exitChannel <- true
				return
			case <-time.After(t.Sub(now)):
				handleNextAction(&nextAction, outputChannel, &queue, exitChannel)
			}
		}
	}()
	return nil
}

func initQueue(actions []configuration_loader.ActionTime, queue *ordered_queue.OrderedQueue) error {
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
		err := queue.Push(ProgrammedAction{actionTime, true})
		if err != nil {
			return errors.New("[message_generator]: Could not push elements into the queue: " + err.Error())
		}
	}
	return nil
}

func handleNextAction(nextAction *ProgrammedAction, outputChannel chan configuration_loader.Action, queue *ordered_queue.OrderedQueue, exitChannel chan bool) {
	// Enqueue the action to the gpio manager
	outputChannel <- nextAction.Action.Action
	// Push the action again but with the time increased 24 hours
	if nextAction.Repeat {
		newTime := time.Now().Add(time.Hour * 24)
		newAction := ProgrammedAction{
			configuration_loader.ActionTime{
				nextAction.Action.Action,
				configuration_loader.MyTime(newTime)},
			true}
		err := queue.Push(newAction)
		if err != nil {
			fmt.Println("[message_generator]: Could not push elements into the queue: ", err.Error())
			exitChannel <- true
			return
		}
	}
}
