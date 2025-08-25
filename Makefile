APP_NAME=golang_course
BIN_NAME=gobin
BUILD_DIR=./bin
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

run: 
	@echo " Runnning the server"
	@go run main.go || true 

deps:
	@echo "Installing dependencies"
	@go mod tidy 


fmt:
	@echo "Formatting code......"
	@go fmt ./...

build:
	@echo "Builind the project"
	@mkdir -p ${BUILD_DIR}
	@go build -o $(BUILD_DIR)/$(BIN_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)"


clean:
	@echo "Cleaning up"
	@rm -rf ${BUILD_DIR}
	@echo "CleanUp complete"

stop:
	@echo "Stopping server....."
	@pkill -f "go run main.go" || echo "no process found"

migrate-up:
	@echo "Runnning database migration"
	@dbmate up
	@echo "migration applied"

migrate-down:
	@echo "Rolling back last migration"
	@dbmate down
	@echo "migration rolled down"

help:
	@echo "Available commands"
	@echo " make run     - Run the server"
	@echo " make deps     - Install dependencies"
	@echo " make fmt      - Formats the code"
	@echo " make build    - Build the binary"
	@echo "make clean     - Clean up the binary"
	@echo " make stop     - stop all Runnning server"
