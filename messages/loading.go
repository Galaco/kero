package messages

import (
	"github.com/galaco/kero/event"
)

const (
	// TypeLoadingLevelParsed
	TypeLoadingLevelParsed   = event.Type("LoadingLevelParsed")
	// TypeLoadingLevelProgress
	TypeLoadingLevelProgress = event.Type("LoadingLevelProgress")
)

const (
	// LoadingProgressStateError
	LoadingProgressStateError              = -1
	// LoadingProgressStateStarted
	LoadingProgressStateStarted            = 0
	// LoadingProgressStateBSPParsed
	LoadingProgressStateBSPParsed          = 1
	// LoadingProgressStateGeometryLoaded
	LoadingProgressStateGeometryLoaded     = 2
	// LoadingProgressStateStaticPropsLoaded
	LoadingProgressStateStaticPropsLoaded  = 3
	// LoadingProgressStateEntitiesLoaded
	LoadingProgressStateEntitiesLoaded     = 4
	// LoadingProgressStateDynamicPropsLoaded
	LoadingProgressStateDynamicPropsLoaded = 5
	// LoadingProgressStateFinished
	LoadingProgressStateFinished           = 6
)

type LoadingLevelParsed struct {
	level interface{}
}

// Type
func (msg *LoadingLevelParsed) Type() event.Type {
	return TypeLoadingLevelParsed
}

// Level
func (msg *LoadingLevelParsed) Level() interface{} {
	return msg.level
}

// NewLoadingLevelParsed
func NewLoadingLevelParsed(level interface{}) *LoadingLevelParsed {
	return &LoadingLevelParsed{
		level: level,
	}
}

// LoadingLevelProgress
type LoadingLevelProgress struct {
	state int
}

// Type
func (msg *LoadingLevelProgress) Type() event.Type {
	return TypeLoadingLevelProgress
}

// State
func (msg *LoadingLevelProgress) State() int {
	return msg.state
}

// LoadingLevelProgress
func NewLoadingLevelProgress(state int) *LoadingLevelProgress {
	return &LoadingLevelProgress{
		state: state,
	}
}
