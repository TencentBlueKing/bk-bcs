TARGET = bk-bscp-e2e-testing

GOENVPATH = $(shell go env GOPATH)

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET)...\033[0m"
	go test -c ./... -o $(TARGET)
	@echo -e "\e[34;1mBuild $(TARGET) success!\n\033[0m"

clean:
	@rm -rf $(TARGET)
