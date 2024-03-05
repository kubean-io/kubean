
import time
from jinja2 import Template

# Copyright 2023 Authors of kubean-io
# SPDX-License-Identifier: Apache-2.0

class CR_Template:
  CR_LOCALARTIFACTSET_TEMPLATE = '''
apiVersion: kubean.io/v1alpha1
kind: LocalArtifactSet
metadata:
{%- if sprayInfo.sprayRlease != "master" and sprayInfo.sprayRlease != "" %}
  name:
  {%- if not isPatch -%}
    {{ " " }}"localartifactset-{{ sprayInfo.sprayRlease }}-{{ sprayInfo.sprayCommitShort }}"
  {%- else -%}
    {{ " " }}"localartifactset-{{ sprayInfo.sprayRlease }}-{{ sprayInfo.sprayCommitShort }}-{{ currentTime }}"
  {%- endif %}
  labels:
    kubean.io/sprayRelease: "{{ sprayInfo.sprayRlease }}"
  annotations:
    kubean.io/sprayTimestamp: "{{ sprayInfo.sprayCommitTimestamp }}"
    kubean.io/sprayRelease: "{{ sprayInfo.sprayRlease }}"
    kubean.io/sprayCommit: "{{ sprayInfo.sprayCommitShort }}"
{%- else %}
  name:
  {%- if not isPatch -%}
    {{ " " }}"localartifactset-{{ currentTime }}"
  {%- else -%}
    {{ " " }}"localartifactset-patch-{{ currentTime }}"
  {%- endif %}
  labels:
    kubean.io/sprayRelease: master
{%- endif %}
spec:
  kubespray: "{{ sprayInfo.sprayCommit }}"
  items:
  {%- for name, versions in components.items() %}
    - name: {{ name }}
      versionRange:
      {%- for version in versions %}
        - "{{ version }}"
      {%- endfor %}
  {%- endfor %}

'''

  CR_MANIFEST_TEMPLATE = '''
apiVersion: kubean.io/v1alpha1
kind: Manifest
metadata:
{%- if sprayInfo.sprayRlease != "master" and sprayInfo.sprayRlease != "" %}
  name: "manifest-{{ sprayInfo.sprayRlease }}-{{ sprayInfo.sprayCommitShort }}"
  labels:
    kubean.io/sprayRelease: "{{ sprayInfo.sprayRlease }}"
  annotations:
    kubean.io/sprayTimestamp: "{{ sprayInfo.sprayCommitTimestamp }}"
    kubean.io/sprayRelease: "{{ sprayInfo.sprayRlease }}"
    kubean.io/sprayCommit: "{{ sprayInfo.sprayCommitShort }}"
{%- else %}
  name: "manifest-{{ kubeanTag|replace('.', '-') }}"
  labels:
    kubean.io/sprayRelease: master
{%- endif %}
spec:
  kubesprayVersion: "{{ sprayInfo.sprayCommit }}"
  kubeanVersion: "{{ kubeanTag }}"
  docker:
  {%- for name, version in dockers.items() %}
    - os: {{ name }}
      defaultVersion: "{{ version.defaultVersion }}"
      versionRange:
      {%- for versionItem in version.versionRange %}
        - "{{ versionItem }}"
      {%- endfor %}
  {%- endfor %}
  components:
  {%- for name, version in components.items() %}
    - name: {{ name }}
      defaultVersion: "{{ version.defaultVersion }}"
      versionRange:
      {%- if version.versionRange %}
        {%- for versionItem in version.versionRange %}
        - "{{ versionItem }}"
        {%- endfor %}
      {%- else -%}
        {{ ' ' }}[]
      {%- endif %}
  {%- endfor %}

'''

  def __init__(self, kubeanTag: str, sprayInfo: dict, components: dict, dockers: dict, isPatch: bool) -> None:
    self.data = {
      "isPatch": isPatch,
      "kubeanTag": kubeanTag,
      "sprayInfo": sprayInfo,
      "components": components,
      "dockers": dockers,
      "currentTime": int(time.time())
    }
    
  def render_template(self, template_string, output_path):
    template = Template(template_string)
    rendered_output = template.render(self.data)
    with open(output_path, 'w') as output_file:
        output_file.write(rendered_output)
    print(f"The rendered output path is located at: {output_path}\n")
