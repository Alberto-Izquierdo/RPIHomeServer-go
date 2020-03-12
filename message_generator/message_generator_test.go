package message_generator

import (
	"testing"
	"time"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/gpio_manager"
	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionTwoSecondsDelay(t *testing.T) {
	programmedActions := []types.ProgrammedAction{
		types.ProgrammedAction{Action: types.Action{"light", false, 0}, Time: types.MyTime(time.Now().Add(time.Minute * -10)), Repeat: true},
		types.ProgrammedAction{Action: types.Action{"light", true, 0}, Time: types.MyTime(time.Now().Add(time.Second * 2)), Repeat: true},
	}
	gpio_manager.Setup([]types.PairNamePin{types.PairNamePin{Name: "light", Pin: 2}})
	assert.False(t, gpio_manager.GetPinState("light"))
	exitChan := make(chan bool)
	telegramChannel := make(chan types.TelegramMessage)
	programmedActionOperationsChannel := make(chan types.ProgrammedActionOperation)
	err := Run(programmedActions, programmedActionOperationsChannel, telegramChannel, exitChan)
	require.Nil(t, err)
	select {
	case _ = <-exitChan:
		t.Errorf("Something terrible happened")
	case <-time.After(time.Second * 3):
		assert.True(t, gpio_manager.GetPinState("light"))
	}
	exitChan <- true
	time.Sleep(100 * time.Millisecond)
}
