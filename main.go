package main

import "flag"
import "fmt"
import "io/ioutil"
import "log"
import "os"

func usage() {
	fmt.Fprintln(os.Stderr,
		"Usage:\t",
		os.Args[0], "[options] IMAGE BASE\n\t",
		os.Args[0], "[options] -build -- [docker build args ...]\nRecognised options:")
	flag.PrintDefaults()
}

var DEBUG bool

func Debug(v ...interface{}) {
	if DEBUG {
		v = append([]interface{}{"DEBUG:"}, v...)
		log.Println(v...)
	}
}

func init() {
	flag.BoolVar(&DEBUG, "debug", false, "Show debug output")
}

func isNonempty(v interface{}) bool {
	switch v := v.(type) {
	case bool:
		return v
	case string:
		return v != ""
	default:
		return v != nil
	}
	return false
}

func countNonempty(vv ...interface{}) int {
	n := 0
	for _, v := range vv {
		if isNonempty(v) {
			n++
		}
	}
	return n
}

func usageln(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	usage()
	os.Exit(1)
}

// Global state
var gr *Graph
var load, zload, save, zsave, base_id, image_id string

func loadGraph() error {
	switch {
	case load != "":
		Debug("Loading graph from", load)
		return WithOpen(load, gr.Load)
	case zload != "":
		Debug("Loading gzipped graph from", zload)
		return WithOpenZ(zload, gr.Load)
	default:
		Debug("Loading graph from `docker save' of", image_id)
		return CloseAfterReading(Docker.ReadPipe("save", image_id), gr.Load)
	}
}

func saveImage(img *Image) error {
	switch {
	case save != "":
		Debug("Saving image to", save)
		return WithCreate(save, img.Save)
	case zsave != "":
		Debug("Saving gzipped image to", zsave)
		return WithCreateZ(zsave, img.Save)
	default:
		Debug("Loading image into Docker")
		return fmt.Errorf("Loading image into Docker not implemented yet")
	}
}

func main() {
	var build bool
	var workdir string

	flag.Usage = usage
	flag.BoolVar(&build, "build", false, "Build image and then rebase it")
	flag.StringVar(&load, "load", "", "Load image graph from a tar archive (\"-\" for stdin)")
	flag.StringVar(&zload, "zload", "", "Load image graph from a gzipped tar archive (\"-\" for stdin)")
	flag.StringVar(&save, "save", "", "Save rebased image to a tar archive (\"-\" for stdout)")
	flag.StringVar(&zsave, "zsave", "", "Save rebased image to a gzipped tar archive (\"-\" for stdout)")
	flag.StringVar(&workdir, "workdir", "", "Use given working directory and keep it when finished")

	BoolFlag("help", "show usage info", func() error {
		usage()
		os.Exit(0)
		return nil
	})

	flag.Parse()

	// Err on disabled flag combinations
	if countNonempty(build, load, zload) > 1 {
		usageln("You cannot specify more than one of: -build, -load, -zload")
	}

	if countNonempty(save, zsave) > 1 {
		usageln("You cannot specify more than one of: -save -zsave")
	}

	if !build && flag.NArg() != 2 {
		usageln("You need to specify base and target image")
	}

	// Actual rebase workflow

	// 1. Build image if requested

	if build {
		base_id, image_id = buildImage(flag.Args())
	} else {
		base_id, image_id = flag.Arg(1), flag.Arg(0)
	}
	if load == "" && zload == "" {
		// We load graph from Docker, we can get full image IDs as well
		base_id, image_id = getId(base_id), getId(image_id)
	}

	// 2. Create graph with workspace in a temporary directory

	if workdir == "" {
		workdir, err := ioutil.TempDir("", "docker-rebase.")
		if err != nil {
			log.Fatalln("TempDir:", err)
		}

		defer func() {
			if err := os.RemoveAll(workdir); err != nil {
				log.Println("ERROR: cannot remove", workdir, ":", err)
			}
		}()
	}

	gr = NewGraph(workdir)

	// 3. Load images into graph

	if err := loadGraph(); err != nil {
		log.Fatalln("Error loading graph:", err)
	}

	// 5. Rebase

	log.Printf("Rebasing %v onto %v\n", image_id, base_id)

	img, err := gr.Rebase(image_id, base_id)
	if err != nil {
		log.Fatalln(err)
	}

	// 6. Save rebased image

	if err := saveImage(img); err != nil {
		log.Fatalln(err)
	}
}
