package console

import (
	"strings"
)

type ConvarType byte

const (
	ConvarTypeBool ConvarType = iota
	ConvarTypeInt
	ConvarTypeFloat
	ConvarTypeString
)

type convarList struct {
	convars map[string]Convar
}

var convarSingleton convarList

type Convar struct {
	Key         string
	Description string
	Type        ConvarType
	Value       interface{}
}

func GetConvar(key string) *Convar {
	if convar, ok := convarSingleton.convars[key]; ok {
		return &convar
	}
	return nil
}

func GetConvarBoolean(key string) bool {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(bool)
	}
	return false
}

func SetConvarBoolean(key string, value bool) {
	setConvarAny(key, value, ConvarTypeBool)
}

// SetConvarBooleanInt sets a boolean convar using an integer value.
// 0 is treated as false, 1 is treated as true, other values are ignored.
func SetConvarBooleanInt(key string, value int) {
	if value == 0 {
		setConvarAny(key, false, ConvarTypeBool)
	} else if value == 1 {
		setConvarAny(key, true, ConvarTypeBool)
	}
	// Other values are ignored as requested
}

func GetConvarInt(key string) int {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(int)
	}
	return 0
}

func SetConvarInt(key string, value int) {
	setConvarAny(key, value, ConvarTypeInt)
}

func GetConvarString(key string) string {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(string)
	}
	return ""
}

func SetConvarString(key string, value string) {
	setConvarAny(key, value, ConvarTypeString)
}

func GetConvarFloat(key string) float32 {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(float32)
	}
	return 0
}

func SetConvarFloat(key string, value float32) {
	setConvarAny(key, value, ConvarTypeFloat)
}

func setConvarAny(key string, value interface{}, expectedType ConvarType) {
	if cv, ok := convarSingleton.convars[key]; ok {
		// Ensure the original convar type does not change under the hood
		if cv.Type != expectedType {
			return
		}
		cv.Value = value
		convarSingleton.convars[key] = cv
	}
}

func AddConvarBool(key string, description string, value bool) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Type:        ConvarTypeBool,
		Value:       value,
	}
}

func AddConvarInt(key string, description string, value int) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Type:        ConvarTypeInt,
		Value:       value,
	}
}

func AddConvarString(key string, description string, value string) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Type:        ConvarTypeString,
		Value:       value,
	}
}

func AddConvarFloat(key string, description string, value float32) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Type:        ConvarTypeFloat,
		Value:       value,
	}
}

// GetConvarList returns all convar keys that match the given prefix
func GetConvarList(prefix string) []string {
	convars := make([]string, 0, len(convarSingleton.convars))

	for key := range convarSingleton.convars {
		if len(prefix) == 0 {
			convars = append(convars, key)
			continue
		}

		if strings.HasPrefix(key, prefix) {
			convars = append(convars, key)
		}
	}

	return convars
}

func init() {
	convarSingleton.convars = map[string]Convar{}
}
