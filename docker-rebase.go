package main

import "flag"
import "fmt"
import "io/ioutil"
import "log"
import "os"

func usage() {
	fmt.Fprintln(os.Stderr,
		"Usage:", os.Args[0], "[options] IMAGE BASE\n",
		"      ", os.Args[0], "[options] -build -- [docker build args ...]\nRecognised options:")
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
	flag.BoolVar(&DEBUG, "debug", false, "Print debug info, leave work directory around")
}

// Global flags

func main() {
	var build bool
	var base_id, image_id string
	var base_image, target_image string

	flag.Usage = usage
	flag.BoolVar(&build, "build", false, "Build image and then rebase it")
	flag.Parse()

	if build {
		base_image, target_image = buildImage(flag.Args())
	} else {
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(1)
		}
		base_image = flag.Arg(1)
		target_image = flag.Arg(0)
	}

	base_id = getId(base_image)
	image_id = getId(target_image)

	log.Printf("Rebasing %v (%v) onto %v (%v)\n", target_image, image_id, base_image, base_id)

	workdir, err := ioutil.TempDir("", "docker-rebase.")
	if err != nil {
		log.Fatalln("TempDir:", err)
	}

	defer func() {
		if DEBUG {
			Debug("keeping workdir", workdir)
		} else if err := os.RemoveAll(workdir); err != nil {
			log.Printf("ERROR: cannot remove %v: %v\n", workdir, err)
		}
	}()

	gr := NewGraph(workdir)

	if _, err := gr.ReadFrom(Docker.ReadPipe("save", image_id)); err != nil {
		log.Fatalln("Load", image_id, ":", err)
	}

	img, err := gr.Rebase(image_id, base_id)
	if err != nil {
		log.Fatalln(err)
	}

	of, err := os.Create("rebased.tar")
	if err != nil {
		log.Fatalln(err)
	}
	defer of.Close()

	n, err := img.WriteTo(of)
	log.Println("Wrote", n, "bytes:", err)
}
