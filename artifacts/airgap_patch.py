#!/usr/bin/env python

import os
import re
import subprocess
import sys
import yaml
import json
from pathlib import Path
from shutil import which, rmtree
from cr_template import CR_Template

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

CUR_DIR = os.getcwd()
KUBEAN_TAG = "airgap_patch"
MODE = os.getenv("MODE", default="INCR")  ## INCR or FULL
ZONE = os.getenv("ZONE", default="DEFAULT")  ## DEFAULT or CN
OPTION = os.getenv("OPTION", default="all")  # all create_files create_images

SPRAY_COMMIT = os.getenv("SPRAY_COMMIT", default="")
SPRAY_RELEASE = os.getenv("SPRAY_RELEASE", default="master")
SPRAY_COMMIT_TIMESTAMP = os.getenv("SPRAY_COMMIT_TIMESTAMP", default="")

SPRAY_REPO_PATH = os.path.join(CUR_DIR, "kubespray")

OFFLINE_TMP_REL_PATH = "contrib/offline/temp"
OFFLINE_TMP_ABS_PATH = os.path.join(SPRAY_REPO_PATH, OFFLINE_TMP_REL_PATH)

KEYWORDS = {
    "kube_version": [
        "kubelet", "kubectl", "kubeadm", "kube-apiserver",
        "kube-controller-manager", "kube-scheduler", "kube-proxy",
        "etcd", "pause", "coredns", "crictl", "cri-o",
    ],
    "cni_version": ["cni"],
    "containerd_version": ['containerd'],
    "calico_version": ['calico'],
    "cilium_version": ['cilium'],
    "runc_version": ['runc'],
}

def file_lines_to_list(filename):
    with open(filename) as file:
        return [line.rstrip() for line in file]

def get_list_include_keywords(list, *keywords):
    result = []
    for line in list:
        for keyword in keywords:
            if keyword in line:
                result.append(line.strip())
    return result

def check_dependencies():
    if not os.path.exists(SPRAY_REPO_PATH):
        print("kubespray repo path not found")
        sys.exit(1)
    if which("skopeo") is None:
        print("skopeo command not found")
        sys.exit(1)

def get_manifest_version(key, manifest_dict):
    result = []
    value = manifest_dict.get(key, [])
    if isinstance(value, str):
        result.append(value.strip())
    if isinstance(value, list):
        for v in value:
            result.append(str(v).strip())
    return list(set(result))

def execute_gen_airgap_pkgs(arg_option, arch):
    if not os.path.exists("artifacts/gen_airgap_pkgs.sh"):
        print("gen_airgap_pkgs.sh not found in artifacts")
        sys.exit(1)
    if subprocess.run(["bash", "artifacts/gen_airgap_pkgs.sh", "offline_dir"],
                      env={"KUBEAN_TAG": KUBEAN_TAG, "ARCH": arch, "ZONE": ZONE}).returncode != 0:
        print("execute gen_airgap_pkgs.sh but failed")
        sys.exit(1)
    if subprocess.run(["bash", "artifacts/gen_airgap_pkgs.sh", str(arg_option)],
                      env={"KUBEAN_TAG": KUBEAN_TAG, "ARCH": arch, "ZONE": ZONE}).returncode != 0:
        print("execute gen_airgap_pkgs.sh but failed")
        sys.exit(1)

def create_files(file_urls, arch):
    os.chdir(CUR_DIR)
    with open(os.path.join(OFFLINE_TMP_ABS_PATH, "files.list"), "w") as f:
        f.write("\n".join(file_urls))
        f.flush()
    execute_gen_airgap_pkgs("files", arch)

def create_images(image_urls, arch):
    os.chdir(CUR_DIR)
    with open(os.path.join(OFFLINE_TMP_ABS_PATH, "images.list"), "w") as f:
        f.write("\n".join(image_urls))
        f.flush()
    execute_gen_airgap_pkgs("images", arch)

def create_localartifactset_cr(manifest_data):
    spray_info = {
        "sprayRlease": SPRAY_RELEASE if SPRAY_RELEASE != "" else "master",
        "sprayCommit": SPRAY_COMMIT,
        "sprayCommitShort": SPRAY_COMMIT[0:7],
        "sprayCommitTimestamp": SPRAY_COMMIT_TIMESTAMP,
    }
    components = { re.split('_', key_item)[0]: [] for key_item in KEYWORDS }
    for key in components:
        versions = manifest_data.get(f"{key}_version")
        if MODE == "FULL" and key != 'kube' and versions is None:
            versions = ['default']
        if isinstance(versions, list):
            components[key] = versions
        if isinstance(versions, str):
            components[key].append(versions)

    path = Path(KUBEAN_TAG)
    path.mkdir(parents=True, exist_ok=True)
    KUBEAN_LOCALARTIFACTSET_CR=f"{KUBEAN_TAG}/localartifactset.cr.yaml"
    cr = CR_Template(KUBEAN_TAG, spray_info, components, {}, True)
    cr.render_template(cr.CR_LOCALARTIFACTSET_TEMPLATE, KUBEAN_LOCALARTIFACTSET_CR)

def get_manifest_data():
    manifest_yml_file = os.getenv("MANIFEST_CONF", default="manifest.yml")
    if (not os.path.exists(manifest_yml_file)) or (Path(manifest_yml_file).read_text().replace("\n", "").strip() == ""):
        print("manifest yaml file does not exist or empty.")
        sys.exit(1)
    with open(manifest_yml_file, "r") as stream:
        return yaml.safe_load(stream)

def get_other_required_keywords(manifest_dict):
    other_required_keywords = [
        "crun", "runsc", "cri-dockerd", "yq", "nginx", "k8s-dns-node-cache", "cluster-proportional-autoscaler"]
    manifest_keys = [ key for key in manifest_dict]
    keys_range = [ key for key in KEYWORDS]
    list_diff = list(set(keys_range) - set(manifest_keys))
    print(f'- keys_range: {keys_range}\n- manifest_keys: {manifest_keys}\n- list_diff: {list_diff}\n')
    for key in list_diff:
        other_required_keywords += KEYWORDS[key]
    return other_required_keywords

def build_jobs_params(manifest_dict):
    print(f'- manifest_dict: {manifest_dict}\n')
    max_len = max(len(item) for _, item in manifest_dict.items() if isinstance(item, list))
    other_required_keywords = get_other_required_keywords(manifest_dict)
    jobs_params = {
        "arch": manifest_dict.get('image_arch', ['amd64']),
        "jobs": [{"keywords": [], "extra_vars": [],} for i in range(max_len)],
        "other_keywords": other_required_keywords,
    }
    manifest_keys=['image_arch']
    manifest_keys += [ key for key in KEYWORDS]
    for index, job in enumerate(jobs_params.get('jobs', [])):
        for component, versions in manifest_dict.items():
            if component not in manifest_keys:
                print(f"unknown component version key: {component}")
                sys.exit(1)
            if isinstance(versions, str) and index == 0 and component != "image_arch":
                job['keywords'] += KEYWORDS.get(component, [])
                job['extra_vars'].append(f"{component}='{versions}'")
            if isinstance(versions, list) and index < len(versions) and component != "image_arch":
                job['keywords'] += KEYWORDS.get(component, [])
                job['extra_vars'].append(f"{component}='{versions[index]}'")
    print(f'- jobs_params: {json.dumps(jobs_params, indent=2)}\n')
    return jobs_params

def gen_airgap_packages(option, arch, bin_urls, img_urls):
    if option == "all":
        create_files(bin_urls, arch=arch)
        create_images(img_urls, arch=arch)
        execute_gen_airgap_pkgs("copy_import_sh", arch=arch)
    if option == "create_files":
        create_files(bin_urls, arch=arch)
    if option == "create_images":
        create_images(img_urls, arch=arch)

def batch_gen_airgap_resources(jobs_params):
    other_required_list = {key: [] for key in ['file_list', 'image_list']}
    list_data = {key: [] for key in ['file_list', 'image_list']}
    is_executed = False
    for arch in jobs_params.get('arch', []):
        for job in jobs_params.get('jobs',[]):
            extra_vars_cmd = []
            for var in job.get('extra_vars'):
                extra_vars_cmd.extend(["-e", var])
            os.chdir(SPRAY_REPO_PATH)
            if os.path.exists(f"{OFFLINE_TMP_REL_PATH}"):
                rmtree(f"{OFFLINE_TMP_REL_PATH}")
            cmd = ["bash", "contrib/offline/generate_list.sh", "-e", f"image_arch='{arch}'"]
            cmd += extra_vars_cmd
            print(f"\n- cmd: {cmd}\n")
            result = subprocess.run(cmd, capture_output=True, text=True)
            if result.returncode != 0:
                print(result.stdout)
                print(result.stderr)
                sys.exit(1)
            if not os.path.exists(f"{OFFLINE_TMP_REL_PATH}/images.list"):
                print(f"not found '{OFFLINE_TMP_REL_PATH}/images.list'")
                sys.exit(1)
            if not os.path.exists(f"{OFFLINE_TMP_REL_PATH}/files.list"):
                print(f"not found '{OFFLINE_TMP_REL_PATH}/files.list'")
                sys.exit(1)

            files_list = file_lines_to_list(f"{OFFLINE_TMP_REL_PATH}/files.list")
            images_list = file_lines_to_list(f"{OFFLINE_TMP_REL_PATH}/images.list")
            file_urls, image_urls = get_list_include_keywords(files_list, *job.get('keywords')), get_list_include_keywords(images_list, *job.get('keywords'))
            if MODE == "FULL" and not is_executed:
                other_required_keywords = jobs_params.get('other_keywords', [])
                other_required_list['file_list'] = get_list_include_keywords(files_list, *other_required_keywords)
                other_required_list['image_list'] = get_list_include_keywords(images_list, *other_required_keywords)
            is_executed = True

            os.chdir(CUR_DIR)
            list_data['file_list'] += file_urls
            list_data['image_list'] += image_urls

        list_data['file_list'] += other_required_list['file_list']
        list_data['image_list'] += other_required_list['image_list']
        list_data['file_list'], list_data['image_list'] = list(set(list_data['file_list'])), list(set(list_data['image_list']))
        print_list(list_data['file_list'],  list_data['image_list'])
        gen_airgap_packages(OPTION, arch, list_data['file_list'], list_data['image_list'])

def print_list(file_list, image_list):
    print("---------------- file urls -----------------\n")
    for file_url in file_list:
        print(f'* {file_url}\n')
    print("---------------- image urls -----------------\n")
    for image_url in image_list:
        print(f'* {image_url}\n')

if __name__ == '__main__':
    print(f"OPTION:{OPTION}, ZONE: {ZONE}, MODE: {MODE}\n")
    check_dependencies()
    manifest_data = get_manifest_data()
    batch_gen_airgap_resources(build_jobs_params(manifest_data))
    create_localartifactset_cr(manifest_data)
