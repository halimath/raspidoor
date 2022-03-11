VERSION = 0.1.0
BUILD_TIMESTAMP := $(shell date --utc --iso-8601=seconds)
REVISION := $(shell git rev-parse HEAD)

DEB_ARCH ?= armhf
DEB_REVISION = 1

GO ?= go
GOOS ?= linux 
# GOARCH := arm64
GOARCH ?= arm
GOARM ?= 5

CLI_SOURCES = $(wildcard cli/*/*.go)
DAEMON_SOURCES = $(wildcard daemon/*/*.go)

 DEV_HOST ?= raspberrypi
DEV_USER ?= pi
DEV_DIR ?= ~/raspidoor

MKDIR ?= mkdir
MKDIR_OPTS ?= -p

CP ?= cp
CP_OPTS ?= -r

SSH ?= ssh
SSH_OPTS ?= 

SCP ?= scp
SCP_OPTS ?= 

RSYNC ?= rsync
RSYNC_OPTS ?= -avz

PROTOC ?= protoc

RM ?= rm
RM_OPTS ?= -rf

M4 ?= m4

.PHONY: clean install docker-build-deb

raspidoord.$(GOARCH): $(DAEMON_SOURCES)
	cd daemon && env GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -ldflags '-X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BuildTimestamp=$(BUILD_TIMESTAMP)' -o ../$@ .

raspidoor.$(GOARCH): $(CLI_SOURCES)
	cd cli && env GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -ldflags '-X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BuildTimestamp=$(BUILD_TIMESTAMP)' -o ../$@ .

controller/controller.pb.go: controller/controller.proto
	$(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

install: raspidoord.$(GOARCH)
	$(RSYNC) $(RSYNC_OPTS) raspidoord.$(GOARCH) raspidoord.yaml $(DEV_USER)@$(DEV_HOST):$(DEV_DIR)

raspidoor_$(VERSION)-$(DEB_REVISION)_$(DEB_ARCH).deb: raspidoor_$(VERSION)-$(DEB_REVISION)_$(DEB_ARCH)
	dpkg-deb --build --root-owner-group $<

raspidoor_$(VERSION)-$(DEB_REVISION)_$(DEB_ARCH): raspidoor.$(GOARCH) raspidoord.$(GOARCH)
	$(MKDIR) $(MKDIR_OPTS) $@/usr/bin
	$(CP) $(CP_OPTS) raspidoord.$(GOARCH) $@/usr/bin/raspidoord
	$(CP) $(CP_OPTS) raspidoor.$(GOARCH) $@/usr/bin/raspidoor

	$(MKDIR) $(MKDIR_OPTS) $@/etc/raspidoor
	$(CP) $(CP_OPTS) daemon/etc/raspidoord.yaml $@/etc/raspidoor

	$(MKDIR) $(MKDIR_OPTS) $@/etc/systemd/system
	$(CP) $(CP_OPTS) daemon/etc/systemd/raspidoor.service $@/etc/systemd/system

	$(MKDIR) $(MKDIR_OPTS) $@/DEBIAN
	$(M4) -DVERSION=$(VERSION) -DARCH=$(DEB_ARCH) DEBIAN/control > $@/DEBIAN/control

docker-build-deb:
	docker build --build-arg version=$(VERSION) --build-arg debrevision=$(DEB_REVISION) --build-arg goarch=$(GOARCH) --build-arg debarch=$(DEB_ARCH) -t raspidoor-builder:$(VERSION) .
	$(MKDIR) $(MKDIR_OPTS) out
	docker run --rm -it -v $(shell pwd)/out:/out raspidoor-builder:$(VERSION)

clean:
	$(RM) $(RM_OPTS) raspidoor.$(GOARCH)