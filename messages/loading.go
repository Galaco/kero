package messages

import (
	"github.com/galaco/kero/event"
	"github.com/galaco/kero/valve"
)

const (
	TypeLoadingLevelParsed   = event.Type("LoadingLevelParsed")
	TypeLoadingLevelProgress = event.Type("LoadingLevelProgress")

	LoadingProgressStateError              = -1
	LoadingProgressStateStarted            = 0
	LoadingProgressStateBSPParsed          = 1
	LoadingProgressStateGeometryLoaded     = 2
	LoadingProgressStateStaticPropsLoaded  = 3
	LoadingProgressStateEntitiesLoaded     = 4
	LoadingProgressStateDynamicPropsLoaded = 5
	LoadingProgressStateFinished           = 6
)

type LoadingLevelParsed struct {
	level *valve.Bsp
}

func (msg *LoadingLevelParsed) Type() event.Type {
	return TypeLoadingLevelParsed
}

func (msg *LoadingLevelParsed) Level() *valve.Bsp {
	return msg.level
}

func NewLoadingLevelParsed(level *valve.Bsp) *LoadingLevelParsed {
	return &LoadingLevelParsed{
		level: level,
	}
}

type LoadingLevelProgress struct {
	state int
}

func (msg *LoadingLevelProgress) Type() event.Type {
	return TypeLoadingLevelProgress
}

func (msg *LoadingLevelProgress) State() int {
	return msg.state
}

func NewLoadingLevelProgress(state int) *LoadingLevelProgress {
	return &LoadingLevelProgress{
		state: state,
	}
}
