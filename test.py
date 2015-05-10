#!/usr/bin/env python3
# Simple test script
# Goes through all files in the test/ directory
# and runs all of the files matching this: *_test.aly

import os, subprocess

def bold(s):
	return "\033[1m" + s + "\033[0m"

def red(s):
	return "\x1B[31m" + s + "\x1B[00m"

for name in os.listdir("tests"):
	if name.endswith("_test.aly"):
		output_file = name + ".test"
		print(bold("Compiling ") + name + "...")
		
		compile_result = subprocess.call(["alloyc", "tests/" + name, "-o", "tests/" + output_file])
		
		if compile_result != 0:
			print(red(bold("Compilation failed:")) + " returned with " + str(compile_result))
			exit()
		
		print(bold("Running ") + name + "...")
		
		try:
			run_result = subprocess.call(["./tests/" + output_file])
		except FileNotFoundError:
			print(red(bold("Something went wrong... file not found")))
			exit()
		
		os.remove("tests/" + output_file)
		if run_result != 0:
			print(red(bold("Running failed:")) + " returned with " + str(run_result))
			exit()
