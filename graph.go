package main

import "archive/tar"
import "errors"
import "fmt"
import "io"
import "io/ioutil"
import "log"
import "os"
import "path"
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

func (gr *Graph) ReadFrom(r io.Reader) (int64, error) {
	c := NewCounter(r)
	tr := tar.NewReader(c)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.Bytes, err
		}

		switch path.Base(hdr.Name) {
		case "json":
			json_bb, err := ioutil.ReadAll(tr)
			if err != nil {
				return c.Bytes, err
			}
			img, err := NewImageJSON(json_bb)
			if err != nil {
				return c.Bytes, err
			}
			gr.AddImage(img)
		case "layer.tar":
			id := path.Base(path.Dir(hdr.Name))
			tar_path := gr.ImageRoot(id)
			if tarf, err := os.Create(tar_path); err != nil {
				return c.Bytes, err
			} else {
				_, err := io.Copy(tarf, tr)
				tarf.Close()
				if err != nil {
					return c.Bytes, err
				}
			}
		}
	}

	return c.Bytes, nil
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

func (gr *Graph) Rebase(id, base_id string) (*Image, error) {
	img, _ := gr.Images[id]

	if img == nil {
		return nil, fmt.Errorf("%s: Image not found", id)
	}

	history, err := img.History()
	if err != nil {
		return nil, err
	}

	i := len(history) - 1
	for history[i].ID != base_id {
		Debug("Skipping", history[i].ID)
		i--
		if i < 0 {
			return nil, fmt.Errorf("Base %s not found in history of %s", base_id, id)
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
		log.Println(history[i].ID, history[i].Comment, history[i].ContainerConfig.Cmd)
		if err := lr.Apply(gr.ImageRoot(history[i].ID)); err != nil {
			return nil, err
		}
		i--
	}

	lrf, err := ioutil.TempFile(gr.Workdir, "layer.tar.")
	if err != nil {
		return nil, err
	}

	if _, err := lr.WriteTo(lrf); err != nil {
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
