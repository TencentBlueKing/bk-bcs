FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bscp-ui /bscp-ui
COPY build/bk-bscp /bk-bscp
ENTRYPOINT ["/bscp-ui"]
CMD []
