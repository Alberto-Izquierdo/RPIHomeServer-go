package gpio_manager

import "testing"

func TestGpioManager(t *testing.T) {
	if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	pins := []PairNamePin{PairNamePin{"test", 18}}
	err := Setup(pins)
	if err != nil {
		t.Errorf("Setup error: %s", err)
	}
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

func TestGpioManagerEmptyPins(t *testing.T) {
	pins := []PairNamePin{}
	err := Setup(pins)
	if err == nil {
		t.Errorf("Setup with empty pins should have failed")
	}
}
