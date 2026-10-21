ARG BKAI_INIT_IMAGE=hub.bktencent.com/blueking/bk-aidev-init:0.1.5
FROM ${BKAI_INIT_IMAGE}

COPY --chown=10001:10001 ./bkai.yaml /bk-bcs/bkai.yaml
COPY --chown=10001:10001 ./agents /bk-bcs/agents
COPY --chown=10001:10001 ./skills /bk-bcs/skills
COPY --chown=10001:10001 ./knowledgebases /bk-bcs/knowledgebases
