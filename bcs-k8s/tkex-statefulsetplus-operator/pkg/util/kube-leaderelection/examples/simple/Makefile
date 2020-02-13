server:
	go build -a -installsuffix cgo -ldflags '-w' -o server main.go

container: server
	@docker build -t csighub.tencentyun.com/admin/kube-leaderelection-example:1.0.0 .

push: container
	@docker push  csighub.tencentyun.com/admin/kube-leaderelection-example:1.0.0

run: container
	@kubectl run kube-leaderelection-example --image csighub.tencentyun.com/admin//kube-leaderelection-example:0.0.1 --replicas=3

clean:
	@rm -f server
	#@kubectl delete deploy kube-leaderelection-example


