package hookfs

import (
	"path/filepath"
	"time"

	// log "github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func newHookServer(hookfs *HookFs) (*fuse.Server, error) {
	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}
	pathFs := pathfs.NewPathNodeFs(hookfs, nil)
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	originalAbs, _ := filepath.Abs(hookfs.Original)
	mOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       hookfs.FsName,
		FsName:     originalAbs,
	}
	server, err := fuse.NewServer(conn.RawFS(), hookfs.Mountpoint, mOpts)
	if err != nil {
		return nil, err
	}

	if LogLevel() == LogLevelMax {
		server.SetDebug(true)
	}

	return server, nil
}
