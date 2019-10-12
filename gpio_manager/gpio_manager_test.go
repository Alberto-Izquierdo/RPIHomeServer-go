package gpio_manager

import "testing"

func TestGpioManager(t *testing.T) {
	if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	pins := []PairNamePin{PairNamePin{"test", 18}}
	Setup(pins)
	if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	TurnPinOn("test")
	if pinState := GetPinState("test"); pinState != true {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, true)
	}
	TurnPinOff("test")
	if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	ClearAllPins()
}
