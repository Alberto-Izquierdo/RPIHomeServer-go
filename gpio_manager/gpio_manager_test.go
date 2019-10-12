package gpio_manager

import "testing"

func TestGpioManager(t *testing.T) {
	pinState := GetPinState("test")
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	pins := []PairNamePin{PairNamePin{"test", 18}}
	Setup(pins)
	pinState = GetPinState("test")
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	TurnPinOn("test")
	pinState = GetPinState("test")
	if pinState != true {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, true)
	}
	TurnPinOff("test")
	pinState = GetPinState("test")
	if pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	ClearAllPins()
}
