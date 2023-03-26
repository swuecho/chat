from pathlib import Path
import json
import argparse

# Recursive function to find missing keys in dictionaries
def find_missing_keys(base_dict, other_dict):
    missing_keys = {}
    for key in base_dict:
        if key not in other_dict:
            missing_keys[key] = base_dict[key]
        elif isinstance(base_dict[key], dict) and isinstance(other_dict[key], dict):
            sub_missing_keys = find_missing_keys(base_dict[key], other_dict[key])
            if sub_missing_keys:
                missing_keys[key] = sub_missing_keys
    return missing_keys



def check_locales(dir_name: str, base_locale: str = 'zh-CN'):
    # Load the zh-CN JSON file
    zh_cn_file = Path(dir_name) / f'{base_locale}.json'
    with zh_cn_file.open('r') as f:
        zh_cn = json.load(f)

    # Look for other JSON files in the current directory
    for file in Path(dir_name).glob('*.json'):
        cur_locale = file.stem
        if cur_locale != base_locale:
            with file.open('r') as f:
                other_dict = json.load(f)
            missing_keys = find_missing_keys(zh_cn, other_dict)
            print(f'\n\n please translate to {cur_locale}:', missing_keys)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Check missing keys in language localization files')
    parser.add_argument('dir_name', type=str, help='directory where the JSON files are located')
    parser.add_argument('--base', type=str, default='zh-CN', help='base locale to compare against')
    args = parser.parse_args()
    check_locales(args.dir_name, args.base)
    # python check_locales.py /path/to/locales --base en-US
