package port

// Logger records structured gateway events.
type Logger interface {
	Info(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}
