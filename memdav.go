// Program MemDAV is a WebDAV server which stores files in memory
package main

/*
 * memdav.go
 * WebDAV server which stores files in memory
 * By J. Stuart McMurray
 * Created 20191105
 * Last Modified 20191219
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
	noDelete  bool            /* Don't actually handle DELETE */
	w         *webdav.Handler /* Wrapped WebDAV handler */
	serveFile string          /* Single file to serve */
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
		dir = flag.String(
			"dir",
			"",
			"Serve files fron `directory`, not memory",
		)
		serveFile = flag.String(
			"serve-file",
			"",
			"If set, serves `file` for every GET request",
		)
		noSave = flag.Bool(
			"no-save",
			false,
			"Save NULs instead of file contents",
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

	/* Work out the filesystem to use */
	var fs webdav.FileSystem
	if "" != *dir {
		/* Make sure we have the directory */
		if err := os.MkdirAll(*dir, 0700); nil != err {
			log.Fatalf(
				"Unable to make directory %q: %v",
				*dir,
				err,
			)
		}
		fs = webdav.Dir(*dir)
	} else {
		fs = webdav.NewMemFS()
	}
	if *noSave {
		fs = NewNoSaveFS(fs)
	}

	/* WebDAV handler */
	s := server{
		noDelete: *noDelete,
		w: &webdav.Handler{
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
		},
		serveFile: *serveFile,
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

	/* Special cases sometimes */
	switch r.Method {
	case http.MethodDelete: /* We may not allow deletes */
		if s.noDelete {
			return
		}
	case http.MethodGet: /* Maybe serve a single file */
		if "" != s.serveFile {
			http.ServeFile(w, r, s.serveFile)
			return
		}
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
