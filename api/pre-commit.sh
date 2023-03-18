#!/bin/bash

echo "Running 'go fmt' check..."

files=$(git diff --cached --name-only --diff-filter=ACM "*.go")

if [ -z "$files" ]; then
  echo "No Go files to check."
  exit 0
fi

unformatted_files=""
for file in ${files}; do
  if [[ ! -z $(go fmt ${file}) ]]; then
    unformatted_files="$unformatted_files ${file}"
  fi
done

if [ ! -z "$unformatted_files" ]; then
  echo "The following files are not properly formatted:"
  echo "$unformatted_files"
  echo "Please run 'go fmt' before committing."
  exit 1
fi

echo "'go fmt' check passed."
exit 0
