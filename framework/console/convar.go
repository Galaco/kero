package console

import "reflect"

type convarList struct {
	convars map[string]Convar
}

var convarSingleton convarList

type Convar struct {
	Key         string
	Description string
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
	setConvarAny(key, value)
}

func GetConvarInt(key string) int {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(int)
	}
	return 0
}

func SetConvarInt(key string, value int) {
	setConvarAny(key, value)
}

func GetConvarString(key string) string {
	if convar, ok := convarSingleton.convars[key]; ok {
		return convar.Value.(string)
	}
	return ""
}

func SetConvarString(key string, value string) {
	setConvarAny(key, value)
}

func setConvarAny(key string, value interface{}) {
	if cv, ok := convarSingleton.convars[key]; ok {
		// Ensure the original convar type does not change under the hood
		if reflect.TypeOf(cv.Value) != reflect.TypeOf(value) {
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
		Value:       value,
	}
}

func AddConvarInt(key string, description string, value int) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Value:       value,
	}
}

func AddConvarString(key string, description string, value string) {
	convarSingleton.convars[key] = Convar{
		Key:         key,
		Description: description,
		Value:       value,
	}
}

func init() {
	convarSingleton.convars = map[string]Convar{}
}
