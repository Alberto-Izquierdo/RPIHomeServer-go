package gpio_manager

import "testing"

func TestGpioManager(t *testing.T) {
	defer ClearAllPins()
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
	stateChanged, err := TurnPinOn("test")
	if err != nil {
		t.Errorf("TurnPinOn(%v) should not return an error", "test")
	} else if !stateChanged {
		t.Errorf("TurnPinOn(%v) should have changed the state", "test")
	} else if pinState := GetPinState("test"); pinState != true {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, true)
	}
	stateChanged, err = TurnPinOff("test")
	if err != nil {
		t.Errorf("TurnPinOff(%v) should not return an error", "test")
	} else if !stateChanged {
		t.Errorf("TurnPinOff(%v) should have changed the state", "test")
	} else if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
	stateChanged, err = SetPinState("test", true)
	if err != nil {
		t.Errorf("SetPinState(%v, true) should not return an error", "test")
	} else if !stateChanged {
		t.Errorf("SetPinState(%v, true) should have changed the state", "test")
	} else if pinState := GetPinState("test"); pinState != true {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, true)
	}
	stateChanged, err = SetPinState("test", false)
	if err != nil {
		t.Errorf("SetPinState(%v, false) should not return an error", "test")
	} else if !stateChanged {
		t.Errorf("SetPinState(%v, false) should have changed the state", "test")
	} else if pinState := GetPinState("test"); pinState != false {
		t.Errorf("GetPinState(%v) == %v, want %v", "test", pinState, false)
	}
}

func TestGpioManagerEmptyPins(t *testing.T) {
	pins := []PairNamePin{}
	err := Setup(pins)
	if err == nil {
		t.Errorf("Setup with empty pins should have failed")
	}
	pinsActive := GetPinsAvailable()
	if len(pinsActive) != 0 {
		t.Errorf("Error, the lenght of pins active should be 0 after an error, instead it is %d", len(pinsActive))
	}
}

func TestWrongNamePin(t *testing.T) {
	pins := []PairNamePin{PairNamePin{"GetPinsAvailable", 18}, PairNamePin{"test2", 11}}
	err := Setup(pins)
	if err == nil {
		t.Errorf("Pin with name \"GetPinsAvailable\" should return an error")
	}
	pinsActive := GetPinsAvailable()
	if len(pinsActive) != 0 {
		t.Errorf("Error, the lenght of pins active should be 0 after an error, instead it is %d", len(pinsActive))
	}
}

func TestGetPinsAvailable(t *testing.T) {
	pins := []PairNamePin{PairNamePin{"test", 18}, PairNamePin{"test2", 11}}
	err := Setup(pins)
	if err != nil {
		t.Errorf("Setup error: %s", err)
	}
	pinsActive := GetPinsAvailable()
	if len(pinsActive) != 2 {
		t.Errorf("Error, the lenght of pins active should be 2, instead it is %d", len(pinsActive))
	}
	if pinsActive[0] != "test" {
		if pinsActive[0] != "test2" {
			t.Errorf("Error, the pin should be \"test\" or \"test2\" and it is \"%s\"", pinsActive[1])
		} else if pinsActive[1] != "test" {
			t.Errorf("Error, the pin should be \"test\" and it is \"%s\"", pinsActive[1])
		}
	} else {
		if pinsActive[1] != "test2" {
			t.Errorf("Error, the pin should be \"test2\" and it is \"%s\"", pinsActive[1])
		}
	}
}
