#!/bin/bash
TOTAL=$(find parser-tests/*.ay -type f | wc -l)
OUTPUT="$(ls)"
NUM=0

for filename in parser-tests/*.ay; do
	OUTPUT="$(alloyc $filename)"
	((NUM++))
	if [[ $OUTPUT != *"Finished"* ]]; then
		echo "Failed to parse '$filename', ($NUM/$TOTAL)"
		echo "$OUTPUT"
		exit 1
	fi
done

echo "Successfully parsed $NUM/$TOTAL file/s"
