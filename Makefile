GOPATH := $(shell go env GOPATH)
GODEP   = $(GOPATH)/bin/godep
TARGET  = bin/docker-rebase

all: $(TARGET)

$(TARGET): $(GODEP) *.go
	$(GODEP) go build -o $(TARGET)

test: $(TARGET) bundle fixtures
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

fixtures: features/fixtures/smoke.tar.gz

features/fixtures/%.tar.gz: features/fixtures/%/Dockerfile
	docker build -t fixture/$* $(dir $<)
	docker save fixture/$* | gzip > $@
	docker rmi fixture/$*

clean:
	rm -rf tmp
