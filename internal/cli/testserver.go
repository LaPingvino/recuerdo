package cli

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/LaPingvino/recuerdo/internal/testserver"
	"github.com/LaPingvino/recuerdo/internal/version"
)

// defaultTestDB is where the test server keeps its data by default.
func defaultTestDB() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "recuerdo", "testserver.db")
}

func testServer(c *ctx, args []string) error {
	fs := flags(c, "testserver")
	addr := fs.String("addr", "127.0.0.1:8770", "address to listen on (127.0.0.1: this computer only; :8770: the network too)")
	db := fs.String("db", defaultTestDB(), "the database file")
	admin := fs.String("admin", "", "make an admin account with this name (its password is printed once)")
	useTLS := fs.Bool("tls", false, "HTTPS with a self-signed certificate (made once, next to the database)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*db), 0o755); err != nil {
		return err
	}
	store, err := testserver.Open(*db)
	if err != nil {
		return err
	}
	defer store.Close()
	fmt.Fprintf(c.out, "Database: %s\n", *db)
	if *admin != "" {
		if err := makeAdmin(c, store, *admin); err != nil {
			return err
		}
	}
	if users, _ := store.Users(); len(users) == 0 {
		fmt.Fprintln(c.out, "There are no accounts yet: start with -admin NAME to make the first.")
	}

	// the web version next to the API, for students without Recuerdo
	var web http.Handler
	dirs := webDirs()
	if fs.NArg() > 0 {
		dirs = []string{fs.Arg(0)}
	}
	for _, d := range dirs {
		if web, err = webHandler(d); err == nil {
			break
		}
	}
	if web == nil {
		web = http.NotFoundHandler()
		fmt.Fprintln(c.out, "No web version found (scripts/build-web.sh); the API works without it.")
	}

	api := &testserver.API{Store: store, Version: version.Version, Secure: *useTLS}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: api.Handler(web), ReadHeaderTimeout: 10 * time.Second}
	if *useTLS {
		cert, key, fingerprint, err := testserver.Certificate(filepath.Dir(*db))
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Recuerdo test server: https://%s/  (Ctrl+C stops it)\n", ln.Addr())
		fmt.Fprintf(c.out, "Certificate fingerprint (SHA-256), to check when connecting:\n  %s\n", fingerprint)
		err = srv.ServeTLS(ln, cert, key)
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
	fmt.Fprintf(c.out, "Recuerdo test server: http://%s/  (Ctrl+C stops it)\n", ln.Addr())
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// makeAdmin makes an admin account with a random password, printed once.
func makeAdmin(c *ctx, store *testserver.Store, name string) error {
	password := randomPassword(14)
	u, err := store.CreateUser(name, testserver.Admin, password)
	if errors.Is(err, testserver.ErrExists) {
		fmt.Fprintf(c.out, "The account %q already exists.\n", name)
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "Admin account %q made. Its password (shown only now): %s\n", u.Name, password)
	return nil
}

// randomPassword is n characters that are easy to read and type.
func randomPassword(n int) string {
	const chars = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		k, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[k.Int64()]
	}
	return string(b)
}
