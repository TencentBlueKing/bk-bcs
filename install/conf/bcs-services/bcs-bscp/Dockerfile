FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY hyper/bscp-ui /bscp-ui
COPY hyper/bk-bscp-apiserver /bk-bscp/bk-bscp-apiserver
COPY hyper/bk-bscp-authserver /bk-bscp/bk-bscp-authserver
COPY hyper/bk-bscp-configserver /bk-bscp/bk-bscp-configserver
COPY hyper/bk-bscp-dataservice /bk-bscp/bk-bscp-dataservice
COPY hyper/bk-bscp-feedserver /bk-bscp/bk-bscp-feedserver
COPY hyper/bk-bscp-cacheservice /bk-bscp/bk-bscp-cacheservice
ENTRYPOINT ["/bscp-ui"]
CMD []
