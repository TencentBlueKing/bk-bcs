export controllergen="$GOPATH/bin/controller-gen"

echo "Generating CRDs..."
$controllergen \
  crd \
  schemapatch:manifests=./manifests/crds \
  paths=./pkg/apis/... \
  output:dir=./manifests/crds
echo "Generate CRDs in ./manifests/crds done."