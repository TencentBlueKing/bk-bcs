FROM tencentos/tencentos4-minimal
COPY hyper/bscp-ui /bscp-ui
COPY hyper/bk-bscp-apiserver /bk-bscp/bk-bscp-apiserver
COPY hyper/bk-bscp-authserver /bk-bscp/bk-bscp-authserver
COPY hyper/bk-bscp-configserver /bk-bscp/bk-bscp-configserver
COPY hyper/bk-bscp-dataservice /bk-bscp/bk-bscp-dataservice
COPY hyper/bk-bscp-feedserver /bk-bscp/bk-bscp-feedserver
COPY hyper/bk-bscp-cacheservice /bk-bscp/bk-bscp-cacheservice
COPY hyper/bk-bscp-vaultserver /bk-bscp/bk-bscp-vaultserver
COPY hyper/vault /bk-bscp/vault
COPY hyper/vault-sidecar /bk-bscp/vault-sidecar
COPY hyper/bk-bscp-secret /etc/vault/vault-plugins/bk-bscp-secret
ENTRYPOINT ["/bscp-ui"]
CMD []
