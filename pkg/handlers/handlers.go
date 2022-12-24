package handlers

import (
	"github.com/viveksahu26/kubewatchpods/pkg/config"
	"github.com/viveksahu26/kubewatchpods/pkg/pkg/event"
)

// handlers handel events
type Handler interface {
	Init(c *config.Config) error
	Handle(e event.Event)
}

// Map maps each event handler function to a name for easily lookup
var Map = map[string]interface{}{
	"default": &Default{},
}

// Default handler implements Handler interface,
// print each event with JSON format
type Default struct{}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c *config.Config) error {
	return nil
}

// Handle handles an event.
func (d *Default) Handle(e event.Event) {
}
