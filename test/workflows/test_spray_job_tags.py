"""Run with: python3 -m unittest discover -s test/workflows -v (requires PyYAML)."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml

ROOT = Path(__file__).resolve().parents[2]
WORKFLOWS = ROOT / ".github/workflows"
KUBEAN_WORKFLOW = WORKFLOWS / "call-build-imgs-for-kubean.yaml"
PUBLISH = ROOT / "hack/publish-image-tags.sh"
IMAGE = "ghcr.io/example/spray-job"
DIGEST = "sha256:" + "a" * 64
DOCKER = r'''#!/usr/bin/env python3
import json
import os
import pathlib
import sys

args = sys.argv[1:]
with open(os.environ["DOCKER_LOG"], "a") as log:
    log.write(json.dumps(args) + "\n")
if args[0] == "create":
    if os.environ["EXPECTED_IMAGE_REF"] not in args:
        sys.exit(90)  # Never inspect a mutable tag or a base image.
    print("fake-container")
elif args[0] == "cp":
    if os.environ["CP_FAILS"]:
        sys.exit(91)
    pathlib.Path(args[2]).write_text(os.environ["CORE"] + "\n")
elif args[0] == "rm":
    pass
elif args[:3] == ["buildx", "imagetools", "create"]:
    pass
else:
    sys.exit(92)  # Reading the version must never execute code from the image.
'''


def workflow(path):
    return yaml.safe_load(path.read_text())


def strip_comments(text):
    return "\n".join(line for line in text.splitlines() if not line.strip().startswith("#"))


def run_steps(path):
    for job in workflow(path)["jobs"].values():
        for step in job.get("steps") or []:
            if "run" in step:
                yield strip_comments(step["run"])


def tag_fields(node):
    """Every field that can carry an image reference, at any depth."""
    if isinstance(node, dict):
        for key, value in node.items():
            if key in ("tags", "aliases", "build-args") and isinstance(value, str):
                yield value
            else:
                yield from tag_fields(value)
    elif isinstance(node, list):
        for item in node:
            yield from tag_fields(item)


class PublishImageTagsTest(unittest.TestCase):
    def publish(self, core="2.21", digest=DIGEST, cp_fails=False, aliases=("v1-modern", "latest")):
        with tempfile.TemporaryDirectory() as temp:
            docker = Path(temp) / "docker"
            docker.write_text(DOCKER)
            docker.chmod(0o755)
            log = Path(temp) / "docker.log"
            env = dict(os.environ, PATH=f"{temp}:{os.environ['PATH']}",
                       EXPECTED_IMAGE_REF=f"{IMAGE}@{digest}", CORE=core,
                       CP_FAILS="1" if cp_fails else "", DOCKER_LOG=str(log))
            result = subprocess.run([str(PUBLISH), IMAGE, digest, "v1", *aliases],
                                    cwd=ROOT, env=env, capture_output=True, text=True)
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
        return result, calls

    @staticmethod
    def publications(calls):
        return [call for call in calls if call[:3] == ["buildx", "imagetools", "create"]]

    def test_tags_come_from_final_digest(self):
        for core, aliases in (("2.16", ("v1",)), ("2.21", ("v1-modern", "latest")), ("2.23", ("v1-modern",))):
            with self.subTest(core=core):
                result, calls = self.publish(core=core, aliases=aliases)
                self.assertEqual(result.returncode, 0, result.stderr)
                publications = self.publications(calls)
                self.assertEqual(len(publications), 1)
                call = publications[0]
                self.assertEqual(call[-1], f"{IMAGE}@{DIGEST}")
                tags = [call[i + 1] for i, arg in enumerate(call) if arg == "--tag"]
                self.assertEqual(tags, [f"{IMAGE}:v1-ansible-core{core}",
                                        *(f"{IMAGE}:{alias}" for alias in aliases)])

    def test_unreadable_image_does_not_publish(self):
        result, calls = self.publish(cp_fails=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(self.publications(calls))

    def test_unexpected_core_version_does_not_publish(self):
        for core in ("", "2", "not-a-version"):
            with self.subTest(core=core):
                result, calls = self.publish(core=core)
                self.assertNotEqual(result.returncode, 0)
                self.assertFalse(self.publications(calls))

    def test_missing_digest_fails_before_touching_docker(self):
        result, calls = self.publish(digest="")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(calls, [])


class WorkflowTagLayoutTest(unittest.TestCase):
    def test_only_spray_job_entries_are_pushed_by_digest(self):
        entries = workflow(KUBEAN_WORKFLOW)["jobs"]["build-imgs"]["strategy"]["matrix"]["include"]
        by_digest = {entry["name"] for entry in entries if entry.get("outputs")}
        self.assertEqual(by_digest, {"spray-job", "spray-job-ansible9"})
        for entry in entries:
            with self.subTest(name=entry["name"]):
                # 按 digest 推送的条目不能再填 tags，否则 build-push-action 会推出未经检查的 tag。
                self.assertEqual(bool(entry.get("tags")), entry["name"] not in by_digest)
                self.assertEqual(bool(entry.get("aliases")), entry["name"] in by_digest)

    def test_no_workflow_assembles_core_tags_inline(self):
        for path in sorted(WORKFLOWS.glob("*.y*ml")):
            for run in run_steps(path):
                if "-ansible-core" in run:
                    with self.subTest(workflow=path.name):
                        self.assertIn("publish-image-tags.sh", run)

    def test_no_workflow_still_refers_to_the_ansible9_tag(self):
        # 基座裸 tag 就是 ansible9 线，现代线必须显式写 -modern，避免两种语义再次混淆。
        # matrix 条目名里的 ansible9 描述的是 FIX_ANSIBLE9 构建变体，不是 tag，不在此列。
        for path in sorted(WORKFLOWS.glob("*.y*ml")):
            with self.subTest(workflow=path.name):
                for value in tag_fields(workflow(path)):
                    self.assertNotIn("-ansible9", value)
                for run in run_steps(path):
                    self.assertNotIn("-ansible9", run)


if __name__ == "__main__":
    unittest.main()
