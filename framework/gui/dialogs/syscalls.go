package dialogs

import (
	"github.com/sqweek/dialog"
)

// OpenFile
func OpenFile(title, startDir, filterDescription, extension string) (string, error) {
	return dialog.File().Title(title).Filter(filterDescription, extension).SetStartDir(startDir).Load()
}

// ErrorMessage
func ErrorMessage(err error) {
	dialog.Message("%s", err.Error()).Error()
}
