import difflib
import os
from pathlib import Path

def apply_diff_file(diff_file):
    with open(diff_file, 'r') as f:
        diff_text = f.read()

    # Split the diff into files, and parse the file headers
    file_diffs = diff_text.split('diff ')
    print(file_diffs)
    for file_diff in file_diffs[1:]:
        # Parse the file header to get the file names
        file_header, diff = file_diff.split('\n', 1)
        old_file, new_file = file_header.split(' ')[-2:]

        # Apply the diff to the old file
        with open(old_file, 'r') as f:
            old_text = f.read()
        patched_lines = difflib.unified_diff(old_text.splitlines(), diff.splitlines(), lineterm='', fromfile=old_file, tofile=new_file)
        patched_text = os.linesep.join(list(patched_lines)[2:])  # Skip the first two lines of the unified diff

        with open(old_file, 'w') as f:
            f.write(patched_text)

current_dir = Path(__file__).parent

apply_diff_file(current_dir/  'tools/stream.diff')