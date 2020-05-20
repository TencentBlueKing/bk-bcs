all:
	@cd bscp-client && make
	@cd bscp-gateway && make
	@cd bscp-accessserver && make
	@cd bscp-businessserver && make
	@cd bscp-templateserver && make
	@cd bscp-datamanager && make
	@cd bscp-bcs-controller && make
	@cd bscp-connserver && make
	@cd bscp-bcs-sidecar && make
	@cd bscp-integrator && make

clean:
	@cd bscp-client && make clean
	@cd bscp-gateway && make clean
	@cd bscp-accessserver && make clean
	@cd bscp-businessserver && make clean
	@cd bscp-templateserver && make clean
	@cd bscp-datamanager && make clean
	@cd bscp-bcs-controller && make clean
	@cd bscp-connserver && make clean
	@cd bscp-bcs-sidecar && make clean
	@cd bscp-integrator && make clean

images:
	@cd bscp-bcs-sidecar && make image

bscpimages:
	@cd bscp-gateway && make image
	@cd bscp-accessserver && make image
	@cd bscp-businessserver && make image
	@cd bscp-templateserver && make image
	@cd bscp-datamanager && make image
	@cd bscp-bcs-controller && make image
	@cd bscp-connserver && make image
	@cd bscp-integrator && make image
