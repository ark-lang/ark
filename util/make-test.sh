#!/bin/bash

# no arguments no test file
if [[ $# -eq 0 ]] ; then
	echo 'Expected an Ark source file'
	echo 'usage:'
	echo '   ./util/make-test.sh tests/some-test.ark'
	exit 0
fi

# just the sourcefile with the extension
source_file=$(basename "$1")

# sourcefile with no extension
source_filename="${source_file%.*}"

# remove the extension, put a .toml on the end
# note the extension is .ark which is 4 chars long
# kind of hacky but it works
source_no_ext=$(echo "$1"|sed 's/.\{4\}$//').toml

echo "- Creating test template for '$source_filename'"
echo 'Name       = "'$source_filename'"
Sourcefile = "'$source_file'"

CompilerArgs = []
RunArgs      = []

CompilerError = 0
RunError      = 0

Input = ""

CompilerOutput = ""
RunOutput      = ""' > $source_no_ext
echo "- Complete"