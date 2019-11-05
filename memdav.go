// Program MemDAV is a WebDAV server which stores files in memory
package main

/*
 * memdav.go
 * WebDAV server which stores files in memory
 * By J. Stuart McMurray
 * Created 20191105
 * Last Modified 20191105
 */

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/webdav"
)

/* server is a WebDAV handler */
type server struct {
	noDelete bool            /* Don't actually handle DELETE */
	w        *webdav.Handler /* Wrapped WebDAV handler */
}

func main() {
	var (
		httpAddr = flag.String(
			"listen-http",
			"0.0.0.0:80",
			"Listen `address`",
		)
		httpsAddr = flag.String(
			"listen-https",
			"",
			"HTTPS listen `address`",
		)
		certFile = flag.String(
			"cert",
			"",
			"TLS certificate `file`",
		)
		keyFile = flag.String(
			"key",
			"",
			"TLS key `file`",
		)
		noDelete = flag.Bool(
			"no-delete",
			false,
			"Do not actually DELETE files",
		)
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %v [options]

Serves WebDAV from memory, not touching the disk.  With -no-delete, requests to
DELETE files will not actually delete the files.

Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	/* WebDAV handler */
	s := server{
		noDelete: *noDelete,
		w: &webdav.Handler{
			FileSystem: webdav.NewMemFS(),
			LockSystem: webdav.NewMemLS(),
		},
	}

	/* Register handler */
	http.HandleFunc("/", s.Handle)

	/* Serve HTTP and maybe HTTPS */
	if "" != *httpsAddr {
		go func() {
			log.Printf("Will serve HTTPS on %v", *httpsAddr)
			log.Fatalf("HTTPS Error: %v", http.ListenAndServeTLS(
				*httpsAddr,
				*certFile,
				*keyFile,
				nil,
			))
		}()
	}
	log.Printf("Will serve HTTP on %v", *httpAddr)
	log.Fatalf("HTTP Error: %v", http.ListenAndServe(*httpAddr, nil))
}

/* handle Handles an HTTP connection */
func (s server) Handle(w http.ResponseWriter, r *http.Request) {
	logReq(r)
	/* If we don't delete, life is easy */
	if s.noDelete && http.MethodDelete == r.Method {
		return
	}

	s.w.ServeHTTP(w, r)
}

func logReq(r *http.Request) {
	log.Printf(
		"[%v] %v %v",
		r.RemoteAddr,
		r.Method,
		r.URL,
	)
}
