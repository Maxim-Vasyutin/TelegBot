package events

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}


type Processor interface {
	Process(e Event) error
}

type Type int 

const (
	Unknow Type = iota //йота первому значению присвоит 0, а дальше +1(или что-то такое)
	Message 
)

type Event struct {
	Type Type
	Text string
	Meta interface{}
}