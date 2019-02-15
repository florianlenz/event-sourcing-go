package main

type EventHandler struct {
}

func (eh *EventHandler) Handle() {

}

func main() {

}

type Projector interface {
	OnEvents()
	Handle()
}
