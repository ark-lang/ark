import os

template = \
"""Name       = "{}"
Sourcefile = "{}"

CompilerArgs = []
RunArgs      = []

CompilerError = 0
RunError      = 0

Input = ""

CompilerOutput = ""
RunOutput      = ""
"""

for f in os.listdir("."):
	if not f.endswith("_test.ark"):
		continue

	test_name = f.replace("_test.ark", "")
	toml_name = test_name + ".toml"
	ark_name = test_name + ".ark"
	print(f, "=>", test_name, toml_name, ark_name)
	print("====================================================================")
	print(template.format(test_name, ark_name))

	with open(toml_name, "w") as tf:
		tf.write(template.format(test_name, ark_name))

	os.rename(f, ark_name)