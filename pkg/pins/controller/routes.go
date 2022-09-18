package controller

import (
	"net/http"

	"github.com/crossedbot/common/golang/server"
	middleware "github.com/crossedbot/simplemiddleware"
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
		middleware.Authorize(GetPin),
		http.MethodGet,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// FindPins
	Route{
		middleware.Authorize(FindPins),
		http.MethodGet,
		"/pins",
		[]server.ResponseSetting{},
	},
	// CreatePin
	Route{
		middleware.Authorize(CreatePin),
		http.MethodPost,
		"/pins",
		[]server.ResponseSetting{},
	},
	// UpdatePin
	Route{
		middleware.Authorize(UpdatePin),
		http.MethodPut,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// UpdatePin (POST alternative)
	Route{
		middleware.Authorize(UpdatePin),
		// XXX POST for compatibility with existing pinning api specs
		http.MethodPost,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// PatchPin
	Route{
		middleware.Authorize(PatchPin),
		http.MethodPatch,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
	// RemovePin
	Route{
		middleware.Authorize(RemovePin),
		http.MethodDelete,
		"/pins/:id",
		[]server.ResponseSetting{},
	},
}
