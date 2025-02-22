#!/bin/bash
# Usage: ./print_files.sh /path/to/directory
# This script prints the full path and contents of every file found
# under the provided directory. Each file's output is separated by clear headers.

if [ -z "$1" ]; then
  echo "Usage: $0 <directory>"
  exit 1
fi

directory="$1"

# Use 'find' to get all files and process each file found.
find "$directory" -type f | while read -r file; do
    echo "==== START FILE: $file ===="
    cat "$file"
    echo "==== END FILE: $file ===="
    echo ""
done
