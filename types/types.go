package types

import (
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
		otherTime.Time == this.Time
	return equal
}
