package watcher

type EventType int

const (
	CreateEvent EventType = iota
	RemoveEvent
	WriteEvent
)

type Event struct {
	Type EventType
	Path string
}

func (e Event) String() string {
	var name string
	switch e.Type {
	case CreateEvent:
		name = "Create"
	case RemoveEvent:
		name = "Remove"
	case WriteEvent:
		name = "Write"
	}
	return "{Event:" + name + " Path:" + e.Path + "}"
}

type Watcher interface {
	Run(chan Event)
	Close()
}
