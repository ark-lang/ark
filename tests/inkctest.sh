#!/bin/bash
TOTAL=$(find parser-tests/*.ink -type f | wc -l)
OUTPUT="$(ls)"
NUM=0

for filename in parser-tests/*.ink; do
	OUTPUT="$(inkc $filename)"
	((NUM++))
	if [[ $OUTPUT != "Finished"* ]]; then
		echo "Failed to parse $filename, ($NUM/$TOTAL)"
		set -e
	fi
done

echo "Successfully parsed $NUM/$TOTAL file/s"