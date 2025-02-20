package ik

type EventHandler = func(data []byte) []byte

type EventHandlerMap map[string]EventHandler

func (e *EventHandlerMap) Register(event string, handler EventHandler) {
	(*e)[event] = handler
}
