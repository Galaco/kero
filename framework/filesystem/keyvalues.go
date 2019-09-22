package filesystem

import (
	keyvalues "github.com/galaco/KeyValues"
	"os"
)

func ReadKeyValues(path string) (*keyvalues.KeyValue, error) {
	f,err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := keyvalues.NewReader(f)
	kv,err := r.Read()
	if err != nil {
		return nil, err
	}
	return &kv,nil
}
