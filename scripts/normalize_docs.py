"""Normalize Markdown docs:
- remove emoji characters in common Unicode emoji ranges
- trim trailing whitespace
- create .bak backup before writing

Run from repo root: python3 scripts/normalize_docs.py
"""
import os
import sys

emoji_ranges=[
    (0x1F300,0x1F5FF),
    (0x1F600,0x1F64F),
    (0x1F680,0x1F6FF),
    (0x2600,0x26FF),
    (0x1F900,0x1F9FF),
    (0x1F1E6,0x1F1FF),
    (0x2700,0x27BF),
]

def is_emoji(ch):
    o=ord(ch)
    for a,b in emoji_ranges:
        if a<=o<=b:
            return True
    return False

count=0
modified_files=[]
for dirpath,dirs,files in os.walk('.'):
    if '.git' in dirpath or 'node_modules' in dirpath:
        continue
    for fn in files:
        if not fn.endswith('.md'):
            continue
        fp=os.path.join(dirpath,fn)
        with open(fp,'r',encoding='utf-8') as f:
            txt=f.read()
        new_chars=[]
        changed=False
        for ch in txt:
            if is_emoji(ch):
                # remove emoji
                changed=True
                count+=1
            else:
                new_chars.append(ch)
        new=''.join(new_chars)
        # trim trailing whitespace on each line
        lines=new.splitlines()
        new2='\n'.join([ln.rstrip() for ln in lines])
        # preserve trailing newline if original had
        if txt.endswith('\n') and not new2.endswith('\n'):
            new2 += '\n'
        if new2!=txt:
            # backup
            bak=fp+'.bak'
            open(bak,'w',encoding='utf-8').write(txt)
            open(fp,'w',encoding='utf-8').write(new2)
            modified_files.append(fp)

print(f"Processed {len(modified_files)} files, removed {count} emoji characters.")
if modified_files:
    print('Modified files (first 200):')
    for p in modified_files[:200]:
        print(p)

sys.exit(0)
