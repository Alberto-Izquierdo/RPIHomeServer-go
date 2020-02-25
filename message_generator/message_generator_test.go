package message_generator

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/stretchr/testify/require"
)

func TestActionTwoSecondsDelay(t *testing.T) {
	programmedActions := []types.ProgrammedAction{
		types.ProgrammedAction{Action: types.Action{"light", false}, Time: types.MyTime(time.Now().Add(time.Minute * -10)), Repeat: true},
		types.ProgrammedAction{Action: types.Action{"light", true}, Time: types.MyTime(time.Now().Add(time.Second * 2)), Repeat: true},
	}
	c := make(chan types.Action)
	exitChan := make(chan bool)
	err := Run(programmedActions, c, exitChan)
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
