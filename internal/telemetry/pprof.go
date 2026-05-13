package telemetry

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"go.uber.org/dig"
)

type PProfListenerDeps struct {
	dig.In

	// Config
	Enabled bool   `name:"config.pprofListener.enabled"`
	Addr    string `name:"config.pprofListener.addr"`
}

// StartPProfListener starts pprof listener in a separate goroutine.
// If enabled, you can collect pprof with the go tool pprof. Some examples:
//   - go tool pprof http://localhost:6060/debug/pprof/goroutine
//   - go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
//   - go tool pprof http://localhost:6060/debug/pprof/heap
//   - curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=10
//     go tool trace trace.out
//
// Also you can use the pprof UI at http://localhost:6060/debug/pprof/
func StartPProfListener(deps PProfListenerDeps) error { // coverage-ignore
	if !deps.Enabled {
		return nil
	}
	go func() {
		pprofMux := http.NewServeMux()

		pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// We don't need to worry about graceful shutdown here
		//nolint:gosec // Development-only pprof listener.
		if err := http.ListenAndServe(deps.Addr, pprofMux); err != nil {
			panic(fmt.Errorf("failed to start pprof listener on %s: %w", deps.Addr, err))
		}
	}()
	return nil
}
