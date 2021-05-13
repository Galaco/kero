package messages

import (
	"github.com/galaco/kero/framework/event"
)

const (
	// TypeLoadingLevelParsed
	TypeLoadingLevelParsed = event.Type("ui:LoadingLevelParsed")
	// TypeLoadingLevelProgress
	TypeLoadingLevelProgress = event.Type("ui:LoadingLevelProgress")
)

const (
	// LoadingProgressStateError
	LoadingProgressStateError = -1
	// LoadingProgressStateStarted
	LoadingProgressStateStarted = 0
	// LoadingProgressStateBSPParsed
	LoadingProgressStateBSPParsed = 1
	// LoadingProgressStateGeometryLoaded
	LoadingProgressStateGeometryLoaded = 2
	// LoadingProgressStateStaticPropsLoaded
	LoadingProgressStateStaticPropsLoaded = 3
	// LoadingProgressStateEntitiesLoaded
	LoadingProgressStateEntitiesLoaded = 4
	// LoadingProgressStateDynamicPropsLoaded
	LoadingProgressStateDynamicPropsLoaded = 5
	// LoadingProgressStateFinished
	LoadingProgressStateFinished = 6
)

type LoadingLevelParsed struct {
	level    interface{}
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
		level:    level,
	}
}
