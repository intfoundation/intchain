BUILD_FLAGS = -tags "$(BUILD_TAGS)" -ldflags "

build:
	@ echo "start building......"
	@ go build -o $(GOPATH)/bin/intchain ./cmd/intchain/
	@ echo "Done building."
#.PHONY: intchain
intchain:
	@ echo "start building......"
	@ go build -o $(GOPATH)/bin/intchain ./cmd/intchain/
	@ echo "Done building."
	@ echo "Run intchain to launch intchain network."

install:
	@ echo "start install......"
	@ go install -mod=readonly $(BUILD_FLAGS) ./cmd/intchain
	@ echo "install success......"