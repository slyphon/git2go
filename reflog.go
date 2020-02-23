package git

/*
#include <git2.h>
#include <git2/types.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type Reflog struct {
	ptr *C.git_reflog
	repo *Repository
	// it doesn't seem to be possible to access this on the git_reflog struct
	// via cgo
	name string
}

type ReflogCollection struct {
	repo *Repository
}


func (c *ReflogCollection) Read(name string) (*Reflog, error) {
	reflog := new(Reflog)
	reflog.repo = c.repo
	reflog.name = name

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_reflog_read(&reflog.ptr, c.repo.ptr, cname); ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.KeepAlive(c)

	return reflog, nil
}

func newReflogFromC(ptr *C.git_reflog, repo *Repository, name string) *Reflog {
	ref := &Reflog{
		ptr:  ptr,
		repo: repo,
		name: name,
	}
	runtime.SetFinalizer(ref, (*Reflog).Free)
	return ref
}

func (r *Reflog) Rename(newName string) (*Reflog, error) {
	cold := C.CString(r.name)
	defer C.free(unsafe.Pointer(cold))

	cnew := C.CString(newName)
	defer C.free(unsafe.Pointer(cnew))

	ret := C.git_reflog_rename(r.repo.ptr, cold, cnew)
	runtime.KeepAlive(r)
	runtime.KeepAlive(r.repo)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newReflogFromC(r.ptr, r.repo, newName), nil
}

func (r *Reflog) Delete() error {
	cname := C.CString(r.name)
	defer C.free(unsafe.Pointer(cname))

	ret := C.git_reflog_delete(r.repo.ptr, cname)
	runtime.KeepAlive(r)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (r *Reflog) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_reflog_free(r.ptr)
}

func (r *Reflog) Append(oid *Oid, who *Signature, msg string) error {
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))

	cwho, err := who.toC()
	if err != nil {
		return err
	}
	defer C.git_signature_free(cwho)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_reflog_append(r.ptr, oid.toC(), cwho, cmsg)
	runtime.KeepAlive(r)
	runtime.KeepAlive(cwho)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (r *Reflog) EntryCount() uint64 {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	size := C.git_reflog_entrycount(r.ptr)
	runtime.KeepAlive(r)
	return uint64(size)
}
