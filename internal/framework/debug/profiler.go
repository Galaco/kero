package debug

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/galaco/kero/internal/framework/console"
)

// StartProfiler starts pprof on localhost:6060
// To save a dump of the current state, run: `curl http://localhost:6060/debug/pprof/heap --output heap.tar.gz`
// Then analze with: go tool pprof heap.tar.gz
func StartProfiler() {
	go func() {
		console.PrintInterface(console.LevelInfo, http.ListenAndServe("localhost:6060", nil))
	}()
}
