FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bscp-ui /bscp-ui
COPY bk-bscp-apiserver /bk-bscp
COPY bk-bscp-authserver /bk-bscp-authserver
COPY bk-bscp-configserver /bk-bscp-configserver
COPY bk-bscp-dataservice /bk-bscp-dataservice
COPY bk-bscp-feedserver /bk-bscp-feedserver
COPY bk-bscp-cacheservice /bk-bscp-cacheservice
ENTRYPOINT ["/bscp-ui"]
CMD []
