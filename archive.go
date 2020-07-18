package main

import (
	"compress/flate"
	"github.com/mholt/archiver/v3"
)

func archiveFolder(src string, dest string) error {
	z := archiver.Zip{
		CompressionLevel:       flate.DefaultCompression,
		MkdirAll:               true,
		SelectiveCompression:   true,
		ContinueOnError:        false,
		OverwriteExisting:      true,
		ImplicitTopLevelFolder: false,
	}

	return z.Archive([]string{src}, dest)
}
