#!/usr/bin/env python

import subprocess
import yaml
import os
import re
from jinja2 import Template
from datetime import datetime

IMAGE_REPO = os.getenv("IMAGE_REPO", default="kubean-io")

KUBEAN_PATCH_TEMPLATE = '''
{% for release, infos in releases.items() %}
### ▶️ release-{{ release }}
> ⚓ kube_version range: [ {{ infos[0].kube_version_range[infos[0].kube_version_range|length-1] }} ~ {{ infos[0].kube_version_range[0] }} ]

<table>
  <thead>
    <tr>
      <th>Commit Date</th>
      <th>Artifacts</th>
    </tr>
  </thead>
  {% for info in infos %}
  <tbody>
      <tr>
        <td rowspan=3> 📅 {{ info.commit_date }} </td>
        <td rowspan=1>
           📝 <code><a href="https://raw.githubusercontent.com/{{ repo_name }}/kubean-manifest/main/manifests/manifest-{{ release }}-{{ info.commit_short_sha }}.yml">manifest-{{ release }}-{{ info.commit_short_sha }}.yml</a></code>
        </td>
      </tr>
      <tr>
        <td rowspan=1> 📦 <code>{{ image_registry }}/{{ repo_name }}/spray-job:{{ release }}-{{ info.commit_short_sha }}</code> </td>
      </tr>
      <tr>
        <td rowspan=1> 📦 <code>{{ image_registry }}/{{ repo_name }}/airgap-patch:{{ release }}-{{ info.commit_short_sha }}</code> </td>
      </tr>
  </tbody>
  {% endfor %}
</table>
{% endfor %}
'''


if __name__ == '__main__':

  repo_name="kubean-manifest"
  subprocess.getoutput(f"git clone https://github.com/{IMAGE_REPO}/{repo_name}.git")
  manifests_path=f'{repo_name}/manifests'
  rls_pattern = re.compile(r'manifest-(\d+\.\d+)-.*\.yml')

  releases = {}
  for manifest in os.listdir(manifests_path):
    match = rls_pattern.match(manifest)
    if match:
      rls_version = match.group(1)
      releases[rls_version] = []
      with open(f'{manifests_path}/{manifest}', 'r') as stream:
        data = yaml.safe_load(stream)
        commit_short_sha = data.get('metadata', {}).get('annotations', {}).get('kubean.io/sprayCommit')
        commit_timestamp = int(data.get('metadata', {}).get('annotations', {}).get('kubean.io/sprayTimestamp'))
        commit_date = datetime.fromtimestamp(commit_timestamp)
        components = data.get('spec', {}).get('components', {})
        kube_info = next(item for item in data.get('spec', {}).get('components', {}) if item['name'] == 'kube')
        kube_version_range = kube_info.get('versionRange', [])
        sorted_versions = sorted(kube_version_range, key=lambda x: [int(x) if x.isdigit() else x for x in re.split('([0-9]+)', x)], reverse=True)
        releases[rls_version].append({
          'commit_short_sha': commit_short_sha,
          'commit_timestamp': commit_timestamp,
          'commit_date': commit_date,
          'kube_version_range': sorted_versions})

  for key in releases.keys():
    releases[key].sort(key=lambda item:item['commit_timestamp'], reverse=True)

  # print(f'releases: {releases}')

  t = Template(KUBEAN_PATCH_TEMPLATE)
  md_contents = t.render(releases=releases, image_registry='ghcr.io', repo_name=IMAGE_REPO)
  # print(md_contents)

  artifact_md = open('artifacts.md', 'w')
  artifact_md.write(md_contents)
  artifact_md.close()

  subprocess.getoutput(f"rm -rf {repo_name}")
