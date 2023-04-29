import json
from pathlib import Path

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

def merge_json_files(file1, file2):
    """Merges the contents of two JSON files recursively by key."""
    
    # Read in the JSON data from the files
    with open(file1, 'r') as f1:
        data1 = json.load(f1)
    with open(file2, 'r') as f2:
        data2 = json.load(f2)

    merge_dicts(data1, data2)
     # write the merged content back to file2
    with open(file1, 'w') as fp1:
        json.dump(data1, fp1, indent=4,ensure_ascii=False, sort_keys=True)


# main
locale_dir = Path(__file__).parent.parent / "web/src/locales"
extra_jsons = locale_dir.glob("*-more.json")
# web/src/locales/en-US.json web/src/locales/en-US-more.json 
for extra in extra_jsons:
    print(extra)
    origin = extra.parent / extra.name.replace('-more', '')
    print(origin, extra)
    merge_json_files(origin, extra)
