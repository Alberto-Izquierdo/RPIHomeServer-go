package message_generator

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/configuration_loader"
	"github.com/stretchr/testify/require"
)

func TestActionTwoSecondsDelay(t *testing.T) {
	actionTimes := []configuration_loader.ActionTime{
		configuration_loader.ActionTime{configuration_loader.Action{"light", false}, configuration_loader.MyTime(time.Now().Add(time.Minute * -10))},
		configuration_loader.ActionTime{configuration_loader.Action{"light", true}, configuration_loader.MyTime(time.Now().Add(time.Second * 2))},
	}
	c := make(chan configuration_loader.Action)
	exitChan := make(chan bool)
	err := Run(actionTimes, c, exitChan)
	require.Nil(t, err)
	select {
	case nextAction := <-c:
		require.Equal(t, nextAction.Pin, "light", "Action pin should be light, instead it is: %s", nextAction.Pin)
		require.Equal(t, nextAction.State, true, "Pin value should be true")
		exitChan <- true
		<-exitChan
	case _ = <-exitChan:
		t.Errorf("Something terrible happened")
	}
}
