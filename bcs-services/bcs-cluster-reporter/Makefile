SRV_NAME = bcs-cluster-reporter
VER = v0.0.8
CURRENT_VERSION = $(VER)
NAMESPACE = build/blueking
DH_URL = mirrors.tencent.com

LDFLAG=-ldflags "-X bcs-services/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
-X github.com/Tencent/bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key}"

clean:
	-rm ./$(SRV_NAME)

build:clean
	CGO_ENABLED=1 go build -o $(SRV_NAME) ${LDFLAG} -v ./main.go

publish:build
	docker build --build-arg SRV_NAME=$(SRV_NAME) --rm -t $(SRV_NAME):$(CURRENT_VERSION) .
	docker tag $(SRV_NAME):$(CURRENT_VERSION) $(DH_URL)/${NAMESPACE}/$(SRV_NAME):$(CURRENT_VERSION)
	docker push $(DH_URL)/${NAMESPACE}/$(SRV_NAME):$(CURRENT_VERSION)