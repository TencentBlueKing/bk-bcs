FROM alpine
RUN apk --update --no-cache add ca-certificates bash vim curl
COPY bcs-monitor /bcs-monitor
ENTRYPOINT ["/bcs-monitor"]
CMD []
