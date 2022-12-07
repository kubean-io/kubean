FROM alpine:3.17 as spray-build

ARG SPRAY_REF=master

ARG TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apk add --update --no-cache python3 openssh curl sshpass rsync bash \
    gcc musl-dev python3-dev py3-pip git

RUN git clone https://github.com/kubernetes-sigs/kubespray.git /kubespray \
    && cd /kubespray \
    && git checkout $SPRAY_REF

WORKDIR /kubespray

# Some tools like yamllint need this
# Pip needs this as well at the moment to install ansible
# (and potentially other packages)
# See: https://github.com/pypa/pip/issues/10219
ENV LANG=C.UTF-8

RUN /usr/bin/python3 -m pip install --no-cache-dir pip -U \
    && python3 -m pip install --no-cache-dir -r requirements.txt

RUN ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" \
    && KUBE_VERSION=$(sed -n 's/^kube_version: //p' roles/kubespray-defaults/defaults/main.yaml) \
    && curl -LO https://storage.googleapis.com/kubernetes-release/release/$KUBE_VERSION/bin/linux/$ARCH/kubectl \
    && curl -LO https://github.com/mikefarah/yq/releases/download/v4.25.2/yq_linux_$ARCH \
    && chmod a+x kubectl yq_linux_$ARCH \
    && mv kubectl /usr/local/bin/kubectl \
    && mv yq_linux_$ARCH /usr/local/bin/yq

RUN pip3 cache purge \
    && apk del --quiet gcc musl-dev python3-dev py3-pip git \
    && find / -type d -name '*__pycache__' -print0 | xargs -0 -r rm -rf \
    && bash -c "rm -rf /kubespray/{.git,docs}"

FROM scratch
COPY --from=spray-build / /

ENV ANSIBLE_CONFIG=/kubespray/ansible.cfg
WORKDIR /kubespray