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
		manager.PinStates[pinName.Name] = new(pinState)
		manager.PinStates[pinName.Name].state = false
		manager.PinStates[pinName.Name].pin = pinName.Pin
	}
	manager.gpioAvailable = true
	if err := rpio.Open(); err != nil {
		manager.gpioAvailable = false
		fmt.Println("Unable to open gpio, error:", err.Error())
		fmt.Println("The program will continue for testing purpouses")
	}
	return err
}

func SetPinState(pin string, state bool) {
	if state {
		TurnPinOn(pin)
	} else {
		TurnPinOff(pin)
	}
}

func TurnPinOn(pin string) {
	manager.turnPinOn(pin)
}

func TurnPinOff(pin string) {
	manager.turnPinOff(pin)
}

func ClearAllPins() {
	manager.clearAllPins()
	if manager.gpioAvailable {
		rpio.Close()
	}
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

func (m gpioManager) turnPinOn(pin string) {
	if v, ok := m.PinStates[pin]; !ok {
		fmt.Println("Pin ", pin, " not set in the initial configuration")
	} else if !v.state {
		if m.gpioAvailable {
			rpio.Pin(v.pin).High()
		}
		m.PinStates[pin].state = true
		fmt.Println("Pin ", pin, " turned on")
	}
}

func (m gpioManager) turnPinOff(pin string) {
	if v, ok := m.PinStates[pin]; !ok {
		fmt.Println("Pin ", pin, " not set in the initial configuration")
	} else if v.state {
		if m.gpioAvailable {
			rpio.Pin(v.pin).Low()
		}
		m.PinStates[pin].state = false
		fmt.Println("Pin ", pin, " turned off")
	}
}

func (m gpioManager) clearAllPins() {
	for _, v := range m.PinStates {
		if m.gpioAvailable {
			rpio.Pin(v.pin).Low()
		}
		v.state = false
	}
}

func (m gpioManager) getPinState(pin string) bool {
	if v, ok := m.PinStates[pin]; ok {
		return v.state
	}
	return false
}
