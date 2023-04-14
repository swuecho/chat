from unidiff import PatchSet
from pathlib import Path
current_dir = Path(__file__).parent

data = (current_dir/  'tools/stream.diff').read_text()
patch = PatchSet(data)
print(len(patch))
for i in patch:
    print(i)