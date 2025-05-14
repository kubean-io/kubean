FROM alpine:3.17 as spray-build

ARG SPRAY_REF=master

ARG TZ=Etc/UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apk add --update --no-cache python3 openssh curl sshpass rsync bash \
    gcc musl-dev python3-dev libffi-dev py3-pip git patch

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
    && if [ -d 'roles/kubespray-defaults/defaults/main' ]; then \
         KUBE_VERSION=$(sed -n 's/^kube_version: //p' roles/kubespray-defaults/defaults/main/main.yml); \
         if [[ "$KUBE_VERSION" == *{{* ]]; then \
           KUBE_VERSION=$(python3 -c "import yaml,pathlib;f=pathlib.Path('roles/kubespray-defaults/defaults/main/checksums.yml');d=yaml.safe_load(f.read_text());print(list(d['kubelet_checksums']['amd64'].keys())[0])"); \
         fi \
       elif [ -d 'roles/kubespray_defaults/defaults/main' ]; then \
         KUBE_VERSION=$(sed -n 's/^kube_version: //p' roles/kubespray_defaults/defaults/main/main.yml); \
         if [[ "$KUBE_VERSION" == *{{* ]]; then \
           KUBE_VERSION=$(python3 -c "import yaml,pathlib;f=pathlib.Path('roles/kubespray_defaults/defaults/main/checksums.yml');d=yaml.safe_load(f.read_text());print(list(d['kubelet_checksums']['amd64'].keys())[0])"); \
         fi \
       else \
         KUBE_VERSION=$(sed -n 's/^kube_version: //p' roles/kubespray-defaults/defaults/main.yaml); \
       fi \
    && curl -LO https://dl.k8s.io/release/v$KUBE_VERSION/bin/linux/$ARCH/kubectl \
    && curl -LO https://github.com/mikefarah/yq/releases/download/v4.44.2/yq_linux_$ARCH \
    && chmod a+x kubectl yq_linux_$ARCH \
    && mv kubectl /usr/local/bin/kubectl \
    && mv yq_linux_$ARCH /usr/local/bin/yq

RUN pip3 cache purge \
    && apk del --quiet gcc musl-dev python3-dev libffi-dev git \
    && find / -type d -name '*__pycache__' -print0 | xargs -0 -r rm -rf \
    && bash -c "rm -rf /kubespray/{.git,docs}"

COPY build/images/ansible_patchs/* /tmp/
RUN ANSIBLE_CORE_VERSION=$( pip show -V ansible-core | grep Version: | awk '{print $2}' | cut -d. -f1-2) \
    && echo "ANSIBLE_CORE_VERSION: ${ANSIBLE_CORE_VERSION}" \
    && SITE_PKG_PATH=$(python3 -c 'import sysconfig; print(sysconfig.get_paths()["purelib"])') \
    && PKG_MGR_PATH="${SITE_PKG_PATH}/ansible/module_utils/facts/system/pkg_mgr.py" \
    && echo "PKG_MGR_PATH: ${PKG_MGR_PATH}" \
    && if [[ "${ANSIBLE_CORE_VERSION}" == "2.14" ]]; then \
            echo "patch-2.14"; patch ${PKG_MGR_PATH} /tmp/pkg_mgr_2.14.patch; \
        elif [[ "${ANSIBLE_CORE_VERSION}" == "2.15" ]]; then  \
            echo "patch-2.15"; patch ${PKG_MGR_PATH} /tmp/pkg_mgr_2.15.patch; \
        fi

FROM scratch

ARG DOCKERFILE_COMMIT

LABEL io.kubean.dockerfile-commit=${DOCKERFILE_COMMIT}

COPY --from=spray-build / /

ENV ANSIBLE_CONFIG=/kubespray/ansible.cfg
WORKDIR /kubespray
