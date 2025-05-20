package handlers

// AsyncHandler is an interface for asynchronous request handlers
type AsyncHandler[Args any, Result any] interface {
	// Handle processes the given arguments and returns a result
	Handle(args Args) (Result, error)
}
