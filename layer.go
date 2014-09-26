package main

import "sort"
import "strings"
import "archive/tar"
import "io"
import "io/ioutil"
import "os"
import "path"

type LayerFile struct {
	*tar.Header
	Path string
}

type Layer struct {
	Files   map[string]LayerFile
	Workdir string
}

func NewLayer(workdir string) *Layer {
	lr := Layer{}
	lr.Files = make(map[string]LayerFile)
	if workdir == "" {
		lr.Workdir = os.TempDir()
	} else {
		lr.Workdir = workdir
	}
	return &lr
}

func (lr *Layer) Apply(tarball_path string) error {
	Debug("apply", path.Base(tarball_path))
	inf, err := os.Open(tarball_path)
	if err != nil {
		return err
	}
	defer inf.Close()

	tr := tar.NewReader(inf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		basename := path.Base(hdr.Name)

		switch {
		case hdr.FileInfo().IsDir():
			// Directories are always additions; directory removal is a
			// file. Key for a directory entry doesn't include trailing
			// slash to make it easier with deletions.
			Debug("`- mkdir", hdr.Name)
			lr.Files[hdr.Name[:len(hdr.Name)-1]] = LayerFile{hdr, ""}
		case strings.HasPrefix(basename, ".wh..wh."):
			// Aufs' magic ".wh..wh." files should be added rather than
			// treated as deletion
			fallthrough
		default:
			// File addition / overwrite
			tmpf, err := ioutil.TempFile(lr.Workdir, basename+".")
			if err != nil {
				return err
			}

			_, err = io.Copy(tmpf, tr)
			tmpf.Close()
			if err != nil {
				return err
			}

			Debug("`- add  ", hdr.Name, "->", tmpf.Name())
			lr.Files[hdr.Name] = LayerFile{hdr, tmpf.Name()}
		case strings.HasPrefix(basename, ".wh."):
			// File or directory deletion
			lf, exists := lr.Files[path.Join(path.Dir(hdr.Name), basename[4:])]
			switch {
			case !exists:
				// File did not exist in previous layers
				Debug("`- del  ", hdr.Name)
				lr.Files[hdr.Name] = LayerFile{hdr, ""}
			case lf.FileInfo().IsDir():
				// Remove directory
				Debug("`- rm -r", lf.Name)
				for entry, elf := range lr.Files {
					if strings.HasPrefix(elf.Name, lf.Name) {
						delete(lr.Files, entry)
						Debug("  `- rm", entry)
					}
				}
			default:
				// Remove file
				Debug("`- rm  ", lf.Name, "//", hdr.Name)
				delete(lr.Files, lf.Name)
			}
		}
	}
	return nil
}

func (lr *Layer) Size() int64 {
	size := int64(0)
	for _, lf := range lr.Files {
		size += lf.Header.Size
	}
	return size
}

func (lr *Layer) Save(w io.Writer) error {
	// Sort files for sanity
	files := make([]string, len(lr.Files))
	i := 0
	for k := range lr.Files {
		files[i] = k
		i++
	}
	sort.Strings(files)

	tw := tar.NewWriter(w)
	defer tw.Close()

	for _, fn := range files {
		lf := lr.Files[fn]
		if err := tw.WriteHeader(lf.Header); err != nil {
			return err
		}
		if lf.Path != "" {
			copier := func(r io.Reader) error {
				_, err := io.Copy(tw, r)
				return err
			}
			if err := WithOpen(lf.Path, copier); err != nil {
				return err
			}
		}
	}
	return nil
}
