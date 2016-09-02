package hookfs

// Base interface for user-written hooks.
//
// You have to implement HookXXX (e.g. HookOnOpen, HookOnRead, HookOnWrite, ..) interfaces.
type Hook interface{}

// Context objects for interaction between prehooks and posthooks.
type HookContext interface{}

// Called on mount. This also implements Hook.
type HookWithInit interface {
	Init() (err error)
}

// Called on open. This also implements Hook.
type HookOnOpen interface {
	// if hooked is true, the real open() would not be called
	PreOpen(path string, flags uint32) (err error, hooked bool, ctx HookContext)
	PostOpen(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}

// Called on read. This also implements Hook.
type HookOnRead interface {
	// if hooked is true, the real read() would not be called
	PreRead(path string, length int64, offset int64) (buf []byte, err error, hooked bool, ctx HookContext)
	PostRead(realRetCode int32, realBuf []byte, prehookCtx HookContext) (buf []byte, err error, hooked bool)
}

// Called on write. This also implements Hook.
// BUG(AkihiroSuda): HookOnWrite is not yet implemented.
type HookOnWrite interface {
	// if hooked is true, the real write() would not be called
	PreWrite(path string, buf []byte, offset int64) (err error, hooked bool, ctx HookContext)
	PostWrite(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}

// Called on mkdir. This also implements Hook.
type HookOnMkdir interface {
	// if hooked is true, the real mkdir() would not be called
	PreMkdir(path string, mode uint32) (err error, hooked bool, ctx HookContext)
	PostMkdir(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}

// Called on rmdir. This also implements Hook.
type HookOnRmdir interface {
	// if hooked is true, the real rmdir() would not be called
	PreRmdir(path string) (err error, hooked bool, ctx HookContext)
	PostRmdir(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}

// Called on opendir. This also implements Hook.
type HookOnOpenDir interface {
	// if hooked is true, the real opendir() would not be called
	PreOpenDir(path string) (err error, hooked bool, ctx HookContext)
	PostOpenDir(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}

// Called on fsync. This also implements Hook.
type HookOnFsync interface {
	// if hooked is true, the real fsync() would not be called
	PreFsync(path string, flags uint32) (err error, hooked bool, ctx HookContext)
	PostFsync(realRetCode int32, prehookCtx HookContext) (err error, hooked bool)
}
