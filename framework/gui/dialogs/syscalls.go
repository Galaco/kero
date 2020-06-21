package dialogs

import (
	"github.com/sqweek/dialog"
)

// OpenFile
func OpenFile(filterDescription, extension string) (string, error) {
	return dialog.File().Filter(filterDescription, extension).Load()
}

// ErrorMessage
func ErrorMessage(err error) {
	dialog.Message("%s", err.Error()).Error()
}
