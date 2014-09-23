package main

import "io"
import "archive/tar"
import "path"
import "encoding/json"
import "os"

import "github.com/docker/docker/image"

type Image struct {
	image.Image
	Tarball string
}

func NewImage(img *image.Image) *Image {
	if img == nil {
		return &Image{}
	}
	return &Image{*img, ""}
}

func NewImageJSON(json []byte) (*Image, error) {
	img, err := image.NewImgJSON(json)
	if err != nil {
		return nil, err
	}
	return NewImage(img), nil
}

// Deep clone by JSON roundtrip
func (img *Image) Clone() *Image {
	json, err := json.Marshal(img.Image)
	if err != nil {
		panic(err)
	}
	newimg, err := NewImageJSON(json)
	if err != nil {
		panic(err)
	}
	return newimg
}

func (img *Image) WriteTo(w io.Writer) (n int64, err error) {
	c := NewCounter(w)
	tw := tar.NewWriter(c)
	defer tw.Close()

	if err := WriteTarDir(tw, img.ID); err != nil {
		return c.Bytes, err
	}
	WriteTarFile(tw, path.Join(img.ID, "VERSION"), []byte("1.0"))
	if json_bb, err := json.Marshal(img.Image); err != nil {
		return c.Bytes, err
	} else {
		if err := WriteTarFile(tw, path.Join(img.ID, "json"), json_bb); err != nil {
			return c.Bytes, err
		}
	}

	fi, err := os.Stat(img.Tarball)
	if err != nil {
		return c.Bytes, err
	}

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return c.Bytes, err
	}

	hdr.Name = path.Join(img.ID, "layer.tar")

	err = tw.WriteHeader(hdr)
	if err != nil {
		return c.Bytes, err
	}

	if lf, err := os.Open(img.Tarball); err != nil {
		return c.Bytes, err
	} else {
		io.Copy(tw, lf)
		lf.Close()
	}

	tw.Close()
	return c.Bytes, nil
}
