FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bin/bcs-platform-manager /bcs-platform-manager
ENTRYPOINT ["/bcs-platform-manager"]
CMD []
