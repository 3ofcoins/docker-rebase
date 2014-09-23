package main

import "sort"
import "strings"
import "archive/tar"
import "io"
import "io/ioutil"
import "os"
import "path"

type LayerFile struct {
	Header *tar.Header
	Path   string
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
			Debug("`- mkdir", hdr.Name)
			lr.Files[hdr.Name] = LayerFile{hdr, ""}
		case strings.HasPrefix(basename, ".wh..wh."):
			fallthrough
		default:
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
			del := path.Join(path.Dir(hdr.Name), basename[4:])
			if _, isfile := lr.Files[del]; isfile {
				Debug("`- rm  ", del, "//", hdr.Name)
				delete(lr.Files, del)
			} else {
				if _, isdir := lr.Files[del+"/"]; isdir {
					Debug("`- rm -r", del)
					for entry := range lr.Files {
						if strings.HasPrefix(entry, del) {
							delete(lr.Files, entry)
							Debug("  `- rm", entry)
						}
					}
				} else {
					Debug("`- del  ", hdr.Name)
					lr.Files[hdr.Name] = LayerFile{hdr, ""}
				}
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
