package types

import (
	"errors"
	"strings"
	"time"
)

type PairNamePin struct {
	Name string
	Pin  int
}

type Action struct {
	Pin    string
	State  bool
	ChatId int64
}

type TelegramMessage struct {
	Message string
	ChatId  int64
}

type MyTime time.Time

type ProgrammedAction struct {
	Action Action
	Repeat bool
	Time   MyTime
}

type ProgrammedActionOperation struct {
	ProgrammedAction ProgrammedAction
	Operation        int32
}

const (
	CREATE = iota
	REMOVE
	GET_ACTIONS
)

func (a *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	*a = MyTime(t)
	return nil
}

func (a MyTime) Format(s string) string {
	t := time.Time(a)
	return t.Format(s)
}

func (this ProgrammedAction) LessThan(other interface{}) bool {
	return time.Time(this.Time).Before(time.Time(other.(ProgrammedAction).Time))
}

func (this ProgrammedAction) Equals(other interface{}) bool {
	otherTime := other.(ProgrammedAction)
	equal := otherTime.Action.Pin == this.Action.Pin &&
		otherTime.Action.State == this.Action.State &&
		time.Time(otherTime.Time).Hour() == time.Time(this.Time).Hour() &&
		time.Time(otherTime.Time).Minute() == time.Time(this.Time).Minute() &&
		time.Time(otherTime.Time).Second() == time.Time(this.Time).Second()
	return equal
}

func ProgrammedActionFromString(str string, chatId int64) (*ProgrammedAction, error) {
	fields := strings.Split(str, ";")
	if len(fields) != 4 {
		return nil, errors.New("Message not correct")
	}
	deserializedTime := MyTime{}
	err := deserializedTime.UnmarshalJSON([]byte(fields[3]))
	if err != nil {
		return nil, err
	}
	currTime := time.Time(deserializedTime)
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), currTime.Hour(), currTime.Minute(), currTime.Second(), 0, now.Location())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
	}
	state := false
	if strings.EqualFold(fields[1], "true") {
		state = true
	}
	repeat := false
	if strings.EqualFold(fields[2], "true") {
		repeat = true
	}
	result := ProgrammedAction{
		Action: Action{
			Pin:    fields[0],
			State:  state,
			ChatId: chatId,
		},
		Repeat: repeat,
		Time:   MyTime(date),
	}
	return &result, nil
}

func ProgrammedActionToString(p ProgrammedAction) string {
	result := p.Action.Pin + ";"
	if p.Action.State {
		result += "true;"
	} else {
		result += "false;"
	}
	if p.Repeat {
		result += "true;"
	} else {
		result += "false;"
	}
	result += p.Time.Format("15:04:05")
	return result
}
