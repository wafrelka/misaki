GO := go
SRC_DIR := src
DEST := misaki

.PHONY: build clean

build:
	$(GO) build -o $(DEST) $(SRC_DIR)/*
clean:
	$(RM) $(DEST)
