// Copyright (c) 2021 Apptainer a Series of LF Projects LLC
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// Copyright (c) 2018-2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package image

import (
	"os"
)

type sandboxFormat struct{}

func (f *sandboxFormat) initializer(img *Image, fi os.FileInfo) error {
	if fi.IsDir() {
		img.Type = SANDBOX
	} else {
		return debugError("not a directory image")
	}
	img.Partitions = []Section{
		{
			Type:         SANDBOX,
			ID:           1,
			Name:         RootFs,
			AllowedUsage: RootFsUsage | OverlayUsage | DataUsage,
		},
	}
	return nil
}

func (f *sandboxFormat) openMode(writable bool) int {
	return os.O_RDONLY
}

func (f *sandboxFormat) lock(img *Image) error {
	return nil
}
