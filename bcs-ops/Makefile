PKG_NAME = bcs-ops
VER = v0.0.3
CURRENT_VERSION = release-$(VER)

clean:
	-rm ./bcs-ops-script-release-$(VER)*.tar.gz
	-rm ./bcs-ops-offline-release-$(VER)-k8s-*.tgz
	-rm -r ./bcs-ops-offline
	-rm MD5SUMS

build:clean
	find . -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*" -a -not -path "*/functions/*" -a -not -path "*.yaml" -a -not -path "*.md" -type f -print0 | xargs -0 chmod 555
	find ./functions/ -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*"  -type f -print0 | xargs -0 chmod 444
	tar -czvf bcs-ops-script-release-$(VER).tar.gz --exclude=bcs-ops-offline --exclude=\..* --exclude=.*tar.gz ./*
	md5sum bcs-ops-script-release-$(VER).tar.gz >> MD5SUMS

build_offline_pkg:clean
	find . -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*" -a -not -path "*/functions/*" -a -not -path "*.yaml" -a -not -path "*.md" -type f -print0 | xargs -0 chmod 555
	find ./functions/ -not -path "*/.git/*" -a -not -path "*/bcs-ops-offline/*" -a -not -path "*/charts/*" -type f -print0 | xargs -0 chmod 444
	tar -czvf bcs-ops-script-release-$(VER).tar.gz --exclude=bcs-ops-offline --exclude=\..* --exclude=.*tar.gz ./*
	md5sum bcs-ops-script-release-$(VER).tar.gz >> MD5SUMS
	./offline_package.sh env/offline-manifest.yaml
	for f in $(find bcs-ops-offline -iname bcs-ops-offline-*.tgz -type f);do
		mv $f bcs-ops-offline-release-$(VER)-k8s-${f##*-}
	done
	md5sum bcs-ops-offline-release-$(VER)-k8s-*.tgz >> MD5SUMS


build_image:clean
	docker build -f docker/Dockerfile -t bcs-ops-upload:test .
