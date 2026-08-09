package di

// Module registers providers on a Builder for one bounded context.
type Module interface {
	Register(b *Builder)
}

// FuncModule adapts a function as a Module, useful for tests and overrides.
type FuncModule struct {
	Name string
	Fn   func(b *Builder)
}

// Register implements Module.
func (m FuncModule) Register(b *Builder) {
	if m.Fn != nil {
		m.Fn(b)
	}
}
