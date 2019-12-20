package main

/*
 * nosavefs.go
 * WebDAV filesystem which doesn't actually save files
 * By J. Stuart McMurray
 * Created 20191219
 * Last Modified 20191219
 */

import (
	"context"
	"os"

	"golang.org/x/net/webdav"
)

/* noSaveFS wraps a memdav.FileSystem but doesn't allow real writes */
type noSaveFS struct {
	fs webdav.FileSystem
}

// NewNoSaveFS wraps fs but writes 0's to files instead of whatever should
// be written.
func NewNoSaveFS(fs webdav.FileSystem) webdav.FileSystem {
	return noSaveFS{fs}
}

func (n noSaveFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return n.fs.Mkdir(ctx, name, perm)
}

// OpenFile returns a file to which any write will really only write 0's
func (n noSaveFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := n.fs.OpenFile(ctx, name, flag, perm)
	if nil != err {
		return nil, err
	}
	return noSaveFile{f}, nil
}
func (n noSaveFS) RemoveAll(ctx context.Context, name string) error {
	return n.fs.RemoveAll(ctx, name)
}
func (n noSaveFS) Rename(ctx context.Context, oldName, newName string) error {
	return n.fs.Rename(ctx, oldName, newName)
}
func (n noSaveFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return n.fs.Stat(ctx, name)
}

/* noSaveFile wraps memdav.File but writes 0's */
type noSaveFile struct {
	f webdav.File
}

func (n noSaveFile) Close() error                                 { return n.f.Close() }
func (n noSaveFile) Read(p []byte) (rn int, err error)            { return n.f.Read(p) }
func (n noSaveFile) Seek(offset int64, whence int) (int64, error) { return n.f.Seek(offset, whence) }
func (n noSaveFile) Readdir(count int) ([]os.FileInfo, error)     { return n.f.Readdir(count) }
func (n noSaveFile) Stat() (os.FileInfo, error)                   { return n.f.Stat() }

// Write writes all 0's
func (n noSaveFile) Write(p []byte) (wn int, err error) {
	b := make([]byte, len(p))
	return n.f.Write(b)
}
