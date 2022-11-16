cat > manifest.yml << EOF
image_arch:
kube_version:
calico_version:
cni_version:
containerd_version:
cilium_version:
etcd_version:
EOF

if [ -n "${IMAGE_ARCH}" ]; then
  a=0
  for i in ${IMAGE_ARCH}
  do
    export i
    yq -i ".image_arch[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${KUBE_VERSION}" ]; then
  a=0
  for i in ${KUBE_VERSION}
  do
    export i
    yq -i ".kube_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${CALICO_VERSION}" ]; then
  a=0
  for i in ${CALICO_VERSION}
  do
    export i
    yq -i ".calico_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${CNI_VERSION}" ]; then
  a=0
  for i in ${CNI_VERSION}
  do
    export i
    yq -i ".cni_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${CONTAINERD_VERSION}" ]; then
  a=0
  for i in ${CONTAINERD_VERSION}
  do
    export i
    yq -i ".containerd_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${CILIUM_VERSION}" ]; then
  a=0
  for i in ${CILIUM_VERSION}
  do
    export i
    yq -i ".cilium_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if [ -n "${ETCD_VERSION}" ]; then
  a=0
  for i in ${ETCD_VERSION}
  do
    export i
    yq -i ".etcd_version[${a}]=env(i)" manifest.yml
    let a=a+1
  done
fi

if (($(cat manifest.yml|wc -l) < 9 ));then
  echo "There are no updates"
  exit
fi

cat manifest.yml

docker run -v $(pwd)/manifest.yml:/manifest.yml -v $(pwd)/data:/data ghcr.io/hangscer8/airgap-patch:v0.2.2

DATE=$(date "+%Y-%m-%d")
mv manifest.yml kubean-incremental-${DATE}-${FILE_SUFFIX}-manifest.yml
tar zcvf kubean-incremental-${DATE}-${FILE_SUFFIX}.tar.gz $(pwd)/data
sha512sum kubean-incremental-${DATE}-${FILE_SUFFIX}.tar.gz > kubean-incremental-${DATE}-${FILE_SUFFIX}-checksum.txt

sshpass -p ${OFFLINE_NGINX_PASSWORD} scp -o StrictHostKeyChecking=no -rp kubean-incremental-${DATE}-${FILE_SUFFIX}-checksum.txt root@${OFFLINE_NGINX_IP}:/root/release-2.19-offline/offline-files/kubean-incremental-package/ && \
echo "Success Upload checksum file to intranet nginx"
sshpass -p ${OFFLINE_NGINX_PASSWORD} scp -o StrictHostKeyChecking=no -rp kubean-incremental-${DATE}-${FILE_SUFFIX}.tar.gz root@${OFFLINE_NGINX_IP}:/root/release-2.19-offline/offline-files/kubean-incremental-package/ && \
echo "Success Upload incremental offline package to intranet nginx"
sshpass -p ${OFFLINE_NGINX_PASSWORD} scp -o StrictHostKeyChecking=no -rp kubean-incremental-${DATE}-${FILE_SUFFIX}-manifest.yml root@${OFFLINE_NGINX_IP}:/root/release-2.19-offline/offline-files/kubean-incremental-package/ && \
echo "Success Upload incremental offline manifest to intranet nginx"
