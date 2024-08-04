package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/galqiwi/fair-p/internal/utils"
	"io"
	"os"
	"path/filepath"
	"time"
)

func main() {
	err := Main()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func Main() error {
	args, err := getArgs()
	if err != nil {
		return err
	}

	if utils.FileExists(args.logPath) {
		err = archiveFile(args.logPath)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(args.logPath)
	if err != nil {
		return fmt.Errorf("failed to open %q: %s\n", args.logPath, err)
	}

	maxSizeBytes := args.maxSize * 1024 * 1024

	remaining := maxSizeBytes
	s := bufio.NewScanner(os.Stdin)

	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

	for s.Scan() {
		if int64(len(s.Text())) > remaining {
			err := f.Sync()
			if err != nil {
				return fmt.Errorf("failed to sync: %s\n", err)
			}

			err = f.Close()
			if err != nil {
				return fmt.Errorf("error closing: %s\n", err)
			}

			err = archiveFile(args.logPath)
			if err != nil {
				return fmt.Errorf("error archiving: %s\n", err)
			}

			f, err = os.Create(args.logPath)
			if err != nil {
				return fmt.Errorf("error creating new file: %s\n", err)
			}

			remaining = maxSizeBytes
		}
		n, err := fmt.Fprintln(f, s.Text())
		if err != nil {
			return fmt.Errorf("failed to write: %s\n", err)
		}
		remaining -= int64(n)
	}
	if s.Err() != nil {
		return fmt.Errorf("scanner failed with %q\n", s.Err())
	}

	return nil
}

func archiveFile(filename string) error {
	stamp := time.Now().Format("2006-01-02_15:04:05")
	ext := filepath.Ext(filename)
	newNamePrefix := fmt.Sprintf("%s_%s%s", filename[:len(filename)-len(ext)], stamp, ext)

	newName := newNamePrefix

	for i := 1; utils.FileExists(newName); i++ {
		newName = fmt.Sprintf("%s.%v", newNamePrefix, i)
	}

	err := os.Rename(filename, newName)
	if err != nil {
		return err
	}

	return compressFileInplace(newName)
}

func compressFileInplace(filename string) error {
	sourceFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer sourceFile.Close()

	destinationFilename := filename + ".gz"
	destinationFile, err := os.Create(destinationFilename)
	if err != nil {
		return fmt.Errorf("could not create destination file: %v", err)
	}
	defer destinationFile.Close()

	gzipWriter := gzip.NewWriter(destinationFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		return fmt.Errorf("could not copy data to gzip writer: %v", err)
	}

	return os.Remove(filename)
}
