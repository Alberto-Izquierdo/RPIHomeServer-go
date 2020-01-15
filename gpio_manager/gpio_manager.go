package gpio_manager

import (
	"errors"
	"fmt"

	"github.com/stianeikeland/go-rpio"
)

type PairNamePin struct {
	Name string
	Pin  int
}

func Setup(pins []PairNamePin) (err error) {
	if len(pins) == 0 {
		err = errors.New("pins can not be empty")
		return err
	}
	manager.PinStates = make(map[string]*pinState)
	for _, pinName := range pins {
		if pinName.Name == "GetPinsAvailable" {
			err = errors.New("Pin's name should not be \"GetPinsAvailable\", change it in the configuration")
			ClearAllPins()
			return err
		}
		manager.PinStates[pinName.Name] = new(pinState)
		manager.PinStates[pinName.Name].state = false
		manager.PinStates[pinName.Name].pin = pinName.Pin
	}
	manager.gpioAvailable = true
	if err := rpio.Open(); err != nil {
		manager.gpioAvailable = false
		manager.clearAllPins()
		fmt.Println("Unable to open gpio, error:", err.Error())
		fmt.Println("The program will continue for testing purpouses")
	}
	return err
}

func SetPinState(pin string, state bool) (bool, error) {
	if state {
		return TurnPinOn(pin)
	} else {
		return TurnPinOff(pin)
	}
}

func TurnPinOn(pin string) (bool, error) {
	return manager.turnPinOn(pin)
}

func TurnPinOff(pin string) (bool, error) {
	return manager.turnPinOff(pin)
}

func ClearAllPins() {
	manager.clearAllPins()
	if manager.gpioAvailable {
		rpio.Close()
	}
	manager.PinStates = make(map[string]*pinState)
}

func GetPinState(pin string) bool {
	return manager.getPinState(pin)
}

var manager gpioManager

type pinState struct {
	pin   int
	state bool
}

type gpioManager struct {
	PinStates     map[string]*pinState
	gpioAvailable bool
}

func (m *gpioManager) turnPinOn(pin string) (stateChanged bool, err error) {
	err = nil
	stateChanged = false
	if v, ok := m.PinStates[pin]; !ok {
		err = errors.New("[gpio_manager]: Pin " + pin + " not set in the initial configuration")
	} else if !v.state {
		if m.gpioAvailable {
			rpio.Pin(v.pin).High()
		}
		if !m.PinStates[pin].state {
			stateChanged = true
			m.PinStates[pin].state = true
			fmt.Println("Pin ", pin, " turned on")
		}
	}
	return stateChanged, err
}

func (m *gpioManager) turnPinOff(pin string) (stateChanged bool, err error) {
	if v, ok := m.PinStates[pin]; !ok {
		err = errors.New("[gpio_manager]: Pin " + pin + " not set in the initial configuration")
	} else if v.state {
		if m.gpioAvailable {
			rpio.Pin(v.pin).Low()
		}
		if m.PinStates[pin].state {
			stateChanged = true
			m.PinStates[pin].state = false
			fmt.Println("Pin ", pin, " turned off")
		}
	}
	return stateChanged, err
}

func (m *gpioManager) clearAllPins() {
	for _, v := range m.PinStates {
		if m.gpioAvailable {
			rpio.Pin(v.pin).Low()
		}
	}
}

func (m gpioManager) getPinState(pin string) bool {
	if v, ok := m.PinStates[pin]; ok {
		return v.state
	}
	return false
}

func GetPinsAvailable() []string {
	pins := make([]string, 0)
	for k, _ := range manager.PinStates {
		pins = append(pins, k)
	}
	return pins
}
