package output

import (
	"encoding/json"
	"fmt"

	"github.com/coldog/logship/input"
)

var plugins = map[string]func() Output{}

// Register an output plugin. This should be passed a plugin that will return a
// new initialized struct that can be marshalled using json.
func Register(name string, plugin func() Output) {
	plugins[name] = plugin
}

// Spec represents a serialized output plugin. It can be deserialized using the
// Output method.
type Spec struct {
	Type  string          `json:"type"`  // Type is the plugin name.
	Match string          `json:"match"` // Matches the tag of the message.
	Spec  json.RawMessage `json:"spec"`  // Spec is the JSON representation.
}

// Output deserializes the output plugin.
func (s *Spec) Output() (Output, error) {
	cons, ok := plugins[s.Type]
	if !ok {
		return nil, fmt.Errorf("output does not exist: %s", s.Type)
	}
	output := cons()
	err := json.Unmarshal(s.Spec, output)
	return output, err
}

// Output plugins handle pushing log messages from the provided channel to the
// configured destination.
type Output interface {
	// Open should configure the plugin. An error should be returned if the
	// plugin is misconfigured.
	Open() error

	// Run should start the Output plugin, if the passed in channel is closed,
	// the plugin should exit.
	Run(<-chan *input.Message)
}
