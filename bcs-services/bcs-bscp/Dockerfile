FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bscp-ui /bscp-ui
ENTRYPOINT ["/bscp-ui"]
CMD []
