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
	Debug("Applying", tarball_path)
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
			Debug("mkdir", hdr.Name)
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

			Debug("add  ", hdr.Name, "->", tmpf.Name())
			lr.Files[hdr.Name] = LayerFile{hdr, tmpf.Name()}
		case strings.HasPrefix(basename, ".wh."):
			del := path.Join(path.Dir(hdr.Name), basename[4:])
			if _, isfile := lr.Files[del]; isfile {
				Debug("rm   ", del, "//", hdr.Name)
				delete(lr.Files, del)
			} else {
				del += "/"
				Debug("rm -r", del)
				for entry := range lr.Files {
					if strings.HasPrefix(entry, del) {
						delete(lr.Files, entry)
						Debug("`- rm", entry)
					}
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

func (lr *Layer) WriteTo(w io.Writer) (n int64, err error) {
	// Sort files for sanity
	files := make([]string, len(lr.Files))
	i := 0
	for k := range lr.Files {
		files[i] = k
		i++
	}
	sort.Strings(files)

	c := NewCounter(w)
	tw := tar.NewWriter(c)
	defer tw.Close()

	for _, fn := range files {
		lf := lr.Files[fn]
		err = tw.WriteHeader(lf.Header)
		if err != nil {
			goto exit
		}
		if lf.Path != "" {
			df, err := os.Open(lf.Path)
			if err != nil {
				goto exit
			}
			_, err = io.Copy(tw, df)
			if err != nil {
				goto exit
			}
			df.Close()
		}
	}

exit:
	tw.Close()
	return c.Bytes, err
}
