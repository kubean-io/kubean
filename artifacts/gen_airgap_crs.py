#!/usr/bin/env python

import sys
import os
import yaml
from pathlib import Path
from cr_template import CR_Template

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

KUBEAN_TAG = os.getenv("KUBEAN_TAG", default="")
KUBE_VERSION = os.getenv("KUBE_VERSION", default="")
SPRAY_DIR = os.getenv("SPRAY_DIR", default="kubespray")
SPRAY_COMMIT = os.getenv("SPRAY_COMMIT", default="")
SPRAY_RELEASE = os.getenv("SPRAY_RELEASE", default="master")
SPRAY_COMMIT_TIMESTAMP = os.getenv("SPRAY_COMMIT_TIMESTAMP", default="")

COMPONENTS_KEYS = [
  {"name": "cilium", "checksumsKey": None},
  {"name": "flannel", "checksumsKey": None},
  {"name": "kube_ovn", "checksumsKey": None},
  {"name": "runc", "checksumsKey": "runc_checksums.amd64"},
  {"name": "kube", "checksumsKey": "kubelet_checksums.amd64"},
  {"name": "cni", "checksumsKey": "cni_binary_checksums.amd64"},
  {"name": "calico", "checksumsKey": "calico_crds_archive_checksums"},
  {"name": "containerd", "checksumsKey": "containerd_archive_checksums.amd64"},
]

DOCKER_KEYS = ["redhat-7", "redhat", "debian", "ubuntu", "kylin"]


def merge_dir_content_to_file(input_dir_paths, output_file_path):
  merged_content = []
  for input_dir_path in input_dir_paths:
    if not os.path.isdir(input_dir_path):
      print(f"{input_dir_path} is an invalid path.")
      sys.exit(1)
    file_list = os.listdir(input_dir_path)
    with open(output_file_path, "w") as output_file:
      for file_name in file_list:
        file_path = os.path.join(input_dir_path, file_name)
        print(f"file to be merged: {file_path}")
        if os.path.isfile(file_path):
          with open(file_path, "r") as input_file:
            lines = [line for line in input_file.readlines() if "---" not in line]
            merged_content += lines
  with open(output_file_path, "w") as output_file:
    output_file.writelines(merged_content)
  print(f"Merged content has been written to: {output_file_path}")


def merge_spray_components_version_files(merged_file_path):
  # Used to retrieve the kube_version line
  sprayDefaultPath = f"{SPRAY_DIR}/roles/kubespray-defaults/defaults"
  # The directory first appeared in release 2.24
  checkpointPath1 = f"{SPRAY_DIR}/roles/kubespray-defaults/defaults/main"
  # The directory first appeared in release 2.22
  checkpointPath2 = f"{SPRAY_DIR}/roles/download/defaults/main"
  # The file was removed in release 2.23
  checkpointPath3 = f"{SPRAY_DIR}/roles/download/defaults"

  if os.path.exists(checkpointPath1):
    print("checkpoint path 1")
    path = Path(f"{SPRAY_DIR}/roles/download/defaults")
    path.mkdir(parents=True, exist_ok=True)
    merge_dir_content_to_file([checkpointPath1, sprayDefaultPath], merged_file_path)
  elif os.path.exists(f"{checkpointPath2}/main.yml"):
    print("checkpoint path 2")
    merge_dir_content_to_file([checkpointPath2, sprayDefaultPath], merged_file_path)
  elif os.path.exists(f"{checkpointPath3}/main.yml"):
    print("checkpoint path 3")
    merge_dir_content_to_file([checkpointPath3, sprayDefaultPath], merged_file_path)


def get_value_from_yml(yml_file_path, key):
  keys = key.split('.')
  with open(yml_file_path, "r") as file:
    value = yaml.safe_load(file)
    for k in keys:
        if isinstance(value, dict) and k in value:
            value = value[k]
        else:
          print(f"The '{key}' key was not found in the file: {yml_file_path}.")
          sys.exit(1)
    return value


def create_localartifactset():
  """
    Create LocalArtifactSet custom resource manifest
  """
  print("Prepare to create LocalArtifactSet")

  if not os.path.exists(SPRAY_DIR):
    print(f"{SPRAY_DIR} directory doesn't exist.")
    sys.exit(1)
  
  comps_version_conf_file = f"{SPRAY_DIR}/components_version_config.yml"
  merge_spray_components_version_files(comps_version_conf_file)

  components = { key_item["name"]: [] for key_item in COMPONENTS_KEYS }
  verison = KUBE_VERSION
  for key in components:
    if key != "kube" or KUBE_VERSION == "":
      verison = get_value_from_yml(comps_version_conf_file, f"{key}_version")
    components[key].append(verison)

  path = Path(KUBEAN_TAG)
  path.mkdir(parents=True, exist_ok=True)
  KUBEAN_LOCALARTIFACTSET_CR=f"{KUBEAN_TAG}/localartifactset.cr.yaml"
  cr = CR_Template(KUBEAN_TAG, SPRAY_INFO, components, {}, False)
  cr.render_template(cr.CR_LOCALARTIFACTSET_TEMPLATE, KUBEAN_LOCALARTIFACTSET_CR)


def create_manifest():
  """
    Create Manifest custom resource manifest
  """
  print("Prepare to create Manifest")

  if not os.path.exists(SPRAY_DIR):
    print(f"{SPRAY_DIR} directory doesn't exist.")
    sys.exit(1)

  comps_version_conf_file = f"{SPRAY_DIR}/components_version_config.yml"
  merge_spray_components_version_files(comps_version_conf_file)

  components = {key_item["name"]: {"defaultVersion": "", "versionRange": []} for key_item in COMPONENTS_KEYS}
  verison = KUBE_VERSION
  for key in components:
    if key != "kube" or KUBE_VERSION == "":
      verison = get_value_from_yml(comps_version_conf_file, f"{key}_version")
    components[key]["defaultVersion"] = verison
    checksumsKey = next(key_item["checksumsKey"] for key_item in COMPONENTS_KEYS if key_item["name"] == key)
    components[key]["versionRange"] = [] if checksumsKey is None else [key for key in get_value_from_yml(comps_version_conf_file, checksumsKey)]

  dockers = { key_item: {"defaultVersion": "", "versionRange": []} for key_item in DOCKER_KEYS }
  default_docker_verison = get_value_from_yml(f"{SPRAY_DIR}/roles/container-engine/docker/defaults/main.yml", "docker_version")
  for key in dockers:
    dockers[key]["defaultVersion"] = default_docker_verison if key != "kylin" else "19.03"
    docker_versioned_pkg = get_value_from_yml(f"{SPRAY_DIR}/roles/container-engine/docker/vars/{key}.yml", "docker_versioned_pkg")
    versions = [verison for verison in docker_versioned_pkg if verison not in ["latest", "stable", "edge"] ]
    dockers[key]["versionRange"] = versions

  KUBEAN_LOCALARTIFACTSET_CR="charts/kubean/templates/manifest.cr.yaml"
  cr = CR_Template(KUBEAN_TAG, SPRAY_INFO, components, dockers, False)
  cr.render_template(cr.CR_MANIFEST_TEMPLATE, KUBEAN_LOCALARTIFACTSET_CR)


if __name__ == '__main__':
  global SPRAY_INFO
  SPRAY_INFO = {
    "sprayRlease": SPRAY_RELEASE,
    "sprayCommit": SPRAY_COMMIT,
    "sprayCommitShort": SPRAY_COMMIT[0:7],
    "sprayCommitTimestamp": SPRAY_COMMIT_TIMESTAMP,
  }
  option = sys.argv[1]
  if option == "LocalArtifactSet":
    create_localartifactset()
  elif option == "Manifest":
    create_manifest()
  else:
    print("Unknown operation")
