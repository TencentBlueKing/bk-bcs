FROM alpine:3.6

RUN apk add --no-cache ca-certificates

ADD ./bcs-kube-agent /bcs-kube-agent

RUN adduser -D bcs-kube-agent
USER bcs-kube-agent

ENTRYPOINT ["/bcs-kube-agent"]