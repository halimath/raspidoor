VERSION := 0.1.0
BUILD_TIMESTAMP := $(shell date --utc --iso-8601=seconds)
REVISION := $(shell git rev-parse HEAD)

GO := go
GOOS := linux 
# GOARCH := arm64
GOARCH := arm
GOARM := 5

TARGET_DEV_HOST := raspberrypi
TARGET_DEV_USER := pi
TARGET_DEV_DIR := ~/raspidoor

TARGET_PROD_HOST := klingel
TARGET_PROD_USER := root
TARGET_PROD_DIR := /opt/raspidoor
TARGET_PROD_CONFIG_DIR := /etc/raspidoor

SSH := ssh
SSH_OPTS := 

SCP := scp
SCP_OPTS := 

RSYNC := rsync
RSYNC_OPTS := -avz

PROTOC := protoc

RM := rm
RM_OPTS := -rf

.PHONY: raspidoord.$(GOARCH) clean install install-dev install-prod

raspidoord.$(GOARCH):
	env GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -ldflags '-X main.Version=$(VERSION) -X main.Revision=$(REVISION) -X main.BuildTimestamp=$(BUILD_TIMESTAMP)' -o $@ cmd/raspidoord/main.go

controller/controller.pb.go: controller/controller.proto
	$(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $<

install: install-dev

install-dev: raspidoord.$(GOARCH)
	$(RSYNC) $(RSYNC_OPTS) raspidoord.$(GOARCH) raspidoord.yaml $(TARGET_DEV_USER)@$(TARGET_DEV_HOST):$(TARGET_DEV_DIR)

install-prod: install-prod-bin install-prod-systemd install-prod-config

install-prod-bin: raspidoord.$(GOARCH)
	$(SSH) $(SSH_OPTS) $(TARGET_PROD_USER)@$(TARGET_PROD_HOST) 'mkdir -p $(TARGET_PROD_DIR)'
	$(SCP) $(SCP_OPTS) raspidoord.$(GOARCH) $(TARGET_PROD_USER)@$(TARGET_PROD_HOST):$(TARGET_PROD_DIR)/raspidoord

install-prod-systemd:
	$(SCP) $(SCP_OPTS) systemd/raspidoor.* $(TARGET_PROD_USER)@$(TARGET_PROD_HOST):/etc/systemd/system

install-prod-config:
	$(SSH) $(SSH_OPTS) $(TARGET_PROD_USER)@$(TARGET_PROD_HOST) 'mkdir -p $(TARGET_PROD_CONFIG_DIR)'
	$(SCP) $(SCP_OPTS) raspidoord.yaml $(TARGET_PROD_USER)@$(TARGET_PROD_HOST):$(TARGET_PROD_CONFIG_DIR)/raspidoord.yaml

clean:
	$(RM) $(RM_OPTS) raspidoor.$(GOARCH)