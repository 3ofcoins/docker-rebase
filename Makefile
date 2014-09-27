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

fixtures: $(addsuffix .tar.gz,$(patsubst %/,%,$(dir $(wildcard features/fixtures/*/Dockerfile))))

ENVDIR = features/fixtures/_env
features/fixtures/%.tar.gz: env = \
    FIXTURE_$(shell echo $(notdir $(patsubst %/,%,$(dir $<))) | tr a-z A-Z)
features/fixtures/%.tar.gz: features/fixtures/%/Dockerfile
	docker build -t fixture/$* $(dir $<)
	docker save fixture/$* | gzip > $@
	mkdir -p $(ENVDIR)
	docker inspect --format='{{.Id}}' fixture/$* > $(ENVDIR)/$(env)_ID
	docker inspect --format='{{.Id}}' `awk '/^FROM / { print $$2 }' $<` > $(ENVDIR)/$(env)_BASE_ID
	head -c 12 < $(ENVDIR)/$(env)_ID > $(ENVDIR)/$(env)_SHORT_ID
	head -c 12 < $(ENVDIR)/$(env)_BASE_ID > $(ENVDIR)/$(env)_BASE_SHORT_ID
	docker rmi fixture/$*

clean:
	rm -rf tmp

mrproper: clean
	rm -rf features/fixtures/*.tar.gz $(ENVDIR)
