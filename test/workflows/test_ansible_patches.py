"""Run with: python3 -m unittest discover -s test/workflows -v."""
from contextlib import redirect_stdout
import importlib.util
import io
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import types
import unittest
from unittest import mock

try:
    import yaml
except ImportError:  # pragma: no cover - PyYAML is installed in CI
    yaml = None

ROOT = Path(__file__).resolve().parents[2]
ANSIBLE_PATCHS = ROOT / "build/images/ansible_patchs"
DOCKERFILE = ROOT / "build/images/kubespray/Dockerfile"
ENABLE_REPO = ROOT / "playbooks/enable-repo.yml"
PATCH_DIST_PATH = ANSIBLE_PATCHS / "patch_distribution.py"
PKG_MGR_216_PATCH = ANSIBLE_PATCHS / "pkg_mgr_2.16.patch"


def load_patch_distribution():
    spec = importlib.util.spec_from_file_location("patch_distribution", PATCH_DIST_PATH)
    module = importlib.util.module_from_spec(spec)
    prev = sys.dont_write_bytecode
    sys.dont_write_bytecode = True
    try:
        spec.loader.exec_module(module)
    finally:
        sys.dont_write_bytecode = prev
    return module


def make_distribution_source(redhat_entries: str, prefix: str = "") -> str:
    return (
        f"{prefix}"
        "class Distribution(object):\n"
        f"    OS_FAMILY_MAP = {{'RedHat': [{redhat_entries}],\n"
        "                     'Debian': ['Debian', 'Ubuntu'],\n"
        "                     'Suse': ['SLES']}\n"
        "    OS_FAMILY = {}\n"
        "    for family, names in OS_FAMILY_MAP.items():\n"
        "        for name in names:\n"
        "            OS_FAMILY[name] = family\n"
    )


def exec_distribution(source: str):
    ns = {}
    exec(compile(source, "<distribution>", "exec"), ns)
    return ns["Distribution"]


# Matches lines 84-110 of ansible-core 2.16.20 lib/ansible/module_utils/facts/system/pkg_mgr.py.
PKG_MGR_216_SNIPPET = """import os

PKG_MGRS = [
    {'path': '/usr/bin/yum', 'name': 'yum'},
    {'path': '/usr/bin/dnf-3', 'name': 'dnf'},
    {'path': '/usr/bin/dnf5', 'name': 'dnf5'},
]


class BaseFactCollector:
    pass


class PkgMgrFactCollector(BaseFactCollector):
    def __init__(self, *args, **kwargs):
        super(PkgMgrFactCollector, self).__init__(*args, **kwargs)
        self._default_unknown_pkg_mgr = 'unknown'

    def _check_rh_versions(self, pkg_mgr_name, collected_facts):
        if os.path.exists('/run/ostree-booted'):
            return "atomic_container"

        pkg_mgr_name = self._default_unknown_pkg_mgr
        for bin_path in ('/usr/bin/dnf', '/usr/bin/microdnf'):
            if os.path.exists(bin_path):
                pkg_mgr_name = 'dnf5' if os.path.realpath(bin_path) == '/usr/bin/dnf5' else 'dnf'
                break

        try:
            major_version = collected_facts['ansible_distribution_major_version']
            if collected_facts['ansible_distribution'] == 'Kylin Linux Advanced Server':
                major_version = major_version.lstrip('V')
            distro_major_ver = int(major_version)
        except ValueError:
            return self._default_unknown_pkg_mgr

        if (
            (collected_facts['ansible_distribution'] == 'Fedora' and distro_major_ver < 23)
            or (collected_facts['ansible_distribution'] == 'Kylin Linux Advanced Server' and distro_major_ver < 10)
            or (collected_facts['ansible_distribution'] == 'Amazon' and distro_major_ver < 2022)
            or (collected_facts['ansible_distribution'] == 'TencentOS' and distro_major_ver < 3)
            or distro_major_ver < 8  # assume RHEL or a clone
        ) and any(pm for pm in PKG_MGRS if pm['name'] == 'yum' and os.path.exists(pm['path'])):
            pkg_mgr_name = 'yum'

        return pkg_mgr_name

    def collect(self, collected_facts=None):
        collected_facts = collected_facts or {}
        pkg_mgr_name = self._default_unknown_pkg_mgr
        for pkg in PKG_MGRS:
            if os.path.exists(pkg['path']):
                pkg_mgr_name = pkg['name']
        if collected_facts.get('ansible_os_family') == 'RedHat':
            pkg_mgr_name = self._check_rh_versions(pkg_mgr_name, collected_facts)
        return {'pkg_mgr': pkg_mgr_name}
"""


class PatchDistributionTest(unittest.TestCase):
    def setUp(self):
        self.mod = load_patch_distribution()

    def test_patches_2_16_2_21_and_2_22_layouts(self):
        layouts = {
            "2.16": "'RedHat', 'Alibaba', 'TencentOS', 'Kylin Linux Advanced Server'",
            "2.21": "'RedHat', 'Alibaba', 'TencentOS', 'Kylin Linux Advanced Server', 'MIRACLE'",
            "2.22": "'RedHat', 'Alibaba', 'TencentOS', 'Kylin Linux Advanced Server', 'MIRACLE', 'UnionTech'",
        }
        for label, entries in layouts.items():
            with self.subTest(version=label):
                source = make_distribution_source(entries, prefix="# utf8 comment: 中文注释\n")
                updated, changed = self.mod.patch_source(source)
                self.assertTrue(changed)
                distro_cls = exec_distribution(updated)
                for distro in ("Anolis", "Anolis OS", "Alibaba Cloud Linux", "RedHat", "Kylin Linux Advanced Server"):
                    self.assertEqual(distro_cls.OS_FAMILY.get(distro), "RedHat")
                self.assertEqual(distro_cls.OS_FAMILY.get("Ubuntu"), "Debian")
                self.assertEqual(distro_cls.OS_FAMILY.get("SLES"), "Suse")

                # Re-running must be a clean no-op.
                updated_again, changed_again = self.mod.patch_source(updated)
                self.assertFalse(changed_again)
                self.assertEqual(updated_again, updated)

    def test_devel_already_present_is_noop(self):
        source = make_distribution_source(
            "'RedHat', 'UnionTech', 'Anolis', 'Anolis OS', 'Alibaba Cloud Linux'"
        )
        updated, changed = self.mod.patch_source(source)
        self.assertFalse(changed)
        self.assertEqual(updated, source)

    def test_partial_entries_only_adds_missing(self):
        source = make_distribution_source("'RedHat', 'Anolis',")
        updated, changed = self.mod.patch_source(source)
        self.assertTrue(changed)
        distro_cls = exec_distribution(updated)
        self.assertEqual(distro_cls.OS_FAMILY_MAP["RedHat"].count("Anolis"), 1)
        self.assertEqual(distro_cls.OS_FAMILY.get("Anolis OS"), "RedHat")
        self.assertEqual(distro_cls.OS_FAMILY.get("Alibaba Cloud Linux"), "RedHat")

    def test_missing_os_family_map_raises(self):
        with self.assertRaises(RuntimeError):
            self.mod.patch_source("class Other: pass\n")

    def test_cli_main_updates_file(self):
        source = make_distribution_source("'RedHat', 'Alibaba'")
        with tempfile.TemporaryDirectory() as temp:
            target = Path(temp) / "distribution.py"
            target.write_text(source, encoding="utf-8")
            with redirect_stdout(io.StringIO()):
                self.assertEqual(self.mod.main(["patch_distribution.py", str(target)]), 0)
            distro_cls = exec_distribution(target.read_text(encoding="utf-8"))
            self.assertEqual(distro_cls.OS_FAMILY.get("Alibaba Cloud Linux"), "RedHat")
            with redirect_stdout(io.StringIO()):
                self.assertEqual(self.mod.main(["patch_distribution.py", str(target)]), 0)


class PkgMgr216PatchTest(unittest.TestCase):
    @unittest.skipUnless(shutil.which("patch"), "patch binary not installed")
    def test_pkg_mgr_2_16_patch_behavior(self):
        with tempfile.TemporaryDirectory() as temp:
            target = Path(temp) / "pkg_mgr.py"
            # Pad lines so hunk line offset stays small even under strict patch.
            padded = ("\n" * 60) + PKG_MGR_216_SNIPPET
            target.write_text(padded, encoding="utf-8")
            res = subprocess.run(
                ["patch", str(target), str(PKG_MGR_216_PATCH)],
                capture_output=True,
                text=True,
            )
            self.assertEqual(res.returncode, 0, res.stderr + res.stdout)
            ns = {}
            exec(compile(target.read_text(encoding="utf-8"), str(target), "exec"), ns)
            collector = ns["PkgMgrFactCollector"]()

        dnf_paths = {"/usr/bin/yum", "/usr/bin/dnf", "/usr/bin/dnf-3"}
        dnf5_paths = {"/usr/bin/yum", "/usr/bin/dnf", "/usr/bin/dnf5"}
        yum_only_paths = {"/usr/bin/yum"}

        cases = [
            ("Alibaba Cloud Linux", "4", dnf_paths, {}, "dnf"),
            ("Alibaba Cloud Linux", "4", dnf5_paths, {"/usr/bin/dnf": "/usr/bin/dnf5"}, "dnf5"),
            ("Alibaba", "3", dnf_paths, {}, "dnf"),
            ("Alibaba", "2", yum_only_paths, {}, "yum"),
            ("TencentOS", "3", dnf_paths, {}, "dnf"),
            ("TencentOS", "2", yum_only_paths, {}, "yum"),
            ("Anolis", "8", dnf_paths, {}, "dnf"),
            ("Anolis OS", "23", dnf_paths, {}, "dnf"),
            ("Kylin Linux Advanced Server", "V10", dnf_paths, {}, "dnf"),
            ("RedHat", "7", yum_only_paths, {}, "yum"),
            ("RedHat", "8", dnf_paths, {}, "dnf"),
        ]
        for distro, major, paths, realpaths, expected in cases:
            with self.subTest(distro=distro, major=major, expected=expected):
                facts = {
                    "ansible_distribution": distro,
                    "ansible_distribution_major_version": major,
                    "ansible_os_family": "RedHat",
                }
                with mock.patch("os.path.exists", side_effect=lambda p, ps=paths: p in ps), \
                     mock.patch("os.path.realpath", side_effect=lambda p, rp=realpaths: rp.get(p, p)):
                    result = collector.collect(collected_facts=facts)
                self.assertEqual(result["pkg_mgr"], expected)


class DockerfileAndPlaybookWiringTest(unittest.TestCase):
    def test_kubespray_dockerfile_wires_and_cleans_patches(self):
        text = DOCKERFILE.read_text(encoding="utf-8")
        self.assertIn("pkg_mgr_2.16.patch", text)
        self.assertIn("patch_distribution.py", text)
        self.assertIn("rm -rf", text)
        self.assertIn("/tmp/patch_distribution.py", text)
        self.assertIn("Alibaba Cloud Linux", text)

    def test_enable_repo_includes_anolis_and_alibaba_cloud_linux(self):
        expected = {"Alibaba Cloud Linux", "Anolis", "Anolis OS"}
        if yaml is not None:
            plays = yaml.safe_load(ENABLE_REPO.read_text(encoding="utf-8"))
            extensions = set(plays[0]["vars"]["redhat_os_family_extensions"])
            self.assertTrue(expected.issubset(extensions), extensions)
        else:
            text = ENABLE_REPO.read_text(encoding="utf-8")
            for name in expected:
                self.assertIn(f'"{name}"', text)


if __name__ == "__main__":
    unittest.main()
