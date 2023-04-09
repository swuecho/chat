import argparse
import json

def merge_json_files(file1, file2):
    """Merges the contents of two JSON files recursively by key."""
    
    # Read in the JSON data from the files
    with open(file1, 'r') as f1:
        data1 = json.load(f1)
    with open(file2, 'r') as f2:
        data2 = json.load(f2)

    # Function to recursively merge two dictionaries
    def merge_dicts(d1, d2):
        for key, val2 in d2.items():
            if key in d1:
                # If both values are dictionaries, merge them recursively
                if isinstance(val2, dict) and isinstance(d1[key], dict):
                    merge_dicts(d1[key], val2)
                # If both values are lists, extend the first list with the second
                elif isinstance(val2, list) and isinstance(d1[key], list):
                    d1[key].extend(val2)
                # Otherwise, overwrite the first value with the second
                else:
                    d1[key] = val2
            else:
                # If the key doesn't exist in the first dict, add it and its value
                d1[key] = val2

    # Merge the second data into the first data
    merge_dicts(data1, data2)

    return data1

# Define a command line parser and arguments
parser = argparse.ArgumentParser(description='Merge two JSON files recursively by key.')
parser.add_argument('file1', type=str, help='The filename of the first JSON file to merge.')
parser.add_argument('file2', type=str, help='The filename of the second JSON file to merge.')
args = parser.parse_args()

# Merge the JSON data from the files
merged_data = merge_json_files(args.file1, args.file2)

# Print the merged JSON data, formatted for readability
print(json.dumps(merged_data, indent=4, ensure_ascii=False, sort_keys=True))

# python json_merge.py A.json A-more.json