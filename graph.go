package main

import "archive/tar"
import "errors"
import "fmt"
import "io"
import "io/ioutil"
import "os"
import "path"
import "strings"
import "time"

import "github.com/docker/docker/daemon/graphdriver"
import "github.com/docker/docker/image"

type Graph struct {
	Images  map[string]*Image // ID -> Image
	Workdir string
}

func NewGraph(workdir string) *Graph {
	if workdir == "" {
		workdir = os.TempDir()
	}
	return &Graph{make(map[string]*Image), workdir}
}

func (gr *Graph) AddImage(img *Image) {
	gr.Images[img.ID] = img
	img.SetGraph(gr)
	img.Tarball = gr.ImageRoot(img.ID)
}

func (gr *Graph) Load(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch path.Base(hdr.Name) {
		case "json":
			json_bb, err := ioutil.ReadAll(tr)
			if err != nil {
				return err
			}
			img, err := NewImageJSON(json_bb)
			if err != nil {
				return err
			}
			gr.AddImage(img)
		case "layer.tar":
			id := path.Base(path.Dir(hdr.Name))
			tar_path := gr.ImageRoot(id)
			if tarf, err := os.Create(tar_path); err != nil {
				return err
			} else {
				_, err := io.Copy(tarf, tr)
				tarf.Close()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// image.Graph interface

func (gr *Graph) Get(id string) (*image.Image, error) {
	if img := gr.Images[id]; img != nil {
		return &img.Image, nil
	} else {
		return nil, errors.New("Not found")
	}
}

func (gr *Graph) ImageRoot(id string) string {
	return path.Join(gr.Workdir, "image."+id+".tar")
}

func (gr *Graph) Driver() graphdriver.Driver {
	return nil
}

//////////

func (gr *Graph) expandID(short_id string) (full_id string, err error) {
	if _, found := gr.Images[short_id]; found {
		return short_id, nil
	}

	for id := range gr.Images {
		if strings.HasPrefix(id, short_id) {
			if full_id != "" && err != nil {
				err = fmt.Errorf("%s: ambiguous prefix", short_id)
			}
			full_id = id
		}
	}

	if full_id == "" {
		err = fmt.Errorf("%s: image not found", short_id)
	}

	return
}

func (gr *Graph) Rebase(id, base_id string) (rimg *Image, err error) {
	if id, err = gr.expandID(id); err != nil {
		return
	}

	if base_id, err = gr.expandID(base_id); err != nil {
		return
	}

	img := gr.Images[id]

	history, err := img.History()
	if err != nil {
		return
	}

	i := len(history) - 1
	for history[i].ID != base_id {
		Debug("Skipping", history[i].ID)
		i--
		if i < 0 {
			return nil, fmt.Errorf("Base %s not found in history of %s", base_id, img.ID)
		}
	}

	Debug("Skipping", history[i].ID)
	i--

	lrdir, err := ioutil.TempDir(gr.Workdir, "layer.")
	if err != nil {
		return nil, err
	}

	lr := NewLayer(lrdir)

	for i > 0 {
		if err := lr.Apply(gr.ImageRoot(history[i].ID)); err != nil {
			return nil, err
		}
		i--
	}

	lrf, err := ioutil.TempFile(gr.Workdir, "layer.tar.")
	if err != nil {
		return nil, err
	}

	if err := lr.Save(lrf); err != nil {
		return nil, err
	}
	lrf.Close()

	oimg := img.Clone()
	oimg.Parent = base_id
	oimg.Created = time.Now()
	oimg.Size = lr.Size()
	oimg.Tarball = lrf.Name()
	return oimg, nil
}
