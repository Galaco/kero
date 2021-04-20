package debug

import (
	"github.com/galaco/kero/framework/console"
	"net/http"
	_ "net/http/pprof"
)

// StartProfiler starts pprof on localhost:6060
// To save a dump of the current state, run: `curl http://localhost:6060/debug/pprof/heap --output heap.tar.gz`
// Then analze with: go tool pprof heap.tar.gz
func StartProfiler() {
	go func() {
		console.PrintInterface(console.LevelInfo, http.ListenAndServe("localhost:6060", nil))
	}()
}