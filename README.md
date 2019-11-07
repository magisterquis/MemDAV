MemDAV
======

MemDAV is a simple little WebDAV server which reads and writes files to/from
memory.  It's more or less a wrapper around `golang.org/x/net/webdav` with the
one feature that DELETE requests can be ignored but returned a 200.

By default only HTTP is served.  With `-https-addr` HTTPS will be served as
well.

Files can be served from disk as well as memory with the `-dir` flag.

For legal use only.
