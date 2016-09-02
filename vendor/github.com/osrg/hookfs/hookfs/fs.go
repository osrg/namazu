package hookfs

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

// HookFS object
type HookFs struct {
	Original   string
	Mountpoint string
	FsName     string
	fs         pathfs.FileSystem
	hook       Hook
}

// Instantiate a new HookFS object
func NewHookFs(original string, mountpoint string, hook Hook) (*HookFs, error) {
	log.WithFields(log.Fields{
		"original":   original,
		"mountpoint": mountpoint,
	}).Debug("Hooking a fs")

	loopbackfs := pathfs.NewLoopbackFileSystem(original)
	hookfs := &HookFs{
		Original:   original,
		Mountpoint: mountpoint,
		FsName:     "hookfs",
		fs:         loopbackfs,
		hook:       hook,
	}
	return hookfs, nil
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) String() string {
	return fmt.Sprintf("HookFs{Original=%s, Mountpoint=%s, FsName=%s, Underlying fs=%s, hook=%s}",
		this.Original, this.Mountpoint, this.FsName, this.fs.String(), this.hook)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) SetDebug(debug bool) {
	this.fs.SetDebug(debug)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return this.fs.GetAttr(name, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Chmod(name string, mode uint32, context *fuse.Context) fuse.Status {
	return this.fs.Chmod(name, mode, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Chown(name string, uid uint32, gid uint32, context *fuse.Context) fuse.Status {
	return this.fs.Chown(name, uid, gid, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) fuse.Status {
	return this.fs.Utimens(name, Atime, Mtime, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	return this.fs.Truncate(name, size, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Access(name string, mode uint32, context *fuse.Context) fuse.Status {
	return this.fs.Access(name, mode, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Link(oldName string, newName string, context *fuse.Context) fuse.Status {
	return this.fs.Link(oldName, newName, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	hook, hookEnabled := this.hook.(HookOnMkdir)
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookErr, prehooked, prehookCtx = hook.PreMkdir(name, mode)
		if prehooked {
			log.WithFields(log.Fields{
				"this":       this,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("Mkdir: Prehooked")
			if prehookErr == nil {
				log.WithFields(log.Fields{
					"this":       this,
					"prehookErr": prehookErr,
					"prehookCtx": prehookCtx,
				}).Fatal("Mkdir is prehooked, but did not returned an error. this is very strange.")
			}
			return fuse.ToStatus(prehookErr)
		}
	}

	lowerCode := this.fs.Mkdir(name, mode, context)
	if hookEnabled {
		posthookErr, posthooked = hook.PostMkdir(int32(lowerCode), prehookCtx)
		if posthooked {
			log.WithFields(log.Fields{
				"this":        this,
				"posthookErr": posthookErr,
			}).Debug("Mkdir: Posthooked")
			return fuse.ToStatus(posthookErr)
		}
	}

	return lowerCode
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	return this.fs.Mknod(name, mode, dev, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Rename(oldName string, newName string, context *fuse.Context) fuse.Status {
	return this.fs.Rename(oldName, newName, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Rmdir(name string, context *fuse.Context) fuse.Status {
	hook, hookEnabled := this.hook.(HookOnRmdir)
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookErr, prehooked, prehookCtx = hook.PreRmdir(name)
		if prehooked {
			log.WithFields(log.Fields{
				"this":       this,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("Rmdir: Prehooked")
			if prehookErr == nil {
				log.WithFields(log.Fields{
					"this":       this,
					"prehookErr": prehookErr,
					"prehookCtx": prehookCtx,
				}).Fatal("Rmdir is prehooked, but did not returned an error. this is very strange.")
			}
			return fuse.ToStatus(prehookErr)
		}
	}

	lowerCode := this.fs.Rmdir(name, context)
	if hookEnabled {
		posthookErr, posthooked = hook.PostRmdir(int32(lowerCode), prehookCtx)
		if posthooked {
			log.WithFields(log.Fields{
				"this":        this,
				"posthookErr": posthookErr,
			}).Debug("Mkdir: Posthooked")
			return fuse.ToStatus(posthookErr)
		}
	}

	return lowerCode
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Unlink(name string, context *fuse.Context) fuse.Status {
	return this.fs.Unlink(name, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) GetXAttr(name string, attribute string, context *fuse.Context) ([]byte, fuse.Status) {
	return this.fs.GetXAttr(name, attribute, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) ListXAttr(name string, context *fuse.Context) ([]string, fuse.Status) {
	return this.fs.ListXAttr(name, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	return this.fs.RemoveXAttr(name, attr, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	return this.fs.SetXAttr(name, attr, data, flags, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) OnMount(nodeFs *pathfs.PathNodeFs) {
	this.fs.OnMount(nodeFs)
	hook, hookEnabled := this.hook.(HookWithInit)
	if hookEnabled {
		err := hook.Init()
		if err != nil {
			log.Error(err)
			log.Warn("Disabling hook")
			this.hook = nil
		}
	}
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) OnUnmount() {
	this.fs.OnUnmount()
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	hook, hookEnabled := this.hook.(HookOnOpen)
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookErr, prehooked, prehookCtx = hook.PreOpen(name, flags)
		if prehooked {
			log.WithFields(log.Fields{
				"this":       this,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("Open: Prehooked")
			if prehookErr == nil {
				log.WithFields(log.Fields{
					"this":       this,
					"prehookErr": prehookErr,
					"prehookCtx": prehookCtx,
				}).Fatal("Open is prehooked, but did not returned an error. this is very strange.")
			}
			return nil, fuse.ToStatus(prehookErr)
		}
	}

	lowerFile, lowerCode := this.fs.Open(name, flags, context)
	hFile, hErr := newHookFile(lowerFile, name, this.hook)
	if hErr != nil {
		log.WithField("error", hErr).Panic("NewHookFile() should not cause an error")
	}

	if hookEnabled {
		posthookErr, posthooked = hook.PostOpen(int32(lowerCode), prehookCtx)
		if posthooked {
			log.WithFields(log.Fields{
				"this":        this,
				"posthookErr": posthookErr,
			}).Debug("Open: Posthooked")
			return hFile, fuse.ToStatus(posthookErr)
		}
	}

	return hFile, lowerCode
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	lowerFile, lowerCode := this.fs.Create(name, flags, mode, context)
	hFile, hErr := newHookFile(lowerFile, name, this.hook)
	if hErr != nil {
		log.WithField("error", hErr).Panic("NewHookFile() should not cause an error")
	}
	return hFile, lowerCode
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	hook, hookEnabled := this.hook.(HookOnOpenDir)
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookErr, prehooked, prehookCtx = hook.PreOpenDir(name)
		if prehooked {
			log.WithFields(log.Fields{
				"this":       this,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("OpenDir: Prehooked")
			if prehookErr == nil {
				log.WithFields(log.Fields{
					"this":       this,
					"prehookErr": prehookErr,
					"prehookCtx": prehookCtx,
				}).Fatal("OpenDir is prehooked, but did not returned an error. this is very strange.")
			}
			return nil, fuse.ToStatus(prehookErr)
		}
	}

	lowerEnts, lowerCode := this.fs.OpenDir(name, context)
	if hookEnabled {
		posthookErr, posthooked = hook.PostOpenDir(int32(lowerCode), prehookCtx)
		if posthooked {
			log.WithFields(log.Fields{
				"this":        this,
				"posthookErr": posthookErr,
			}).Debug("OpenDir: Posthooked")
			return lowerEnts, fuse.ToStatus(posthookErr)
		}
	}

	return lowerEnts, lowerCode
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Symlink(value string, linkName string, context *fuse.Context) fuse.Status {
	return this.fs.Symlink(value, linkName, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	return this.fs.Readlink(name, context)
}

// implements hanwen/go-fuse/fuse/pathfs.FileSystem. You are not expected to call this manually.
func (this *HookFs) StatFs(name string) *fuse.StatfsOut {
	return this.fs.StatFs(name)
}

// Start the server (blocking).
func (this *HookFs) Serve() error {
	server, err := newHookServer(this)
	if err != nil {
		return err
	}
	server.Serve()
	return nil
}
