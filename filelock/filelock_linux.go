//go:build linux

package filelock

import (
	"os"
	"syscall"
)

func lock(f *os.File, lt lockType) error {
	lc := syscall.LOCK_SH
	if lt == WriteLock {
		lc = syscall.LOCK_EX
	}

	err := syscall.Flock(int(f.Fd()), lc|syscall.LOCK_NB)
	if err != nil {
		if errNo, ok := err.(syscall.Errno); ok && errNo == 0xb {
			return ErrFileLock // 找到文件被锁错误码,返回自定义错误
		}
	}
	return err
}

func unlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
