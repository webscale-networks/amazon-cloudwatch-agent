// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

//go:build linux || darwin || freebsd || netbsd || openbsd
// +build linux darwin freebsd netbsd openbsd

package tail

import (
	"os"
)

func OpenFile(name string) (file *os.File, err error) {
	fileInfo, err := os.Stat(name)
	if err == nil && (fileInfo.Mode() & os.ModeNamedPipe) != 0 {
		// Open named pipes in read+write mode to avoid getting end-of-file after
		// every read when the writer to the pipe is transient.
		return os.OpenFile(name, os.O_RDWR, 0);
	}
	return os.Open(name)
}
