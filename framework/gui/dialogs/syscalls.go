package dialogs

import (
	"github.com/sqweek/dialog"
)

// OpenFile
func OpenFile(filterDescription, extension string) (string, error) {
	return dialog.File().SetStartDir("/Users/galaco/Library/Application Support/Steam/steamapps/common/Counter-Strike Source").Filter(filterDescription, extension).Load()
}

// ErrorMessage
func ErrorMessage(err error) {
	dialog.Message("%s", err.Error()).Error()
}
