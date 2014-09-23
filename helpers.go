package main

import "bytes"
import "compress/gzip"
import "encoding/json"
import "flag"
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

func ensure(body, finalizer func()) {
	defer finalizer()
	body()
}

// Generic, callback-based flag
type callbackValue struct {
	val    string
	isBool bool
	setter func(string) error
}

func (cv callbackValue) String() string {
	return cv.val
}

func (cv callbackValue) IsBoolFlag() bool {
	return cv.isBool
}

func (cv callbackValue) Set(val string) error {
	cv.val = val
	return cv.setter(val)
}

func Flag(name, def, desc string, setter func(string) error) {
	cv := callbackValue{def, false, setter}
	if def != "" {
		cv.Set(def)
	}
	flag.Var(cv, name, desc)
}

func BoolFlag(name, desc string, setter func() error) {
	cv := callbackValue{"false", true, func(string) error { return setter() }}
	flag.Var(cv, name, desc)
}

func CloseAfterReading(v io.ReadCloser, handle func(io.Reader) error) error {
	defer v.Close()
	return handle(v)
}

func WithOpen(path string, handle func(io.Reader) error) error {
	if path == "-" {
		return CloseAfterReading(os.Stdin, handle)
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return CloseAfterReading(f, handle)
}

func WithOpenZ(path string, handle func(io.Reader) error) error {
	return WithOpen(path, func(zr io.Reader) error {
		r, err := gzip.NewReader(zr)
		if err != nil {
			return err
		}
		return CloseAfterReading(r, handle)
	})
}

func CloseAfterWriting(w io.WriteCloser, handle func(io.Writer) error) error {
	defer w.Close()
	return handle(w)
}

func WithCreate(path string, handle func(io.Writer) error) error {
	if path == "-" {
		return CloseAfterWriting(os.Stdout, handle)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	return CloseAfterWriting(f, handle)
}

func WithCreateZ(path string, handle func(io.Writer) error) error {
	return WithCreate(path, func(w io.Writer) error {
		return CloseAfterWriting(gzip.NewWriter(w), handle)
	})
}
