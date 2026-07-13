//go:build darwin

package main

import (
	"os"
	"syscall"
	"time"
)

func GetBirthTime(fileInfo os.FileInfo) time.Time {
	sysType := fileInfo.Sys()
	if stat, ok := sysType.(*syscall.Stat_t); ok {
		return time.Unix(stat.Birthtimespec.Sec, stat.Birthtimespec.Nsec)
	}
	return time.Time{}
}
