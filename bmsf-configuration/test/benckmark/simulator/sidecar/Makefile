TARGET = bscp-simulator-sidecar

GOENVPATH = $(shell go env GOPATH)

default:
	@echo -e "\e[34;1mBuilding $(TARGET)...\033[0m"
	go build -gcflags "-N -l" -o $(TARGET) *.go
	@echo -e "\e[34;1mBuild $(TARGET) success!\n\033[0m"

clean:
	@rm -rf $(TARGET)
