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

func TestCreateProgrammedAction(t *testing.T) {
	programmedActions := []types.ProgrammedAction{}
	gpio_manager.Setup([]types.PairNamePin{types.PairNamePin{Name: "light", Pin: 2}})
	assert.False(t, gpio_manager.GetPinState("light"))
	exitChan := make(chan bool)
	telegramChannel := make(chan types.TelegramMessage)
	programmedActionOperationsChannel := make(chan types.ProgrammedActionOperation)
	err := Run(programmedActions, programmedActionOperationsChannel, telegramChannel, exitChan)
	assert.Nil(t, err)
	actionTime := types.MyTime(time.Now().Add(time.Second * 2))
	programmedActionOperationsChannel <- types.ProgrammedActionOperation{
		Operation: types.CREATE,
		ProgrammedAction: types.ProgrammedAction{
			Action: types.Action{
				Pin:    "light",
				State:  true,
				ChatId: 123,
			},
			Repeat: true,
			Time:   actionTime,
		},
	}
	select {
	case response := <-telegramChannel:
		assert.Equal(t, response.Message, "Programmed action added")
		assert.Equal(t, response.ChatId, int64(123))
	case _ = <-exitChan:
		t.Errorf("Something terrible happened")
	}
	time.Sleep(3 * time.Second)
	assert.True(t, gpio_manager.GetPinState("light"))

	actionTime = types.MyTime(time.Time(actionTime).Add(time.Hour * 24))

	programmedActionOperationsChannel <- types.ProgrammedActionOperation{
		Operation: types.REMOVE,
		ProgrammedAction: types.ProgrammedAction{
			Action: types.Action{
				Pin:    "light",
				State:  true,
				ChatId: 123,
			},
			Repeat: true,
			Time:   actionTime,
		},
	}
	select {
	case response := <-telegramChannel:
		assert.Equal(t, response.Message, "Programmed action removed")
		assert.Equal(t, response.ChatId, int64(123))
	case _ = <-exitChan:
		t.Errorf("Something terrible happened")
	}

	exitChan <- true
	time.Sleep(100 * time.Millisecond)
}
