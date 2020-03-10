// Program MemDAV is a WebDAV server which stores files in memory
package main

/*
 * memdav.go
 * WebDAV server which stores files in memory
 * By J. Stuart McMurray
 * Created 20191105
 * Last Modified 20200306
 */

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/net/webdav"
)

var (
	defaultUser = ""
	defaultPass = ""
)

/* server is a WebDAV handler */
type server struct {
	noDelete  bool /* Don't actually handle DELETE */
	readOnly  bool
	w         *webdav.Handler /* Wrapped WebDAV handler */
	serveFile string          /* Single file to serve */
	username  string          /* HTTP Username */
	password  string          /* HTTP Password */
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
		unixAddr = flag.String(
			"listen-unix",
			"",
			"Unix domain listen `address`",
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
		readOnly = flag.Bool(
			"read-only",
			false,
			"Allow only requests which read files",
		)
		username = flag.String(
			"username",
			defaultUser,
			"Optional HTTP basic auth `username`",
		)
		password = flag.String(
			"password",
			defaultPass,
			"Optional HTTP basic auth `password`",
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
		readOnly: *readOnly,
		w: &webdav.Handler{
			FileSystem: fs,
			LockSystem: webdav.NewMemLS(),
		},
		serveFile: *serveFile,
		username:  *username,
		password:  *password,
	}

	/* Register handler */
	http.HandleFunc("/", s.Handle)

	var (
		ech     = make(chan error)
		serving bool
	)

	/* Listen and serve on all the protocols */
	if "" != *httpsAddr {
		serving = true
		go func() {
			l, err := net.Listen("tcp", *httpsAddr)
			if nil != err {
				ech <- fmt.Errorf(
					"Error listening on %s: %w",
					*httpsAddr,
					err,
				)
				return
			}
			log.Printf("Will serve HTTPS on %s", l.Addr())
			ech <- fmt.Errorf(
				"HTTPS error: %w",
				http.ServeTLS(l, nil, *certFile, *keyFile),
			)
		}()
	}
	if "" != *httpAddr {
		serving = true
		go func() {
			l, err := net.Listen("tcp", *httpAddr)
			if nil != err {
				ech <- fmt.Errorf(
					"Error listening on %s: %w",
					*httpAddr,
					err,
				)
				return
			}
			log.Printf("Will serve HTTP on %s", l.Addr())
			ech <- fmt.Errorf(
				"HTTP error: %w",
				http.Serve(l, nil),
			)
		}()
	}
	if "" != *unixAddr {
		serving = true
		go func() {
			l, err := net.Listen("unix", *unixAddr)
			if nil != err {
				ech <- fmt.Errorf(
					"Error listening on %s: %w",
					*unixAddr,
					err,
				)
				return
			}
			log.Printf("Will serve HTTP on %s", l.Addr())
			ech <- fmt.Errorf(
				"HTTP-over-unix error: %w",
				http.Serve(l, nil),
			)
		}()
	}
	if !serving {
		log.Fatalf("No listen addresses configured")
	}
	log.Fatalf("Error: %v", <-ech)
}

/* handle Handles an HTTP connection */
func (s server) Handle(w http.ResponseWriter, r *http.Request) {
	/* If we have creds set, check them */
	if "" != s.username || "" != s.password {
		w.Header().Set(
			"WWW-Authenticate",
			`Basic realm="Auth Required"`,
		)
		u, p, ok := r.BasicAuth()
		if !ok || ("" == u && "" == p) { /* Client didn't know? */
			log.Printf("[%v] No auth", r.RemoteAddr)
			http.Error(w, "Not authorized", 401)
			return
		}
		if u != s.username || p != s.password {
			log.Printf(
				"[%s] Auth fail (%q / %q)",
				r.RemoteAddr,
				u,
				p,
			)
			http.Error(w, "Not authorized", 401)
			return
		}
	}

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

	/* If we're only allowing read access, whitelist the allowed methods */
	if s.readOnly {
		switch r.Method {
		case "OPTIONS", "GET", "HEAD", "PROPFIND":
			/* These are ok */
		default:
			/* This is not ok */
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
