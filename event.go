package ik

type EventHandler = func(c *ServerClient, data []byte) []byte

type eventHandlerMap map[string]EventHandler

func (e *eventHandlerMap) Register(event string, handler EventHandler) {
	(*e)[event] = handler
}
