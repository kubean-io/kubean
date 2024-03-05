ARG SPRAY_TAG=master
ARG REPO=kubean-io
FROM ghcr.io/${REPO}/kubespray:${SPRAY_TAG}

ARG SPRAY_RELEASE
ARG SPRAY_COMMIT
ARG SPRAY_COMMIT_TIMESTAMP

LABEL spray_release_num=${SPRAY_RELEASE}
LABEL spray_commit_date=${SPRAY_COMMIT_TIMESTAMP}

ENV SPRAY_COMMIT=${SPRAY_COMMIT}
ENV SPRAY_RELEASE=${SPRAY_RELEASE}

WORKDIR /kubean

COPY artifacts /kubean/artifacts
COPY charts /kubean/charts

ARG SKOPEO_VERSION="v1.11.2"
ARG YQ_VERSION="v4.33.3"
RUN ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" \
    && apk add --update --no-cache curl wget \
    && echo "install skopeo" \
    && wget -O /usr/bin/skopeo https://github.com/lework/skopeo-binary/releases/download/${SKOPEO_VERSION}/skopeo-linux-${ARCH} \
    && chmod +x /usr/bin/skopeo \
    && echo "install yq" \
    && wget -O /usr/bin/yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${ARCH} \
    && chmod +x /usr/bin/yq \
    && ln -s /kubespray /kubean/kubespray

CMD echo "show supporting versions" \
    && ./artifacts/gen_airgap_crs.py Manifest \
    && echo $(yq -o=json '.spec.components' charts/kubean/templates/manifest.cr.yaml) \
    && echo "starting generating airgap patch files" \
    && MANIFEST_CONF=/manifest.yml python3 artifacts/airgap_patch.py \
    && echo "copying data" \
    && rm -rf /data/* \
    && mv airgap_patch/* /data \
    && chmod -R 777 /data/ \
    && echo "finished"
