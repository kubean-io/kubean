ARG SPRAY_TAG=master
ARG REPO=kubean-io

FROM ghcr.io/${REPO}/kubespray:${SPRAY_TAG}

WORKDIR /kubespray

COPY playbooks/ /kubespray/

# Add extra python packages and collections needed for the playbooks
RUN python3 -m pip install toml
RUN ansible-galaxy collection install sivel.toiletwater
