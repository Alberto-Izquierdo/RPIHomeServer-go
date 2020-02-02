package message_generator

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
)

func TestActionTwoSecondsDelay(t *testing.T) {
	actionTimes := []configuration_loader.ActionTime{
		configuration_loader.ActionTime{configuration_loader.Action{"light", false}, configuration_loader.MyTime(time.Now().Add(time.Minute * -10))},
		configuration_loader.ActionTime{configuration_loader.Action{"light", true}, configuration_loader.MyTime(time.Now().Add(time.Second * 2))},
	}
	c := make(chan configuration_loader.Action)
	exitChan := make(chan bool)
	err := Run(actionTimes, c, exitChan)
	if err != nil {
		t.Errorf("Correct message generator setup should not return an error, instead it returned" + err.Error())
	}
	select {
	case nextAction := <-c:
		if nextAction.Pin != "light" {
			t.Errorf("Action pin should be light, instead it is: %s", nextAction.Pin)
		} else if nextAction.State != true {
			t.Errorf("Pin value should be true")
		}
		exitChan <- true
		<-exitChan
	case _ = <-exitChan:
		t.Errorf("Something terrible happened")
	}
}
