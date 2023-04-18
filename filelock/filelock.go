package filelock

import (
	"errors"
	"os"
)

type lockType uint32

const (
	ReadLock  lockType = 1
	WriteLock lockType = 2
)

var ErrFileLock = errors.New("file is lock")

// Lock 排它锁锁住文件
func Lock(f *os.File) error { return lock(f, WriteLock) }

// RLock 共享锁锁住文件
func RLock(f *os.File) error { return lock(f, ReadLock) }

// Unlock 释放文件锁
func Unlock(f *os.File) error { return unlock(f) }

type File struct {
	File *os.File
}

func LockOpenFile(name string, flag int, perm os.FileMode, lt lockType) (*File, error) {
	fr, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	if err = lock(fr, lt); err != nil {
		_ = fr.Close()
		return nil, err
	}
	return &File{File: fr}, nil
}

func (f *File) Close() error {
	err := unlock(f.File)
	if closeErr := f.File.Close(); err == nil {
		err = closeErr
	}
	return err
}
