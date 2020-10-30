rm bcs-gamedeployment-operator
go build ../../cmd/gamedeployment-operator/
mv gamedeployment-operator  bcs-gamedeployment-operator


BUILD_ARG="--build-arg TIMEZONE=$IMAGE_TIMEZONE"

docker rmi "${IMAGE_BASE_REPO}gamedeployment:${IMAGE_VERSION}"

docker build -t "${IMAGE_BASE_REPO}gamedeployment:${IMAGE_VERSION}" $BUILD_ARG -f Dockerfile .

docker push "${IMAGE_BASE_REPO}gamedeployment:${IMAGE_VERSION}"
