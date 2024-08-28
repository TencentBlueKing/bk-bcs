FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY build/bk-bscp/bk-bscp-ui/bk-bscp-ui /bk-bscp/
COPY build/bk-bscp/bk-bscp-apiserver/bk-bscp-apiserver /bk-bscp/
COPY build/bk-bscp/bk-bscp-authserver/bk-bscp-authserver /bk-bscp/
COPY build/bk-bscp/bk-bscp-cacheservice/bk-bscp-cacheservice /bk-bscp/
COPY build/bk-bscp/bk-bscp-configserver/bk-bscp-configserver /bk-bscp/
COPY build/bk-bscp/bk-bscp-dataservice/bk-bscp-dataservice /bk-bscp/
COPY build/bk-bscp/bk-bscp-feedserver/bk-bscp-feedserver /bk-bscp/
COPY build/bk-bscp/bk-bscp-feedproxy/bk-bscp-feedproxy /bk-bscp/
COPY build/bk-bscp/bk-bscp-vaultserver/bk-bscp-vaultserver /bk-bscp/
COPY build/bk-bscp/bk-bscp-vaultserver/vault /bk-bscp/
COPY build/bk-bscp/bk-bscp-vaultserver/vault-sidecar /bk-bscp/
COPY build/bk-bscp/bk-bscp-vaultserver/vault-plugins/bk-bscp-secret /etc/vault/vault-plugins/
ENTRYPOINT ["/bk-bscp/bk-bscp-ui"]
CMD []
