#!/usr/bin/env python
# Simple test script
# Goes through all files in the test/ directory
# and runs all of the files matching this: *_test.ark

import os, subprocess, sys, re

# function for bolding a string
def bold(s):
	return "\033[1m" + s + "\033[0m"

# function for making a string red
def red(s):
	return "\x1B[31m" + s + "\x1B[00m"

# function for making a string green
def green(s):
	return "\x1B[32m" + s + "\x1B[00m"

# Class representing our test file
# has its name and whether it failed
# or not
class TestFile:
	name = "?"
	failed = True
	result = ""

	def __init__(self, name, failed, result):
		self.name = name
		self.failed = failed
		self.result = result

# all of the files tested go here
files_tested = []

# how many files passed
num_of_files_passed = 0

# how many files failed
num_of_files_failed = 0

# if we show the ouput to the console
# TODO maybe print out summary even if 
# no output is shown?
show_output = False

FNULL = open(os.devnull, 'w')

# This is kind of hacky and not very scalable,
# will show the output if they added -o as an arg
if "-o" in sys.argv or "--show-output" in sys.argv:
	show_output = True;

def sort_nicely(l):
	"""Sort the given list in the way that humans expect."""
	convert = lambda text: int(text) if text.isdigit() else text
	alphanum_key = lambda key: [convert(c) for c in re.split('([0-9]+)', key)]
	l.sort(key = alphanum_key)

files = [x for x in os.listdir("tests") if x.endswith("_test.ark")]
sort_nicely(files)
for name in files:
	output_file = name + ".test"
	
	if show_output: 
		print(bold("Compiling ") + name + "...")
	
	if show_output:
		compile_result = subprocess.call(["arkc", "tests/" + name, "-o", "tests/" + output_file])
	else:
		compile_result = subprocess.call(["arkc", "tests/" + name, "-o", "tests/" + output_file], stdout=FNULL, stderr=subprocess.STDOUT)

	if compile_result != 0:
		if show_output: 
			print(red(bold("Compilation failed:")) + " returned with " + str(compile_result))
		files_tested.append(TestFile(name, True, str(compile_result)))
		num_of_files_failed += 1
		if show_output: print("")
		continue
	
	if show_output: 
		print(bold("Running ") + name + "...")
	
	try:
		if show_output:
			run_result = subprocess.call(["./tests/" + output_file])
		else:
			run_result = subprocess.call(["./tests/" + output_file], stdout=FNULL, stderr=subprocess.STDOUT)
		
		os.remove("tests/" + output_file)
	except FileNotFoundError:
		print(red(bold("File not found: " + output_file)))
	
	if run_result != 0:
		if show_output: print(red(bold("Running failed:")) + " returned with " + str(run_result))
		files_tested.append(TestFile(name, True, str(compile_result)))
		num_of_files_failed += 1
	else:
		files_tested.append(TestFile(name, False, str(compile_result)))
		num_of_files_passed += 1
		
	if show_output: print("")

# print results
total_num_of_files = num_of_files_passed + num_of_files_failed
print(bold("Results: " + str(num_of_files_passed) + "/" + str(total_num_of_files) + " files passed")) # some margin

for file in files_tested:
	if file.failed:
		print(red(bold("    [-](" + file.result + ")\t " + file.name)))
	else:
		print(green(bold("    [+](" + file.result + ")\t " + file.name)))

exit(num_of_files_failed)
