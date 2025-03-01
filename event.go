package ik

// EventHandler is a function that is passed a ServerClient and event data received from a Client.
type EventHandler = func(c *ServerClient, data []byte) []byte

type eventHandlerMap map[string]EventHandler

func (e *eventHandlerMap) Register(event string, handler EventHandler) {
	(*e)[event] = handler
}
