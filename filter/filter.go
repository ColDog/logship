package filter

import (
	"encoding/json"
	"fmt"

	"github.com/coldog/logship/input"
)

var plugins = map[string]func() Filter{}

// Register a filter plugin. This should be passed a plugin that will return a
// new initialized struct that can be marshalled using json.
func Register(name string, plugin func() Filter) {
	plugins[name] = plugin
}

// Spec represents a serialized filter plugin. It can be deserialized using the
// Filter method.
type Spec struct {
	Type  string          // The plugin name.
	Match string          // Matches the message tag.
	Spec  json.RawMessage // JSON representation of the plugin.
}

// Filter fetches the deserialized filter plugin. The Type parameter is used to
// match the name of the plugin.
func (s *Spec) Filter() (Filter, error) {
	cons, ok := plugins[s.Type]
	if !ok {
		return nil, fmt.Errorf("filter does not exist: %s", s.Type)
	}
	filter := cons()
	err := json.Unmarshal(s.Spec, filter)
	return filter, err
}

// Filter is called once for every message if the Match parameter matches.
type Filter interface {
	Open() error
	Filter(*input.Message) bool
}
