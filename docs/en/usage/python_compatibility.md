# Python versions and image selection

> **Choose the image, then check target-node Python.** Updating `spray-job` does not upgrade Python on target nodes. Meeting the Python requirement does not establish full cluster lifecycle compatibility.

## Requirements for the two images

The control node is the **`spray-job` container** running Ansible. Target nodes are the managed **Kubernetes nodes, including control-plane and worker nodes**.

The following reflects the current source build configuration. The image repository is `ghcr.io/kubean-io/spray-job`; `<version>` denotes a Kubean version. Historical images retain their own dependencies.

> The `2.21` shown below, and its Ansible version, move whenever the pinned Kubespray revision is upgraded. **The image contents are authoritative**, this page may lag behind. Use the command at the end of this section to check.

| Image tag | Ansible / ansible-core | Supported control-node Python | Supported target-node Python |
| --- | --- | --- | --- |
| `<version>-ansible-core2.16`, alias `<version>` | 9.13.0 / 2.16 | 3.10–3.12 | 2.7 or 3.6–3.12 |
| `<version>-ansible-core2.21`, alias `<version>-modern` | Currently 14.4.0 / 2.21 | 3.12–3.14 | 3.9–3.14 |

- Ranges include both endpoints and describe ansible-core support, not the Python version actually installed in an image.
- The core-version tag and its alias refer to the same image build. The core version is read from the image; future upgrades change the explicit tag while `<version>-modern` remains the short alias.
- `modern` follows the pinned Kubespray revision. `latest` also points to the modern image, not the plain version tag. Use an explicit version or digest in production.
- Ansible 9 / core 2.16 is EOL. The older engine is not guaranteed to support every operation in newer Kubespray revisions, such as reset tasks using `mount_facts`.

Check the version actually shipped in an image: `docker run --rm --entrypoint ansible <image> --version`.

## Default Python reference for common systems

The table below covers every offline OS package (`os-pkgs-*`) currently supported by kubean, compiled on 2026-10-08. Rows marked † are not machine-verified; they are the distro's official default or an inference from a closely related distro (e.g. an RHEL-compatible system), and should be confirmed on the actual target node before deployment. Unmarked rows are machine observations. Patch versions and command aliases can vary with OS images, updates and configuration. Check the actual target interpreter; this is not a full OS certification matrix.

| System | os-pkgs tag | `python` | `python3` | Meets core 2.16 target range<br>(2.7 or 3.6–3.12) | Meets core 2.21/modern image target range<br>(3.9–3.14) |
| --- | --- | --- | --- | --- | --- |
| BigCloud 21.10 † | `os-pkgs-bigcloud2110` | Not recorded | 3.6.8 † | Yes | **No** |
| CentOS 7 | `os-pkgs-centos7` | 2.7.5 | Not installed by default (3.6.8 available) | Yes | **No** |
| Kylin V10 SP2 † | `os-pkgs-kylin-v10sp2` | 2.7.18 † | 3.7.9 † | Yes | **No** |
| Kylin V10 SP3 | `os-pkgs-kylin-v10sp3` | 2.7.18 | 3.7.9 | Yes | **No** |
| Kylin V11 2503 (formerly Swan25) | `os-pkgs-kylin-v112503` | Command not found | 3.11.6 | Yes | Yes |
| openEuler 22.03 LTS † | `os-pkgs-openeuler22.03` | Not recorded | 3.9.9 † | Yes | Yes |
| Oracle Linux 8 † | `os-pkgs-oracle8` | Not installed | 3.6.8 † | Yes | **No** |
| Oracle Linux 9 † | `os-pkgs-oracle9` | Not installed | 3.9.18 † | Yes | Yes |
| RHEL 10 † | `os-pkgs-redhat10` | Not installed | 3.12.x † | Yes (right at the upper bound) | Yes |
| RHEL 7 | `os-pkgs-redhat7` | 2.7.5 | Not installed by default (3.6.8 available) | Yes | **No** |
| RHEL 8 | `os-pkgs-redhat8` | Not installed | 3.6.8 | Yes | **No** |
| RHEL 9.2 (Plow) | `os-pkgs-redhat9` | 3.9.16 | 3.9.16 | Yes | Yes |
| Rocky Linux 8 † | `os-pkgs-rocky8` | Not installed | 3.6.8 † | Yes | **No** |
| Rocky Linux 9.2 | `os-pkgs-rocky9` | 3.9.16 | 3.9.16 | Yes | Yes |
| TencentOS Server 3.1 † | `os-pkgs-tencent31` | Not recorded | 3.6.8 † | Yes | **No** |
| Ubuntu 20.04 LTS | `os-pkgs-ubuntu2004` | Command not found | 3.8.10 | Yes | **No** |
| Ubuntu 22.04.4 LTS | `os-pkgs-ubuntu2204` | Command not found | 3.10.12 | Yes | Yes |
| Ubuntu 24.04 LTS | `os-pkgs-ubuntu2404` | Command not found | 3.12.3 | Yes | Yes |

**Quick summary**:

- **Meets the modern image (`-modern` / core 2.21) requirement**, target node can use it directly: Kylin V11 2503, openEuler 22.03, Oracle Linux 9, RHEL 9/10, Rocky Linux 9, Ubuntu 22.04/24.04.
- **Only meets the legacy image (core 2.16) requirement** — default Python is below 3.9. To evaluate the modern image, prepare a supported Python first as described in "Choosing an image": BigCloud 21.10, CentOS 7, Kylin V10 SP2/SP3, Oracle Linux 8, RHEL 7/8, Rocky Linux 8, TencentOS Server 3.1, Ubuntu 20.04.

All Python versions above fall within core 2.16's target-node range. A missing `python` command does not prevent Ansible use; the selected interpreter is what matters.

## Choosing an image

- Targets already have Python 3.9–3.14: reuse that interpreter when evaluating the modern image. **Python 3.12 is not mandatory.**
- RHEL 8 / the Kylin V10 installation above: prepare an additional supported Python before using the modern image, and set `ansible_python_interpreter` to its actual path. Keep the system Python.
- Prepare the interpreter before Python-dependent prechecks and hooks. Offline installations also need OS- and architecture-matched packages with all dependencies; replacing the control image alone is insufficient.

References: [official Ansible support matrix](https://docs.ansible.com/projects/ansible-core/devel/reference_appendices/release_and_maintenance.html#ansible-core-support-matrix), [pinned Kubespray requirements](https://github.com/kubernetes-sigs/kubespray/blob/b33bb4786450395df3c499faa2b7880d4130c693/requirements.txt).
