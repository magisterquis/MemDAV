MemDAV
======

MemDAV is a simple little WebDAV server which reads and writes files to/from
memory.  It's more or less a wrapper around `golang.org/x/net/webdav` with a
few extra features.

By default only HTTP is served.  With `-https-addr` HTTPS will be served as
well.

Backing Storage
---------------
Files can either be served from a directory (with `-dir`) or from memory (the
default).

By default, files will be saved as they are uploaded.  Instead of the contents
sent by the client, NUL bytes can be written to the saved files with the 
`-no-save` flag.  This is meant to make running a honepot a little less
worrisome.

Requests to delete files (i.e. the `DELETE` verb) will be honored unless
`-no-delete` is set.  This was originally written in response to a DLP solution
which prevented exfil over WebDAV by sending `DELETE` requests after files were
`PUT` on servers on the internet.

Serving a Single File
---------------------
Instead of serving files which have been uploaded or are in the served
directory, a single file can be returned for all GET requests with
`-serve-file`.  This is meant to make running a honeypot a little less
worrisome.  Originally it was written to demonstrate that file paths and file
content aren't necessarily related.  Silly blue team...





For legal use only.
