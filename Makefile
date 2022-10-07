SHELL=/bin/bash -o pipefail
GO ?= go

NAME := nomad
OUTPUT := lib$(NAME).so

ifeq ($(DEBUG), 1)
    GODEBUGFLAGS= GODEBUG=cgocheck=2
else
    GODEBUGFLAGS= GODEBUG=cgocheck=0
endif

build: 
	@$(GODEBUGFLAGS) $(GO) build -buildmode=c-shared -o bin/$(OUTPUT) ./plugin

install:
	sudo cp bin/$(OUTPUT) /usr/share/falco/plugins/$(OUTPUT)

clean:
	@rm -rf bin

run:
	falco -c $(PWD)/example/falco.yaml --disable-source=syscall -r $(PWD)/rules/nomad_rules.yaml

