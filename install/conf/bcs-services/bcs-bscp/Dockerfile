FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bin/bscp-ui /bscp-ui
COPY bin/bk-bscp-apiserver /bk-bscp/bk-bscp-apiserver
COPY bin/bk-bscp-authserver /bk-bscp/bk-bscp-authserver
COPY bin/bk-bscp-configserver /bk-bscp/bk-bscp-configserver
COPY bin/bk-bscp-dataservice /bk-bscp/bk-bscp-dataservice
COPY bin/bk-bscp-feedserver /bk-bscp/bk-bscp-feedserver
COPY bin/bk-bscp-cacheservice /bk-bscp/bk-bscp-cacheservice
ENTRYPOINT ["/bscp-ui"]
CMD []
