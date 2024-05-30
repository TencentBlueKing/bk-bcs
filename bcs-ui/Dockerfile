FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bcs-ui /bcs-ui
ENTRYPOINT ["/bcs-ui"]
CMD []
