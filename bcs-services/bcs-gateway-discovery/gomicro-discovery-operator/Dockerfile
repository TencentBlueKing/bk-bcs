FROM golang:1.20 AS builder

COPY ./ /app
WORKDIR /app

ARG BINARY=gomicro-discovery-operator

RUN make build && chmod +x ./build/${BINARY}

FROM debian:bullseye-slim

RUN mkdir /app
COPY --from=builder /app/build/${BINARY} /app/${BINARY}
RUN chmod 755 /app/${BINARY}

CMD ["/app/gomicro-discovery-operator"]