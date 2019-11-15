package console

var (
	convars map[string]*ConVar
)

func init() {
	convars = map[string]*ConVar{}
}

type ConVar struct {
	name         string
	defaultValue string
	flags        int
	helpText     string
	minBool      bool
	maxBool      bool
	minFloat     float64
	maxFloat     float64
}

func (cv *ConVar) IsFlagSet(flag int) bool {
	return false
}

func (cv *ConVar) SetFlag(flag int) {
	cv.flags &= flag
}

func (cv *ConVar) Name() string {
	return cv.name
}

func NewConVar(name, defaultValue string, flags int, helpText string) *ConVar {
	convars[name] = &ConVar{
		name:         name,
		defaultValue: defaultValue,
		flags:        flags,
		helpText:     helpText,
		minBool:      false,
		maxBool:      false,
		minFloat:     0,
		maxFloat:     0,
	}

	return convars[name]
}
