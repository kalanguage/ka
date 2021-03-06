ifeq ($(OS),Windows_NT)
	BINARY = tuskstart.exe
	CLEAN_CMD = del
else
	SET_GOPATH = GOPATH=$(GOPATH)
	BINARY = tuskstart.out
	CLEAN_CMD = rm -f
endif

GOPATH = $(CURDIR)/../../../../

.PHONY: default
default: all

.PHONY: all
all: language

.PHONY: clean
clean:
	-$(CLEAN_CMD) $(BINARY)

.PHONY: language
language:
	$(SET_GOPATH) go build -a -o $(BINARY) tuskstart.go
