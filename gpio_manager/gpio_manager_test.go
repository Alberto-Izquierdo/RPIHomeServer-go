package gpio_manager

import (
	"testing"

	"github.com/Alberto-Izquierdo/RPIHomeServer-go/types"
	"github.com/stretchr/testify/assert"
)

func TestGpioManager(t *testing.T) {
	defer ClearAllPins()
	pinState := GetPinState("test")
	assert.Equal(t, pinState, false, "GetPinState(%v) == %v, want %v", "test", pinState, false)
	pins := []types.PairNamePin{types.PairNamePin{"test", 18}}
	err := Setup(pins)
	assert.Equal(t, err, nil, "Setup error: %s", err)
	pinState = GetPinState("test")
	assert.Equal(t, pinState, false, "GetPinState(%v) == %v, want %v", "test", pinState, false)
	stateChanged, err := TurnPinOn("test")
	assert.Equal(t, err, nil, "TurnPinOn(%v) should not return an error", "test")
	assert.True(t, stateChanged, "TurnPinOn(%v) should have changed the state", "test")
	pinState = GetPinState("test")
	assert.Equal(t, pinState, true, "GetPinState(%v) == %v, want %v", "test", pinState, true)
	stateChanged, err = TurnPinOff("test")
	assert.Equal(t, err, nil, "TurnPinOff(%v) should not return an error", "test")
	assert.True(t, stateChanged, "TurnPinOff(%v) should have changed the state", "test")
	pinState = GetPinState("test")
	assert.Equal(t, pinState, false, "GetPinState(%v) == %v, want %v", "test", pinState, false)
	stateChanged, err = HandleAction(types.Action{"test", true})
	assert.Equal(t, err, nil, "HandlePinAction(types.Action{%v, true}) should not return an error", "test")
	assert.True(t, stateChanged, "HandlePinAction(types.Action{%v, true}) should have changed the state", "test")
	pinState = GetPinState("test")
	assert.Equal(t, pinState, true, "GetPinState(%v) == %v, want %v", "test", pinState, true)
	stateChanged, err = HandleAction(types.Action{"test", false})
	assert.Equal(t, err, nil, "HandleAction(types.Action{%v, false}) should not return an error", "test")
	assert.True(t, stateChanged, "HandleAction(types.Action{%v, false}) should have changed the state", "test")
	pinState = GetPinState("test")
	assert.Equal(t, pinState, false, "GetPinState(%v) == %v, want %v", "test", pinState, false)
}

func TestGpioManagerEmptyPins(t *testing.T) {
	pins := []types.PairNamePin{}
	err := Setup(pins)
	assert.NotEqual(t, err, nil, "Setup with empty pins should have failed")
	pinsActive := GetPinsAvailable()
	assert.Equal(t, len(pinsActive), 0, "Error, the lenght of pins active should be 0 after an error, instead it is %d", len(pinsActive))
}

func TestWrongNamePin(t *testing.T) {
	pins := []types.PairNamePin{types.PairNamePin{"GetPinsAvailable", 18}, types.PairNamePin{"test2", 11}}
	err := Setup(pins)
	assert.NotEqual(t, err, nil, "Pin with name \"GetPinsAvailable\" should return an error")
	pinsActive := GetPinsAvailable()
	assert.Equal(t, len(pinsActive), 0, "Error, the lenght of pins active should be 0 after an error, instead it is %d", len(pinsActive))
}

func TestGetPinsAvailable(t *testing.T) {
	pins := []types.PairNamePin{types.PairNamePin{"test", 18}, types.PairNamePin{"test2", 11}}
	err := Setup(pins)
	assert.Equal(t, err, nil, "Setup error: %s", err)
	pinsActive := GetPinsAvailable()
	assert.Equal(t, len(pinsActive), 2, "Error, the lenght of pins active should be 2, instead it is %d", len(pinsActive))
	if pinsActive[0] != "test" {
		assert.Equal(t, pinsActive[0], "test2", "Error, the pin should be \"test\" or \"test2\" and it is \"%s\"", pinsActive[0])
		assert.Equal(t, pinsActive[1], "test", "Error, the pin should be \"test\" and it is \"%s\"", pinsActive[1])
	} else {
		assert.Equal(t, pinsActive[1], "test2", "Error, the pin should be \"test2\" and it is \"%s\"", pinsActive[1])
	}
}
