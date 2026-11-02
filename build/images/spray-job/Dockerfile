ARG SPRAY_TAG=master
ARG REPO=kubean-io

FROM ghcr.io/${REPO}/kubespray:${SPRAY_TAG}

WORKDIR /kubespray

COPY playbooks/ /kubespray/

# Add extra collections needed for the playbooks (python deps are preinstalled in the base image)
RUN ansible-galaxy collection install sivel.toiletwater
# kubespray 基础镜像在 Dockerfile 未变更时会被复用，此处升级系统包以纳入其后发布的安全修复
RUN apk upgrade --no-cache && apk add --no-cache openssl
