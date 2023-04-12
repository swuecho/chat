import argparse

parser = argparse.ArgumentParser()
parser.add_argument('tag')
args = parser.parse_args()
tag = args.tag

with open('docker-compose.yml', 'r') as f:
    lines = f.readlines()

with open('docker-compose.yml', 'w') as f:
    for line in lines:
        if 'latest' in line:
            line = line.replace('latest', tag)
        f.write(line)
