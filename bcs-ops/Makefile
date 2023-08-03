PKG_NAME = bcs-ops
VER = v0.0.1
CURRENT_VERSION = release-$(VER)

clean:
	-rm ./bcs-ops-script-release-$(VER).tar.gz
	-rm ./bcs-ops-offline-release-$(VER).tar.gz
	-rm MD5SUMS

build:clean
	find . -not -path "*/.git/*" -a -not -path "*/bin/*" -a -not -path "*/image/*" -a -not -path "*/Makefile" -type f -print0
	tar -czvf bcs-ops-script-release-$(VER).tar.gz  --exclude=bin --exclude=image --exclude=Makefile --exclude=\..* --exclude=.*tar.gz ./*
	tar -czvf bcs-ops-offline-release-$(VER).tar.gz  --exclude=Makefile --exclude=\..* --exclude=.*tar.gz ./*
	md5sum bcs-ops-script-release-$(VER).tar.gz >> MD5SUMS
	md5sum bcs-ops-offline-release-$(VER).tar.gz >> MD5SUMS