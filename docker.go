package main

import "bufio"
import "flag"
import "fmt"
import "io"
import "io/ioutil"
import "log"
import "os"
import "regexp"
import "strings"

import "github.com/docker/docker/api"
import "github.com/docker/docker/api/client"

// Constants
var base_rx = regexp.MustCompile(`^Step \d+ : FROM (.*)`)
var image_rx = regexp.MustCompile(`^Successfully built ([0-9a-fA-F]+)$`)

// Connection details

func splitDockerHost(host string) (proto, addr string, err error) {
	host, err = api.ValidateHost(host)
	if err != nil {
		return "", "", err
	}
	pieces := strings.SplitN(host, "://", 2)
	return pieces[0], pieces[1], nil
}

type DockerHost struct {
	Proto, Addr string
}

func (dh DockerHost) String() string {
	return dh.Proto + "://" + dh.Addr
}

func (dh DockerHost) Set(val string) error {
	if proto, addr, err := splitDockerHost(val); err != nil {
		return err
	} else {
		dh.Proto, dh.Addr = proto, addr
	}
	return nil
}

func defaultDockerHost() DockerHost {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		if proto, addr, err := splitDockerHost(host); err == nil {
			return DockerHost{proto, addr}
		}
	}
	return DockerHost{"unix", api.DEFAULTUNIXSOCKET}
}

func (dh DockerHost) Cli(in io.ReadCloser, out io.Writer) *client.DockerCli {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return client.NewDockerCli(in, out, os.Stderr, dh.Proto, dh.Addr, nil)
}

func (dh DockerHost) ReadPipe(cmd ...string) (rc io.ReadCloser) {
	stdout_r, stdout_w := io.Pipe()
	cli := dh.Cli(nil, stdout_w)
	go func() {
		defer func() { stdout_w.Close() }()
		if err := cli.Cmd(cmd...); err != nil {
			log.Fatalln("Docker CLI", cmd, ":", err)
		}
	}()
	return stdout_r
}

var Docker = defaultDockerHost()

func init() {
	flag.Var(Docker, "docker-host", "Docker host address")
}

func buildImage(args []string) (base_id, image_id string) {
	args = append([]string{"build"}, args...)
	scanner := bufio.NewScanner(Docker.ReadPipe(args...))
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if match := base_rx.FindStringSubmatch(line); match != nil {
			base_id = match[1]
		} else if match := image_rx.FindStringSubmatch(line); match != nil {
			image_id = match[1]
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("Error scanning build output:", err)
	}
	return
}

func getId(name string) string {
	id_bytes, err := ioutil.ReadAll(Docker.ReadPipe("inspect", "-f", "{{.Id}}", name))
	if err != nil {
		log.Fatalln("GetID:", err)
	}
	return strings.TrimSpace(string(id_bytes))
}
