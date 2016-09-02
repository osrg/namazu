package hookfs

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"time"
)

type hookFile struct {
	file nodefs.File
	name string
	hook Hook
}

func newHookFile(file nodefs.File, name string, hook Hook) (*hookFile, error) {
	log.WithFields(log.Fields{
		"file": file,
		"name": name,
	}).Debug("Hooking a file")

	hookfile := &hookFile{
		file: file,
		name: name,
		hook: hook,
	}
	return hookfile, nil
}

// implements nodefs.File
func (this *hookFile) SetInode(inode *nodefs.Inode) {
	this.file.SetInode(inode)
}

// implements nodefs.File
func (this *hookFile) String() string {
	return fmt.Sprintf("HookFile{file=%s, name=%s}", this.file.String(), this.name)
}

// implements nodefs.File
func (this *hookFile) InnerFile() nodefs.File {
	return this.file.InnerFile()
}

// implements nodefs.File
func (this *hookFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	hook, hookEnabled := this.hook.(HookOnRead)
	var prehookBuf, posthookBuf []byte
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookBuf, prehookErr, prehooked, prehookCtx = hook.PreRead(this.name, int64(len(dest)), off)
		if prehooked {
			log.WithFields(log.Fields{
				"this": this,
				// "prehookBuf": prehookBuf,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("Read: Prehooked")
			return fuse.ReadResultData(prehookBuf), fuse.ToStatus(prehookErr)
		}
	}

	lowerRR, lowerCode := this.file.Read(dest, off)
	if hookEnabled {
		lowerRRBuf, lowerRRBufStatus := lowerRR.Bytes(make([]byte, lowerRR.Size()))
		if lowerRRBufStatus != fuse.OK {
			log.WithField("error", lowerRRBufStatus).Panic("lowerRR.Bytes() should not cause an error")
		}
		posthookBuf, posthookErr, posthooked = hook.PostRead(int32(lowerCode), lowerRRBuf, prehookCtx)
		if posthooked {
			if len(posthookBuf) != len(lowerRRBuf) {
				log.WithFields(log.Fields{
					"this": this,
					// "posthookBuf": posthookBuf,
					"posthookErr":    posthookErr,
					"posthookBufLen": len(posthookBuf),
					"lowerRRBufLen":  len(lowerRRBuf),
					"destLen":        len(dest),
				}).Warn("Read: Posthooked, but posthookBuf length != lowerrRRBuf length. You may get a strange behavior.")
			}

			log.WithFields(log.Fields{
				"this": this,
				// "posthookBuf": posthookBuf,
				"posthookErr": posthookErr,
			}).Debug("Read: Posthooked")
			return fuse.ReadResultData(posthookBuf), fuse.ToStatus(posthookErr)
		}
	}

	return lowerRR, lowerCode
}

// implements nodefs.File
func (this *hookFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	return this.file.Write(data, off)
}

// implements nodefs.File
func (this *hookFile) Flush() fuse.Status {
	return this.file.Flush()
}

// implements nodefs.File
func (this *hookFile) Release() {
	this.file.Release()
}

// implements nodefs.File
func (this *hookFile) Fsync(flags int) fuse.Status {
	hook, hookEnabled := this.hook.(HookOnFsync)
	var prehookErr, posthookErr error
	var prehooked, posthooked bool
	var prehookCtx HookContext

	if hookEnabled {
		prehookErr, prehooked, prehookCtx = hook.PreFsync(this.name, uint32(flags))
		if prehooked {
			log.WithFields(log.Fields{
				"this":       this,
				"prehookErr": prehookErr,
				"prehookCtx": prehookCtx,
			}).Debug("Fsync: Prehooked")
			return fuse.ToStatus(prehookErr)
		}
	}

	lowerCode := this.file.Fsync(flags)
	if hookEnabled {
		posthookErr, posthooked = hook.PostFsync(int32(lowerCode), prehookCtx)
		if posthooked {
			log.WithFields(log.Fields{
				"this":        this,
				"posthookErr": posthookErr,
			}).Debug("Fsync: Posthooked")
			return fuse.ToStatus(posthookErr)
		}
	}

	return lowerCode
}

// implements nodefs.File
func (this *hookFile) Truncate(size uint64) fuse.Status {
	return this.file.Truncate(size)
}

// implements nodefs.File
func (this *hookFile) GetAttr(out *fuse.Attr) fuse.Status {
	return this.file.GetAttr(out)
}

// implements nodefs.File
func (this *hookFile) Chown(uid uint32, gid uint32) fuse.Status {
	return this.file.Chown(uid, gid)
}

// implements nodefs.File
func (this *hookFile) Chmod(perms uint32) fuse.Status {
	return this.file.Chmod(perms)
}

// implements nodefs.File
func (this *hookFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {
	return this.file.Utimens(atime, mtime)
}

// implements nodefs.File
func (this *hookFile) Allocate(off uint64, size uint64, mode uint32) fuse.Status {
	return this.file.Allocate(off, size, mode)
}
