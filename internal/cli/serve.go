package cli

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/LaPingvino/recuerdo/internal/resources"
)

// webDirs are where the built web version may be: installed with the
// data (/usr/share/recuerdo/web), or built in a source tree (dist/web).
func webDirs() []string {
	return []string{filepath.Join(resources.Dir(), "web"), filepath.Join(resources.Dir(), "dist", "web"), filepath.Join("dist", "web")}
}

// webHandler serves the web version in dir (it must hold recuerdo.wasm).
func webHandler(dir string) (http.Handler, error) {
	if _, err := os.Stat(filepath.Join(dir, "recuerdo.wasm")); err != nil {
		return nil, fmt.Errorf("no web version in %s: build it with scripts/build-web.sh", dir)
	}
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	}), nil
}

func serve(c *ctx, args []string) error {
	fs := flags(c, "serve")
	addr := fs.String("addr", "127.0.0.1:8765", "address to listen on (127.0.0.1: this computer only)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dirs := webDirs()
	if fs.NArg() > 0 {
		dirs = []string{fs.Arg(0)}
	}
	var h http.Handler
	var err error
	for _, d := range dirs {
		if h, err = webHandler(d); err == nil {
			fmt.Fprintf(c.out, "Serving %s\n", d)
			break
		}
	}
	if h == nil {
		return err
	}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "Recuerdo's web version: http://%s/  (Ctrl+C stops it)\n", ln.Addr())
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
