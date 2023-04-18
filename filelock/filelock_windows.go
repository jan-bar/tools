//go:build windows

package filelock

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	reserved = 0
	allBytes = ^uint32(0)
)

var (
	modKernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modKernel32.NewProc("LockFileEx")
	procUnlockFileEx = modKernel32.NewProc("UnlockFileEx")
)

func lock(f *os.File, lt lockType) error {
	var (
		ol = new(syscall.Overlapped)
		lc uint32
	)
	if lt == WriteLock {
		lt = 3 // LOCKFILE_FAIL_IMMEDIATELY | LOCKFILE_EXCLUSIVE_LOCK
	}

	r1, _, e1 := syscall.SyscallN(procLockFileEx.Addr(), f.Fd(),
		uintptr(lc), uintptr(reserved), uintptr(allBytes),
		uintptr(allBytes), uintptr(unsafe.Pointer(ol)))
	if r1 == 0 {
		if e1 != 0 {
			if e1 == 0x21 { // 找到文件被锁错误码,返回自定义错误
				return ErrFileLock
			}
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

func unlock(f *os.File) error {
	ol := new(syscall.Overlapped)
	r1, _, e1 := syscall.SyscallN(procUnlockFileEx.Addr(), f.Fd(),
		uintptr(reserved), uintptr(allBytes), uintptr(allBytes),
		uintptr(unsafe.Pointer(ol)), 0)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}
