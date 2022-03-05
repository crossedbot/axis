package controller

import (
	"net/http"

	"github.com/crossedbot/common/golang/server"

	"github.com/crossedbot/axis/pkg/auth"
)

// Route represents the route of the Pins HTTP API
type Route struct {
	Handler          server.Handler
	Method           string
	Path             string
	ResponseSettings []server.ResponseSetting
}

// Routes is the list of routes of the Pins HTTP API
var Routes = []Route{
	// GetPin
	Route{
		auth.Authenticate(GetPin),
		http.MethodGet,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// FindPins
	Route{
		auth.Authenticate(FindPins),
		http.MethodGet,
		"/pins",
		[]server.ResponseSetting{},
	},
	// CreatePin
	Route{
		auth.Authenticate(CreatePin),
		http.MethodPost,
		"/pins",
		[]server.ResponseSetting{},
	},
	// UpdatePin
	Route{
		auth.Authenticate(UpdatePin),
		http.MethodPut,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// UpdatePin (POST alternative)
	Route{
		auth.Authenticate(UpdatePin),
		// XXX POST for compatibility with existing pinning api specs
		http.MethodPost,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// PatchPin
	Route{
		auth.Authenticate(PatchPin),
		http.MethodPatch,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// RemovePin
	Route{
		auth.Authenticate(RemovePin),
		http.MethodDelete,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
}
