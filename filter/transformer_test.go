package filter

import (
	"testing"

	"github.com/coldog/logship/input"
	"github.com/stretchr/testify/assert"
)

func TestTransformer_Tag(t *testing.T) {
	tr := &Transformer{
		Tag: "{{.Tag}}.{{.Data.name}}",
	}
	err := tr.Open()
	assert.Nil(t, err)

	m := &input.Message{
		Tag: "test",
		Data: map[string]string{
			"name": "test",
		},
	}
	ok := tr.Filter(m)
	assert.True(t, ok)

	assert.Equal(t, "test.test", m.Tag)
}

func TestTransformer_Record(t *testing.T) {
	tr := &Transformer{
		Record: map[string]string{
			"Key": "{{.Tag}}.{{.Data.name}}",
		},
	}
	err := tr.Open()
	assert.Nil(t, err)

	m := &input.Message{
		Tag: "test",
		Data: map[string]string{
			"Key":  "test",
			"name": "test",
		},
	}
	ok := tr.Filter(m)
	assert.True(t, ok)

	assert.Equal(t, "test.test", m.Data["Key"])
}
