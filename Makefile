GO := go
PKGER := pkger
SRC_DIR := src
SRC_PKGER := src/pkged.go
ASSETS_DIR := /assets
DEST := misaki
PKGER_URL := github.com/markbates/pkger/cmd/pkger

# this is a workaround for https://github.com/markbates/pkger/issues/49
DUMMY_GO_FILE := dummy.go

.PHONY: build pack install_pkger clean

build: pack
	$(GO) build -o $(DEST) $(SRC_DIR)/*
	$(RM) $(SRC_PKGER)
pack:
	echo "package dummy" > $(DUMMY_GO_FILE)
	$(PKGER) -include $(ASSETS_DIR) -o $(SRC_DIR)
	$(RM) $(DUMMY_GO_FILE)
install_pkger:
	$(GO) get $(PKGER_URL)
clean:
	$(RM) $(DEST) $(SRC_PKGER)
