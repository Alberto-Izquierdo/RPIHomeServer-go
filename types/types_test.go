package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProgrammedActionFromString(t *testing.T) {
	{
		message := ""
		programmedAction, err := ProgrammedActionFromString(message, 0)
		assert.NotNil(t, err)
		assert.Nil(t, programmedAction)
	}
	{
		message := "action;false;true;23:40:08"
		programmedAction, err := ProgrammedActionFromString(message, 0)
		assert.Nil(t, err)
		assert.NotNil(t, programmedAction)
		assert.Equal(t, programmedAction.Action.Pin, "action")
		assert.Equal(t, programmedAction.Action.State, false)
		assert.Equal(t, programmedAction.Repeat, true)
		assert.Equal(t, time.Time(programmedAction.Time).Hour(), 23)
		assert.Equal(t, time.Time(programmedAction.Time).Minute(), 40)
		assert.Equal(t, time.Time(programmedAction.Time).Second(), 8)
	}
}
