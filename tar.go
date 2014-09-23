package main

import "archive/tar"
import "log"
import "os"
import "time"

type fakeFile struct {
	path string
	data []byte
}

func newFakeFile(path string, data []byte) *fakeFile {
	return &fakeFile{path, data}
}

// os.FileInfo interface
func (ff *fakeFile) Name() string {
	return ff.path
}

func (ff *fakeFile) Size() int64 {
	return int64(len(ff.data))
}

func (ff *fakeFile) Mode() os.FileMode {
	return os.FileMode(0644)
}

func (ff *fakeFile) ModTime() time.Time {
	return time.Now()
}

func (ff *fakeFile) IsDir() bool {
	return false
}

func (ff *fakeFile) Sys() interface{} {
	return nil
}

func (ff *fakeFile) tarHeader() *tar.Header {
	hdr, err := tar.FileInfoHeader(ff, "")
	if err != nil {
		log.Fatalln(err)
	}
	return hdr
}

func (ff *fakeFile) writeTar(tw *tar.Writer) error {
	if err := tw.WriteHeader(ff.tarHeader()); err != nil {
		return err
	}
	for n := 0; n < len(ff.data); {
		i, err := tw.Write(ff.data[n:])
		if err != nil {
			return err
		}
		n += i
	}
	return nil
}

type fakeDir string

// os.FileInfo interface
func (ff fakeDir) Name() string {
	return string(ff)
}

func (ff fakeDir) Size() int64 {
	return 0
}

func (ff fakeDir) Mode() os.FileMode {
	return os.FileMode(0755) | os.ModeDir
}

func (ff fakeDir) ModTime() time.Time {
	return time.Now()
}

func (ff fakeDir) IsDir() bool {
	return true
}

func (ff fakeDir) Sys() interface{} {
	return nil
}

func (ff fakeDir) tarHeader() *tar.Header {
	hdr, err := tar.FileInfoHeader(ff, "")
	if err != nil {
		log.Fatalln(err)
	}
	return hdr
}

func (ff fakeDir) writeTar(tw *tar.Writer) error {
	if err := tw.WriteHeader(ff.tarHeader()); err != nil {
		return err
	}
	return nil
}

func WriteTarFile(tw *tar.Writer, path string, data []byte) error {
	return newFakeFile(path, data).writeTar(tw)
}

func WriteTarDir(tw *tar.Writer, path string) error {
	return fakeDir(path).writeTar(tw)
}
