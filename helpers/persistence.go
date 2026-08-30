// Package helpers contains helper methods
package helpers

import (
	"os"
	"path/filepath"
)

func WriteToFile(data []byte, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(int64(file.Fd()))
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// CreateFile creates a file from the file path, it also creates all the
// directories to the file if they do not exist..
func CreateFile(fp string) error {
	err := CreateDir(filepath.Dir(fp))
	if err != nil {
		return err
	}

	file, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func CreateDir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0o777)
	if err != nil {
		return err
	}

	return nil
}

func PathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
