import os
import re
import subprocess

def search_files(dir_name, pattern):
    """
    Search for files in a directory (and sub-directories) based on a specific pattern.
    """
    for root, dirs, files in os.walk(dir_name):
        for filename in files:
            if re.match(pattern, filename):
                yield os.path.join(root, filename)

def replace_error_handling(file_path):
    """
    Replace error handling code in a file with eris error handling code.
    """
    with open(file_path, 'r') as f:
        content = f.read()

    # Replace error handling code using regex
    old_pattern = r'fmt\.Errorf\(\"(.*)%w\",\s+err\)'
    new_pattern = r'eris.Wrap(err, "\1")'
    new_content = re.sub(old_pattern, new_pattern, content, flags=re.MULTILINE)

    with open(file_path, 'w') as f:
        f.write(new_content)

    print(f"Replaced error handling code in {file_path}")

def main():
    """
    Main function to search for files, replace error handling code, and commit changes to git.
    """
    # Path to directory containing code files to refactor
    dir_name = "./"

    # Regex pattern to match specific file extensions
    pattern = r'^.*\.(go)$'

    # Search for files based on pattern
    files = search_files(dir_name, pattern)

    # Refactor error handling code in each file
    for file_path in files:
        replace_error_handling(file_path)

    # Commit changes to git
    #subprocess.call(["git", "add", "."])
    #subprocess.call(["git", "commit", "-m", "Refactor error handling using eris"])
    #subprocess.call(["git", "push", "origin", "master"])

if __name__ == '__main__':
    main()

