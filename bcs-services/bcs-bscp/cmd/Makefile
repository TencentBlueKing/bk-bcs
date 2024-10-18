all:
	@cd api-server && make
	@cd config-server && make
	@cd data-service && make
	@cd feed-server && make
	@cd auth-server && make
	@cd cache-service && make
	@cd vault-server && make
	@cd feed-proxy && make
	@cd ui && make

ui:
	@cd ui && make

feed:
	@cd feed-server && make
	@cd cache-service && make
	@cd feed-proxy && make

server:
	@cd api-server && make package
	@cd config-server && make package
	@cd data-service && make package
	@cd feed-server && make package
	@cd auth-server && make package
	@cd cache-service && make package

clean:
	@cd api-server && make clean
	@cd config-server && make clean
	@cd data-service && make clean
	@cd feed-server && make clean
	@cd auth-server && make clean
	@cd cache-service && make clean
