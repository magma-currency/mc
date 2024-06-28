package globals

import (
	"mc/helpers/events"
	"mc/helpers/generics"
)

// arguments
var (
	MainEvents  = events.NewEvents[any]()
	MainStarted = generics.Value[bool]{}
)

func init() {
	MainStarted.Store(false)
}
