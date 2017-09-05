package filter

import (
	"regexp"

	"github.com/coldog/logship/input"
)

// Regex is a regex filter that extracts all the named captures provided and
// extends the record with those behaviours.
type Regex struct {
	Regex string
	Field string // Field is the field to apply the regex on.

	regex *regexp.Regexp
}

// Open will compile the regex.
func (r *Regex) Open() error {
	if r.Field == "" {
		r.Field = "message"
	}
	reg, err := regexp.Compile(r.Regex)
	if err != nil {
		return err
	}
	r.regex = reg
	return nil
}

// Filter applies the named captures in aggregates them into the map.
func (r *Regex) Filter(m *input.Message) bool {
	if val, ok := m.Data[r.Field]; ok {
		matches := r.regex.FindStringSubmatch(val)
		for i, name := range r.regex.SubexpNames() {
			if i != 0 && name != "" && matches[i] != "" {
				m.Data[name] = matches[i]
			}
		}
	}
	return true
}
