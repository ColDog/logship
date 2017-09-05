package pipeline

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/coldog/logship/input"
	"github.com/coldog/logship/output"
	"github.com/stretchr/testify/assert"
)

func init() { input.Register("test", func() input.Input { return &TestInput{} }) }

type TestInput struct {
	input.BaseInput
}

func (m *TestInput) Run(out chan *input.Message) {
	defer m.Finish()

	for {
		select {
		case <-m.Done():
			return
		default:
		}
		msg := &input.Message{
			Tag:  "test",
			Data: map[string]string{"message": "testing"},
			Time: time.Now(),
		}
		out <- msg
	}
}

func TestPipeline(t *testing.T) {
	pipe := &Pipeline{
		Inputs: []*input.Spec{
			{Type: "test", Spec: json.RawMessage(`{}`)},
		},
		Outputs: []*output.Spec{
			{Type: "stdout", Match: "*", Spec: json.RawMessage(`{}`)},
		},
	}

	err := pipe.Open()
	assert.Nil(t, err)

	go func() {
		time.Sleep(10 * time.Millisecond)
		pipe.Close()
	}()

	pipe.Run()
}
