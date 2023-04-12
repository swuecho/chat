import argparse

parser = argparse.ArgumentParser()
parser.add_argument('tag')
args = parser.parse_args()
tag = args.tag

with open('docker-compose.yaml', 'r') as f:
    lines = f.readlines()

with open('docker-compose.yaml', 'w') as f:
    for line in lines:
        if 'latest' in line:
            line = line.replace('latest', tag)
        f.write(line)
