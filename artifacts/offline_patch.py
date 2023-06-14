import os
import re
import shutil
import subprocess
import sys
import yaml
from datetime import datetime
from pathlib import Path

CUR_DIR = os.getcwd()

OPTION = os.getenv("OPTION", default="all")  # create_files create_images create_offlineversion_cr

KUBESPRAY_DIR = os.path.join(CUR_DIR, "kubespray")
MANIFEST_YML_FILE = os.getenv("MANIFEST_CONF", default="manifest.yml")
ZONE = os.getenv("ZONE", default="Other")  ## Other or CN

EXTRA_PAUSE_URLS = os.getenv("EXTRA_PAUSE", "").split(",")  ## registry.k8s.io/pause:3.6

OFFLINE_VER_CR_TEMP = os.getenv("OFFLINEVERSION_CR_TEMPLATE",
                                default=os.path.join(CUR_DIR,
                                                     "artifacts/template/localartifactset.template.yml"))

FILE_LIST_TEMP_PATH = os.path.join(KUBESPRAY_DIR, "contrib/offline/temp/files.list")
IMAGE_LIST_TEMP_PATH = os.path.join(KUBESPRAY_DIR, "contrib/offline/temp/images.list")
KUBEAN_TAG = "v_offline_patch"
print(f"CUR_DIR:{CUR_DIR}")


def extra_line_str_with_pattern(filepath, *patterns):
    result = []
    with open(filepath) as f:
        for line in f:
            for pattern in patterns:
                if re.search(pattern, line) is not None:
                    result.append(line.strip())
    return list(set(result))


def fetch_info_list(env_dict, *patterns):
    print(f"generating info list for {env_dict}")
    os.chdir(KUBESPRAY_DIR)
    if os.path.exists("contrib/offline/temp"):
        shutil.rmtree("contrib/offline/temp")
    cmd = ["bash", "contrib/offline/generate_list.sh"]
    for key, value in env_dict.items():
        cmd.append("-e")
        cmd.append(f"{key}='{value}'")
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(result.stdout)
        print(result.stderr)
        sys.exit(1)
    if not os.path.exists("contrib/offline/temp/images.list"):
        print("not found 'contrib/offline/temp/images.list'")
        sys.exit(1)
    if not os.path.exists("contrib/offline/temp/files.list"):
        print("not found 'contrib/offline/temp/files.list'")
        sys.exit(1)
    file_urls = extra_line_str_with_pattern("contrib/offline/temp/files.list", *patterns)
    image_urls = extra_line_str_with_pattern("contrib/offline/temp/images.list", *patterns)
    os.chdir(CUR_DIR)
    return {"files": file_urls, "images": image_urls}


def check_dependencies():
    if not os.path.exists(KUBESPRAY_DIR):
        print(f"not found kubespray git repo")
        sys.exit(1)
    if subprocess.run(["which", "skopeo"]).returncode != 0:
        print("need skopeo")
        sys.exit(1)


def parse_manifest_yml():
    if (not os.path.exists(MANIFEST_YML_FILE)) or (Path(MANIFEST_YML_FILE).read_text().replace("\n", "").strip() == ""):
        print("MANIFEST_YML_FILE does not exist or empty.")
        sys.exit(1)
    result = {}
    f = open(MANIFEST_YML_FILE)
    result = yaml.load(f, Loader=yaml.loader.FullLoader)  # dict
    f.close()
    return result


def get_manifest_version(key, manifest_dict):
    result = []
    value = manifest_dict.get(key, [])
    if isinstance(value, str):
        result.append(value.strip())
    if isinstance(value, list):
        for v in value:
            result.append(str(v).strip())
    return list(set(result))


def execute_generate_offline_package(arg_option, arch):
    script_name = "generate_offline_package.sh"
    if not os.path.exists("artifacts/generate_offline_package.sh"):
        print("generate_offline_package.sh not found in artifacts")
        sys.exit(1)
    if subprocess.run(["bash", "artifacts/generate_offline_package.sh", "offline_dir"],
                      env={"KUBEAN_TAG": KUBEAN_TAG, "ARCH": arch, "ZONE": ZONE}).returncode != 0:
        print("execute generate_offline_package.sh but failed")
        sys.exit(1)
    if subprocess.run(["bash", "artifacts/generate_offline_package.sh", str(arg_option)],
                      env={"KUBEAN_TAG": KUBEAN_TAG, "ARCH": arch, "ZONE": ZONE}).returncode != 0:
        print("execute generate_offline_package.sh but failed")
        sys.exit(1)


def create_files(file_urls, arch):
    file_content = "\n".join(file_urls)
    if ZONE == "CN":
        file_content = file_content.replace("https://github.com", "https://files.m.daocloud.io/github.com")
    os.chdir(CUR_DIR)
    with open(FILE_LIST_TEMP_PATH, "w") as f:
        f.write(file_content)
        f.flush()
    execute_generate_offline_package("files", arch)


def create_images(image_urls, arch):
    os.chdir(CUR_DIR)
    with open(IMAGE_LIST_TEMP_PATH, "w") as f:
        f.write("\n".join(image_urls))
        f.flush()
    execute_generate_offline_package("images", arch)


def create_offlineversion_cr():
    os.chdir(CUR_DIR)
    if not os.path.exists(OFFLINE_VER_CR_TEMP):
        print("not found kubeanofflineversion template")
        sys.exit(1)
    template_file = open(OFFLINE_VER_CR_TEMP)
    offlineversion_cr_dict = yaml.load(template_file, Loader=yaml.loader.FullLoader)  # dict
    template_file.close()
    offlineversion_cr_dict["spec"]["docker"] = []
    offlineversion_cr_dict["metadata"]["name"] = f"offlineversion-patch-{int(datetime.now().timestamp())}"
    items_array = offlineversion_cr_dict["spec"]["items"]

    for index in range(len(items_array)):
        item_dict = items_array[index]
        if item_dict["name"] == "cni":
            item_dict["versionRange"] = cni_versions
        if item_dict["name"] == "containerd":
            item_dict["versionRange"] = containerd_versions
        if item_dict["name"] == "kube":
            item_dict["versionRange"] = kube_versions
        if item_dict["name"] == "calico":
            item_dict["versionRange"] = calico_versions
        if item_dict["name"] == "cilium":
            item_dict["versionRange"] = cilium_versions
        if item_dict["name"] == "etcd":
            item_dict["versionRange"] = etcd_versions
        items_array[index] = item_dict

    offlineversion_cr_dict["spec"]["items"] = items_array
    kubeanofflineversion_file = open(
        os.path.join(KUBEAN_TAG, "kubeanofflineversion.cr.patch.yaml"),
        "w",
        encoding="utf-8")
    yaml.dump(offlineversion_cr_dict, kubeanofflineversion_file)
    kubeanofflineversion_file.close()


def add_pause_image(origin_image_urls):
    pause_images = []
    for image_name in origin_image_urls:
        if "registry.k8s.io/pause" in image_name:
            new_version = format(float(image_name.split(":")[1]) - float(0.1), '.1f')
            pause_images.append("registry.k8s.io/pause:" + new_version)
    return list(pause_images)


image_archs = []
cni_versions = []
containerd_versions = []
kube_versions = []
calico_versions = []
cilium_versions = []
etcd_versions = []

if __name__ == '__main__':
    print(f"OPTION:{OPTION}")
    print(f"ZONE:{ZONE}")
    check_dependencies()
    manifest_dict = parse_manifest_yml()
    # global value setting
    image_archs = get_manifest_version("image_arch", manifest_dict=manifest_dict)
    cni_versions = get_manifest_version("cni_version", manifest_dict=manifest_dict)
    containerd_versions = get_manifest_version("containerd_version", manifest_dict=manifest_dict)
    kube_versions = get_manifest_version("kube_version", manifest_dict=manifest_dict)
    calico_versions = get_manifest_version("calico_version", manifest_dict=manifest_dict)
    cilium_versions = get_manifest_version("cilium_version", manifest_dict=manifest_dict)
    etcd_versions = get_manifest_version("etcd_version", manifest_dict=manifest_dict)
    # global value setting

    for image_arch in image_archs:
        print(f"operating for {image_arch}")
        file_urls = []
        images_urls = []
        for kube_version in kube_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "kube_version": kube_version},
                                         r"kubernetes-release.*/kube.*", r"registry.k8s.io/kube-.*",
                                         r"registry.k8s.io/pause.*", r".*crictl.*", r".*kubelet.*",
                                         r".*kubectl.*", r".*kubeadm.*", r".*coredns.*", r".*etcd.*")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
            images_urls = images_urls + add_pause_image(images_urls)
            images_urls = images_urls + list(EXTRA_PAUSE_URLS)
        for cni_version in cni_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "cni_version": cni_version},
                                         r"containernetworking.*cni-.*")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
        for containerd_version in containerd_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "containerd_version": containerd_version},
                                         r"containerd.*containerd-")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
        for calico_version in calico_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "calico_version": calico_version}, r"calico.*")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
        for cilium_version in cilium_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "cilium_version": cilium_version}, r"cilium.*")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
        for etcd_version in etcd_versions:
            tuple_data = fetch_info_list({"image_arch": image_arch, "etcd_version": etcd_version}, r"etcd.*")
            file_urls = list(file_urls) + list(tuple_data["files"])
            images_urls = list(images_urls) + list(tuple_data["images"])
        file_urls = list(set(file_urls))
        images_urls = filter(lambda item: item != "", images_urls)
        images_urls = list(set(images_urls))
        if "registry.k8s.io/pause:3.4" in images_urls:
            images_urls.remove("registry.k8s.io/pause:3.4")
        print("file_urls:")
        print(file_urls)
        print("")
        print("images_urls:")
        print(images_urls)
        print("")
        if OPTION == "all":
            create_files(file_urls, arch=image_arch)
            create_images(images_urls, arch=image_arch)
            create_offlineversion_cr()
            execute_generate_offline_package("copy_import_sh", arch=image_arch)
        if OPTION == "create_files":
            create_files(file_urls, arch=image_arch)
        if OPTION == "create_images":
            create_images(images_urls, arch=image_arch)
        if OPTION == "create_offlineversion_cr":
            create_offlineversion_cr()
