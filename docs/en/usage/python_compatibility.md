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

These machine observations and supplemental information were compiled on 2026-09-28. Patch versions and command aliases can vary with OS images, updates and configuration. Check the actual target interpreter; this is not a full OS certification matrix.

| System | `python` | `python3` | Does this Python 3 meet the core 2.21 range? |
| --- | --- | --- | --- |
| Ubuntu 22.04.4 LTS | Command not found | 3.10.12 | Yes |
| Ubuntu 24.04 LTS | Command not found | 3.12.3 | Yes |
| Rocky Linux 9.2 | 3.9.16 | 3.9.16 | Yes |
| Kylin V10 (configured as SP3) | 2.7.18 | 3.7.9 | **No** |
| Kylin V11 (Swan25) | Command not found | 3.11.6 | Yes |
| RHEL 9.2 (Plow) | 3.9.16 | 3.9.16 | Yes |
| RHEL 8 | Not recorded | 3.6.8 (supplemental default version) | **No** |

All Python 3 versions above fall within core 2.16's target-node range. A missing `python` command does not prevent Ansible use; the selected interpreter is what matters.

## Choosing an image

- Targets already have Python 3.9–3.14: reuse that interpreter when evaluating the modern image. **Python 3.12 is not mandatory.**
- RHEL 8 / the Kylin V10 installation above: prepare an additional supported Python before using the modern image, and set `ansible_python_interpreter` to its actual path. Keep the system Python.
- Prepare the interpreter before Python-dependent prechecks and hooks. Offline installations also need OS- and architecture-matched packages with all dependencies; replacing the control image alone is insufficient.

References: [official Ansible support matrix](https://docs.ansible.com/projects/ansible-core/devel/reference_appendices/release_and_maintenance.html#ansible-core-support-matrix), [pinned Kubespray requirements](https://github.com/kubernetes-sigs/kubespray/blob/b33bb4786450395df3c499faa2b7880d4130c693/requirements.txt).
