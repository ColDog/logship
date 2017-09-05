package pipeline

import (
	"github.com/coldog/logship/filter"
	"github.com/coldog/logship/input"
	"github.com/coldog/logship/output"
)

type step struct {
	match string
}

func (s *step) matches(tag string) bool {
	for i := 0; i < len(tag); i++ {
		if s.match[i] == '*' {
			return true
		}
		if tag[i] != s.match[i] {
			return false
		}
	}
	return true
}

type filterStep struct {
	step
	filter filter.Filter
}

type inputStep struct {
	input input.Input
}

type outputStep struct {
	step
	ch     chan *input.Message
	output output.Output
}

// Pipeline describes a full logging pipeline.
type Pipeline struct {
	Inputs  []*input.Spec  `json:"inputs"`
	Filters []*filter.Spec `json:"filters"`
	Outputs []*output.Spec `json:"outputs"`

	inputCh     chan *input.Message
	exit        chan struct{}
	inputSteps  []*inputStep
	filterSteps []*filterStep
	outputSteps []*outputStep
}

// Open initializes all inputs, filters and outputs.
func (pipe *Pipeline) Open() error {
	pipe.exit = make(chan struct{})
	pipe.inputCh = make(chan *input.Message)

	for _, in := range pipe.Inputs {
		input, err := in.Input()
		if err != nil {
			return err
		}
		err = input.Open()
		if err != nil {
			return err
		}
		pipe.inputSteps = append(pipe.inputSteps, &inputStep{
			input: input,
		})
	}
	for _, in := range pipe.Filters {
		filter, err := in.Filter()
		if err != nil {
			return err
		}
		err = filter.Open()
		if err != nil {
			return err
		}
		pipe.filterSteps = append(pipe.filterSteps, &filterStep{
			step:   step{match: in.Match},
			filter: filter,
		})
	}
	for _, in := range pipe.Outputs {
		output, err := in.Output()
		if err != nil {
			return err
		}
		err = output.Open()
		if err != nil {
			return err
		}
		pipe.outputSteps = append(pipe.outputSteps, &outputStep{
			step:   step{match: in.Match},
			output: output,
		})
	}
	for _, out := range pipe.outputSteps {
		out.ch = make(chan *input.Message)
	}
	return nil
}

// Run starts the pipeline.
//
// Each input step and output step are started in their own goroutines. A single
// channel is used for all input steps. When messages are pushed into this
// channel the appropriate filters are applied and the messages are pushed into
// all matching output channels.
func (pipe *Pipeline) Run() {
	for _, in := range pipe.inputSteps {
		go in.input.Run(pipe.inputCh)
	}
	for _, out := range pipe.outputSteps {
		go out.output.Run(out.ch)
	}

	for msg := range pipe.inputCh {
		ok := true
		for _, filt := range pipe.filterSteps {
			if filt.matches(msg.Tag) {
				ok = filt.filter.Filter(msg)
			}
		}
		if ok {
			for _, out := range pipe.outputSteps {
				if out.matches(msg.Tag) {
					out.ch <- msg
				}
			}
		}
	}
	close(pipe.exit)
}

// Close stops each input step and then closes the input channel. It then closes
// each output channel.
func (pipe *Pipeline) Close() {
	for _, in := range pipe.inputSteps {
		in.input.Close()
	}
	close(pipe.inputCh)
	<-pipe.exit
	for _, in := range pipe.outputSteps {
		close(in.ch)
	}
}
