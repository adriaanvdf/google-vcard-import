package util

import (
	"io/fs"
	"log"
	"os"
)

func ListFilePathsInDir(path *string) ([]string, error) {
	var entries []fs.DirEntry
	var filePaths []string
	dir, err := os.Open(*path)
	if err != nil {
		log.Fatalf("Unable to open path. %v", err)
	}
	defer dir.Close()
	entries, err = dir.ReadDir(0)
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			filePath := dir.Name() + entry.Name()
			filePaths = append(filePaths, filePath)
		}
	}

	return filePaths, err
}