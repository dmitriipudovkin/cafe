PKG ?= auth

PROTO_FILES = proto/$(PKG)/$(PKG).proto

GEN_DIR = gen/go

PROTOC = protoc
PROTOC_FLAGS = -I proto

GO_OUT = --go_out=$(GEN_DIR) --go_opt=paths=source_relative
GO_GRPC_OUT = --go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative

all: gen

gen:
	$(PROTOC) $(PROTOC_FLAGS) $(GO_OUT) $(GO_GRPC_OUT) $(PROTO_FILES)

clean:
	rm -rf $(GEN_DIR)

.PHONY: all gen clean