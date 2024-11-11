PKG_NAME = bcs-ops
VER = v0.0.3
CURRENT_VERSION = release-$(VER)

clean:
	-rm ./bcs-ops-script-release-$(VER)*.tgz
	-rm ./bcs-ops-offline-release-$(VER)-k8s-*.tgz
	-rm -r ./bcs-ops-offline
	-rm MD5SUMS

build:clean
	find . -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*" -a -not -path "*/functions/*" -a -not -path "*.yaml" -a -not -path "*.md" -type f -print0 | xargs -0 chmod 555
	find ./functions/ -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*"  -type f -print0 | xargs -0 chmod 444
	tar -czvf bcs-ops-script-release-$(VER).tgz --exclude=bcs-ops-offline --exclude=\..* --exclude=.*tar.gz --exclude=.*tgz  --exclude=version-.* --exclude=.*xz ./*
	md5sum bcs-ops-script-release-$(VER).tgz >> MD5SUMS

build_offline_1:build
	./offline_package.sh env/offline-manifest.yaml
	mv ./bcs-ops-offline/*.tgz ./

build_offline_pkg:build_offline_1
	for f in $(shell ls bcs-ops-offline*.tgz);do \
		mv $$f bcs-ops-offline-release-$(VER)-k8s-$${f##*-offline-}; \
	done
	md5sum bcs-ops-offline-release-$(VER)-k8s-*.tgz >> MD5SUMS

build_image:clean
	docker build -f docker/Dockerfile -t bcs-ops-upload:test .