package core

import (
	"fmt"
	"github.com/stianeikeland/go-rpio"
)

type gpioManager struct {
	pinStates     map[int]bool
	gpioAvailable bool
}

func (m gpioManager) turnPinOn(pin int) {
	v, ok := m.pinStates[pin]
	if !ok || !v {
		if m.gpioAvailable {
			rpio.Pin(pin).High()
		}
		m.pinStates[pin] = true
	}
	fmt.Println("Pin ", pin, " turned on")
}

func (m gpioManager) turnPinOff(pin int) {
	v, ok := m.pinStates[pin]
	if !ok || v {
		if m.gpioAvailable {
			rpio.Pin(pin).High()
		}
		m.pinStates[pin] = false
	}
	fmt.Println("Pin ", pin, " turned off")
}

func (m gpioManager) clearAllPins() {
	for k, _ := range m.pinStates {
		if m.gpioAvailable {
			rpio.Pin(k).Low()
		}
		m.pinStates[k] = false
	}
}

func (m gpioManager) getPinState(pin int) bool {
	v, ok := m.pinStates[pin]
	if ok {
		return v
	}
	return false
}

var manager gpioManager

func Setup() {
	manager.pinStates = make(map[int]bool)
	manager.gpioAvailable = true
	err := rpio.Open()
	if err != nil {
		manager.gpioAvailable = false
		fmt.Println("Unable to open gpio, error:", err.Error())
		fmt.Println("The program will continue for testing purpouses")
	}
}

func TurnPinOn(pin int) {
	manager.turnPinOn(pin)
}

func TurnPinOff(pin int) {
	manager.turnPinOff(pin)
}

func ClearAllPins() {
	manager.clearAllPins()
	if manager.gpioAvailable {
		rpio.Close()
	}
}

func GetPinState(pin int) bool {
	return manager.getPinState(pin)
}
