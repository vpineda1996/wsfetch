
BUILD_DIR := $(abspath ./build)
TOOL_DIR := $(BUILD_DIR)/tools
SCRIPTS_BIN_DIR := $(abspath ./hack/scripts/bin)

MOCKGEN := $(TOOL_DIR)/mockgen
GENQLIENT := $(TOOL_DIR)/genqlient

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(TOOL_DIR): $(BUILD_DIR)
	mkdir -p $(TOOL_DIR)

$(MOCKGEN): $(TOOL_DIR)
	GOBIN=$(TOOL_DIR) go install github.com/golang/mock/mockgen

$(GENQLIENT): $(TOOL_DIR)
	GOBIN=$(TOOL_DIR) go install github.com/Khan/genqlient

generate: $(MOCKGEN) $(GENQLIENT)
	PATH="$(PATH):$(TOOL_DIR):$(SCRIPTS_BIN_DIR)" go generate ./...




clean:
	rm -rf $(BUILD_DIR)