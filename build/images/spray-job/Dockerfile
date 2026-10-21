ARG SPRAY_TAG=master
ARG REPO=kubean-io

FROM ghcr.io/${REPO}/kubespray:${SPRAY_TAG}

WORKDIR /kubespray

COPY playbooks/ /kubespray/

# Add extra collections needed for the playbooks (python deps are preinstalled in the base image)
RUN ansible-galaxy collection install sivel.toiletwater
RUN apk add --no-cache openssl
