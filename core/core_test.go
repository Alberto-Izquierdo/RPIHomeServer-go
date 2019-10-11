package core

import (
	"testing"
)

func TestGpioManager(t *testing.T) {
	pinState := GetPinState(18)
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", 18, pinState, false)
	}
	Setup()
	pinState = GetPinState(18)
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", 18, pinState, false)
	}
	TurnPinOn(18)
	pinState = GetPinState(18)
	if pinState != true {
		t.Errorf("GetPinState(%v) == %v, want %v", 18, pinState, true)
	}
	TurnPinOff(18)
	pinState = GetPinState(18)
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", 18, pinState, false)
	}
	ClearAllPins()
}
