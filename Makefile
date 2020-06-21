GO := go
PKGER := pkger
BUILD_DIR := ./build
SRC_PKGER := ./pkg/pkged.go
ASSETS_PATH := /pkg/assets
PKGER_URL := github.com/markbates/pkger/cmd/pkger

# this is a workaround for https://github.com/markbates/pkger/issues/49
DUMMY_GO_FILE := dummy.go

.PHONY: build pack install_pkger clean

build: pack
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR) ./...
	$(RM) $(SRC_PKGER)
pack:
	echo "package dummy" > $(DUMMY_GO_FILE)
	$(PKGER) -include $(ASSETS_PATH) -o $(dir $(SRC_PKGER))
	$(RM) $(DUMMY_GO_FILE)
install_pkger:
	$(GO) get $(PKGER_URL)
clean:
	$(RM) $(BUILD_DIR)/* $(SRC_PKGER)
