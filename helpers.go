package main

import "bytes"
import "encoding/json"
import "fmt"
import "io"
import "log"
import "os"

func jsonPP(o interface{}) {
	var bb []byte
	var err error

	switch o.(type) {
	case []byte:
		bb = o.([]byte)
	default:
		bb, err = json.Marshal(o)
		if err != nil {
			log.Fatalln("JSON", o, ":", err)
		}
	}

	buf := &bytes.Buffer{}
	if err = json.Indent(buf, bb, "", "  "); err != nil {
		log.Fatalln("Indent", o, ":", err)
	} else {
		io.Copy(os.Stdout, buf)
		fmt.Println()
	}
}

type Counter struct {
	io    interface{}
	Bytes int64
}

func NewCounter(io interface{}) *Counter {
	return &Counter{io, 0}
}

func (c *Counter) Write(p []byte) (n int, err error) {
	n, err = c.io.(io.Writer).Write(p)
	c.Bytes += int64(n)
	return
}

func (c *Counter) Read(p []byte) (n int, err error) {
	n, err = c.io.(io.Reader).Read(p)
	c.Bytes += int64(n)
	return
}
