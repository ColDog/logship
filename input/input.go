package input

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessagePool is the global message pool that should be used for initializing
// messages inside the input plugins and putting back messages inside output
// plugins.
var MessagePool = sync.Pool{
	New: func() interface{} { return &Message{Data: make(map[string]string, 5)} },
}

var plugins = map[string]func() Input{}

// Register an input plugin. This should be passed a plugin that will return a
// new initialized struct that can be marshalled using json.
func Register(name string, plugin func() Input) {
	plugins[name] = plugin
}

// Spec represents a serialized input plugin. It can be deserialized using the
// Input fn.
type Spec struct {
	Type string          `json:"type"`
	Spec json.RawMessage `json:"spec"`
}

// Input fetches the deserialized input plugin. The Type parameter is used as to
// match the name of the plugin.
func (s *Spec) Input() (Input, error) {
	cons, ok := plugins[s.Type]
	if !ok {
		return nil, fmt.Errorf("input does not exist: %s", s.Type)
	}
	input := cons()
	err := json.Unmarshal(s.Spec, input)
	return input, err
}

// Message represents a log message.
type Message struct {
	Time time.Time
	Tag  string
	Data map[string]string
}

func (m *Message) String() string {
	return fmt.Sprintf("%v %s: %+s", m.Time.Format(time.RFC3339), m.Tag, m.Data)
}

// Clear clears the message before returning to the pool.
func (m *Message) Clear() *Message {
	for k := range m.Data {
		delete(m.Data, k)
	}
	m.Tag = ""
	m.Time = time.Time{}
	return m
}

// Input is an input plugin. It writes Message objects to the channel provided.
//
// Input plugins should be structs which hold configuration data and can be
// marshalled using json. Once the object is unmarshalled, it will be run.
type Input interface {
	// Open is called to initialize the plugin. It should return an error if it
	// is misconfigured.
	Open() error

	// Run will be called after Open is called in a separate goroutine. It should
	// write messages into the channel. If errors are encountered, it should log
	// them and continue to attempt to write messages.
	Run(out chan *Message)

	// Close should shut down the run function and ensure that it doesn't write
	// to the out channel after this is called.
	Close()
}

// BaseInput is a generic structure that handles the exit behaviour.
type BaseInput struct {
	done chan struct{}
	exit chan struct{}
}

// Open will initialize the input.
func (m *BaseInput) Open() error {
	m.done = make(chan struct{})
	m.exit = make(chan struct{})
	return nil
}

// Close will close the done channel and wait until Finish is called.
func (m *BaseInput) Close() {
	close(m.done)
	<-m.exit
}

// Done returns a channel that will be closed when the program has exited.
func (m *BaseInput) Done() <-chan struct{} { return m.done }

// Finish will signal that the main goroutine has exited. This should be called
// from the main Run function using a defer.
func (m *BaseInput) Finish() { close(m.exit) }
