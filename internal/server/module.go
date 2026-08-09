package server

import "github.com/mauriciomendonca/universal-api-gateway/internal/di"

// Module registers HTTP-edge providers for the dependency graph.
// M0: feature modules supply ports; the HTTP adapter is constructed from
// server.Dependencies after Build(). Future edge adapters register here.
type Module struct{}

// Register implements di.Module.
func (Module) Register(_ *di.Builder) {}
