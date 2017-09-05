package filter

import (
	"testing"

	"github.com/coldog/logship/input"
	"github.com/stretchr/testify/assert"
)

func TestRegex(t *testing.T) {
	r := &Regex{
		Regex: `^(?P<param1>[^ ]+) (?P<param2>[^ ]+)$`,
		Field: "test",
	}
	err := r.Open()
	assert.Nil(t, err)

	m := &input.Message{
		Data: map[string]string{"test": "test test"},
	}
	r.Filter(m)
	assert.Equal(t, "test", m.Data["param1"])
	assert.Equal(t, "test", m.Data["param2"])
}
