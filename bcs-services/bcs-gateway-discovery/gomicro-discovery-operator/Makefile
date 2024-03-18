BUILD_PATH ?= $(shell pwd)/build
.PHONY: build
build:
	@if [ ! -e "$(BUILD_PATH)" ]; then mkdir $(BUILD_PATH); fi
	go build -o $(BUILD_PATH)/gomicro-discovery-operator .

container: build
	docker build . -t gomicro-discovery-operator:latest

run: container
	kubectl apply -f deploy.yaml

update: container
	kubectl scale deploy gomicro-discovery-operator --replicas=0
	kubectl scale deploy gomicro-discovery-operator --replicas=1

stop:
	kubectl delete -f deploy.yaml
