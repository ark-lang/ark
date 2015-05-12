#!/usr/bin/env python
# Simple test script
# Goes through all files in the test/ directory
# and runs all of the files matching this: *_test.aly

import os, subprocess, sys

class TestFile:
	name = "?"
	failed = True

	def __init__(self, name, failed):
		self.name = name
		self.failed = failed

files_tested = []
num_of_files_passed = 0
num_of_files_failed = 0
show_output = False

if len(sys.argv) > 1 and str(sys.argv[1]) == "-o":
	show_output = True

def bold(s):
	return "\033[1m" + s + "\033[0m"

def red(s):
	return "\x1B[31m" + s + "\x1B[00m"

def green(s):
	return "\x1B[32m" + s + "\x1B[00m"

for name in os.listdir("tests"):
	if name.endswith("_test.aly"):
		output_file = name + ".test"
		
		if show_output: 
			print(bold("Compiling ") + name + "...")
		
		if show_output:
			compile_result = subprocess.call(["alloyc", "tests/" + name, "-o", "tests/" + output_file])
		else:
			FNULL = open(os.devnull, 'w')
			compile_result = subprocess.call(["alloyc", "tests/" + name, "-o", "tests/" + output_file], stdout=FNULL, stderr=subprocess.STDOUT)

		if compile_result != 0:
			if show_output: 
				print(red(bold("Compilation failed:")) + " returned with " + str(compile_result))
			exit()
		
		if show_output: 
			print(bold("Running ") + name + "...")
		
		try:
			if show_output:
				run_result = subprocess.call(["./tests/" + output_file])
			else:
				FNULL = open(os.devnull, 'w')
				run_result = subprocess.call(["./tests/" + output_file], stdout=FNULL, stderr=subprocess.STDOUT)
				
		except FileNotFoundError:
			print(red(bold("File not found: " + output_file)))
			exit()
		
		os.remove("tests/" + output_file)
		if run_result != 0:
			if show_output: print(red(bold("Running failed:")) + " returned with " + str(run_result))
			files_tested.append(TestFile(output_file, True))
			num_of_files_failed += 1
		else:
			files_tested.append(TestFile(output_file, False))
			num_of_files_passed += 1

total_num_of_files = num_of_files_passed + num_of_files_failed
print(bold("Results (" + str(num_of_files_passed) + "/" + str(total_num_of_files) + ") files passed: ")) # some margin

for file in files_tested:
	if file.failed:
		print(red(bold("    [-] " + file.name)))
	else:
		print(green(bold("    [+] " + file.name)))