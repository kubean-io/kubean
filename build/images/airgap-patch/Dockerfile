ARG SPRAY_TAG=master
ARG REPO=kubean-io

FROM ghcr.io/${REPO}/kubespray:${SPRAY_TAG}

WORKDIR /kubean

COPY artifacts /kubean/artifacts

COPY charts /kubean/charts

RUN ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" \
    && apt-get update \
    && apt-get install -y wget curl \
    && wget https://github.com/lework/skopeo-binary/releases/download/v1.9.3/skopeo-linux-$ARCH \
    && chmod +x skopeo-linux-$ARCH \
    && mv skopeo-linux-$ARCH /usr/bin/skopeo \
    && wget https://github.com/mikefarah/yq/releases/download/v4.29.2/yq_linux_$ARCH \
    && chmod +x yq_linux_$ARCH \
    && mv yq_linux_$ARCH /usr/bin/yq \
    && ln -s /kubespray /kubean/kubespray

CMD echo "printing supporting versions" \
    && bash artifacts/manage_version_cr.sh create_manifest \
    && echo $(yq -o=json  '.spec.components' charts/kubean/templates/manifest.cr.yaml) \
    && echo "starting generating airgap patch files" \
    && MANIFEST_CONF=/manifest.yml python3 artifacts/offline_patch.py \
    && echo "copying data" \
    && mv v_* /data \
    && chmod -R 777 /data/ \
    && echo "finished"
