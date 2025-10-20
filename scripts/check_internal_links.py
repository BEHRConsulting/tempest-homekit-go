#!/usr/bin/env python3
"""Check internal Markdown links across the repository.

Usage: python3 scripts/check_internal_links.py

It scans all .md files (excluding .git and node_modules) and validates
that relative links point to existing files. It accepts links that
point to files directly, to paths that resolve to a file when adding
`.md`, or to directories that contain `README.md`.
"""
import os
import re
import sys

ROOT='.'
skip_dirs = {'.git','node_modules','db'}
link_re = re.compile(r"\[[^\]]*\]\(([^)]+)\)")

missing = []
checked = 0

for dirpath, dirs, files in os.walk(ROOT):
    # filter out skip dirs in-place to avoid walking them
    dirs[:] = [d for d in dirs if d not in skip_dirs]
    for fn in files:
        if not fn.endswith('.md'):
            continue
        fp = os.path.join(dirpath, fn)
        with open(fp, 'r', encoding='utf-8') as f:
            txt = f.read()
        for m in link_re.finditer(txt):
            target = m.group(1).strip()
            # skip http(s), mailto, anchors and images
            if target.startswith('http://') or target.startswith('https://') or target.startswith('mailto:') or target.startswith('javascript:'):
                continue
            if target.startswith('!'):
                continue
            # split off fragment
            path = target.split('#',1)[0]
            if path == '' or path.startswith('#'):
                # intra-document anchor
                continue
            # ignore absolute links starting with /
            if path.startswith('/'):
                # try to resolve from repo root
                candidate = os.path.normpath(os.path.join(ROOT, path.lstrip('/')))
            else:
                candidate = os.path.normpath(os.path.join(dirpath, path))

            exists = False
            # direct file
            if os.path.isfile(candidate):
                exists = True
            # with .md appended
            elif os.path.isfile(candidate + '.md'):
                exists = True
            # directory with README.md
            elif os.path.isdir(candidate) and os.path.isfile(os.path.join(candidate, 'README.md')):
                exists = True

            checked += 1
            if not exists:
                missing.append((fp, target, candidate))

if not missing:
    print(f"Checked {checked} links: no missing internal markdown links found.")
    sys.exit(0)

print(f"Checked {checked} links: {len(missing)} missing internal links:\n")
for src, target, cand in missing:
    print(f"In {src}: -> {target}  (resolved: {cand})")

sys.exit(2)
