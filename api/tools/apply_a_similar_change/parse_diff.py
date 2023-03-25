import os
import difflib
import shutil
from pathlib import Path


# Initialize variables
diff = []
file_path = ""
new_file_path = ""
patched_files = []

# Load the diff file into a list of lines
with open('/Users/hwu/dev/chat/api/tools/stream.diff', 'r') as f:
    diff_lines = f.readlines()

# Loop through the diff lines
for line in diff_lines:
    print(line)
    if line.startswith('diff --git'):  # Start of a new file
        if diff:
            # Apply the diff to the previous file
            diff_content = ''.join(diff)
            print("x")
            print(file_path, new_file_path)
            print("x")
            patch = difflib.unified_diff(Path(file_path).read_text(), diff_content.splitlines(), fromfile=file_path, tofile=new_file_path)
            patched_file_content = ''.join(patch)
            with open(file_path, 'w') as f:
                f.write(patched_file_content)
                patched_files.append(file_path)

        # Initialize variables for the new file
        diff = []
        _, _, file_path, new_file_path = line.split(' ')
        file_path = file_path[2:]
        new_file_path = new_file_path[2:]

    elif line.startswith('---') or line.startswith('+++'):  # Ignore the old and new file paths
        continue

    elif line.startswith('new file'):  # Handle new files
        if diff:
            # Apply the diff to the previous file
            diff_content = ''.join(diff)
            patch = difflib.unified_diff(shutil.readFile(file_path), diff_content.splitlines(), fromfile=file_path, tofile=new_file_path)
            patched_file_content = ''.join(patch)
            with open(file_path, 'w') as f:
                f.write(patched_file_content)
                patched_files.append(file_path)

        # Initialize variables for the new file
        diff = []
        _, _, new_file_path = line.split(' ')
        new_file_path = new_file_path[2:]
        file_path = new_file_path

    else:
        # Add the line to the diff content
        diff.append(line)

if diff:
    # Apply the diff to the last file
    diff_content = ''.join(diff)
    patch = difflib.unified_diff(Path(file_path).read_text(), diff_content.splitlines(), fromfile=file_path, tofile=new_file_path)
    patched_file_content = ''.join(patch)
    with open(file_path, 'w') as f:
        f.write(patched_file_content)
        patched_files.append(file_path)

# Print the list of patched files
print("Patched files:")
for file in patched_files:
    print(file)
