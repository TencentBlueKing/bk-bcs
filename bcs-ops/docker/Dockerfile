FROM python:3.8-alpine3.18 as builder
WORKDIR /work
RUN apk update && apk add make
COPY . .
RUN make build
RUN cp bcs-ops-script-release-*.tar.gz sops/bcs-ops.tar.gz

FROM python:3.8-alpine3.18
ENV WORKSPACE=/app
ENV PATH=$PATH:${WORKSPACE}
WORKDIR ${WORKSPACE}
RUN pip3 install ujson requests
COPY --from=builder /work/sops/upload.py /work/sops/bcs-ops.tar.gz /work/sops/bcs_bk_sops_common.dat ./
ENTRYPOINT ["python3", "upload.py"]
