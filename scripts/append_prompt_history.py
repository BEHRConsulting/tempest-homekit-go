#!/usr/bin/env python3
"""Append a timestamped paraphrased prompt to PROMPT_HISTORY.md

Usage:
  python3 scripts/append_prompt_history.py "Paraphrased prompt text" "Short outcome"

This script will append a new bullet with the current ISO timestamp.
"""
import sys
import datetime
from pathlib import Path

if len(sys.argv) < 2:
    print("Usage: append_prompt_history.py \"paraphrase\" [outcome]")
    sys.exit(1)

paraphrase = sys.argv[1].strip()
outcome = sys.argv[2].strip() if len(sys.argv) > 2 else ''
filep = Path('PROMPT_HISTORY.md')
if not filep.exists():
    print('PROMPT_HISTORY.md not found in repo root')
    sys.exit(1)

now = datetime.datetime.utcnow().replace(microsecond=0).isoformat() + 'Z'
entry_lines = []
entry_lines.append(f"- [{now}] \"{paraphrase}\"")
if outcome:
    entry_lines[-1] += f" — Outcome: {outcome}"

with filep.open('a', encoding='utf-8') as f:
    f.write('\n')
    f.write('\n'.join(entry_lines))
    f.write('\n')

print('Appended prompt to PROMPT_HISTORY.md')
