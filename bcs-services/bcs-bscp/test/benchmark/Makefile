TARGET = bench-test

PWD        = $(shell pwd)
OUTPUT_DIR = $(PWD)/$(TARGET)

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET)...\033[0m"
	@mkdir -p $(OUTPUT_DIR)/tools
	@cd tools/gen-data && go build -o $(OUTPUT_DIR)/tools/gen-data
	@cd case/cache-service && go test -c -o $(OUTPUT_DIR)/cache_service.test
	@cd case/feed-server && go test -c fs_http_test.go -o $(OUTPUT_DIR)/feed_server_http.test
	@cp tools/start.sh $(OUTPUT_DIR)/
	@cp tools/README.md $(OUTPUT_DIR)/
	@echo -e "\e[34;1mBuild $(TARGET) success!\n\033[0m"

clean:
	@rm -rf $(OUTPUT_DIR)
