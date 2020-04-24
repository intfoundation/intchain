BUILD_FLAGS = -tags "$(BUILD_TAGS)" -ldflags "

build:
	go build -o $(GOPATH)/bin/intchain ./cmd/intchain

#.PHONY: intchain
intchain:
	build/env.sh
	@echo "Done building."
	@echo "Run ./bin/intchain to launch intchain network."

install:
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/intchain
