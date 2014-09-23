GOPATH := $(shell go env GOPATH)
GODEP   = $(GOPATH)/bin/godep
TARGET  = bin/docker-rebase

all: $(TARGET)

$(TARGET): $(GODEP) *.go
	$(GODEP) go build -o $(TARGET)

test: $(TARGET) bundle
	bundle exec cucumber

@%: $(TARGET) bundle
	bundle exec cucumber --tags @$*

bundle: tmp/bundle.stamp

$(GODEP):
	go get github.com/tools/godep
	go install github.com/tools/godep


tmp/bundle.stamp: Gemfile Gemfile.lock
	mkdir -p tmp
	bundle install
	date > $@

clean:
	rm -rf tmp
