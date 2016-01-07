# Contributing to Ark
First of all, thanks for your interest in contributing to Ark! This
document is outlines ways in which you can contribute to Ark, and also
some helpful things to know if you are working on the compiler.

## Index
* [Other Links](#other-links)
* [Communication](#communication)
	* [IRC](#irc)
	* [Reddit](#reddit)
* [Direct Contribution](#direct-contribution)
	* [GitHub Issues](#github-issues)
		* [Issue Labels](#issue-labels)
	* [Pull Requests](#pull-requests)
* [Utilities](#utilities)
	* [`make gen` and `make fmt`](#make-gen-and-make-fmt)
	* [Testing](#testing)
	* [Writing Tests](#writing-tests)

## <a name="other-links"></a> Other Links
* [/r/ark_lang](//reddit.com/r/ark_lang)
* [ark documentation](//book.ark-lang.org)
* [ark website](//ark-lang.org)

## <a name="communication"></a> Communication
### <a name="irc"></a> IRC
The preferred method of communication between developers is
the IRC, this is on the `irc.freenode.net` server, and our 
channel is `#ark-lang`.

### <a name="reddit"></a> Reddit
We also have a sub-reddit for discussion, help, and more. A lot
of the developers are Reddit addicts, so we can help you out
with any questions. We usually post ocassional updates to the
subreddit.

## <a name="direct-contribution"></a> Direct Contribution
### <a name="github-issues"></a> GitHub Issues
This is a more formal method of communcation, any bugs, errors,
or discussion that should be documented in some way goes here.

#### <a name="issue-labels"></a> Issue Labels
To keep things organized, we have introduced various labels that contributors with 
sufficient permissions can add to issues. **This also means that anyone
who is new and looking for issues related to a specific field, or complexity
can easily filter the issues to their advantage!**.

Labels are prefixed with a letter, the letter denotes the category of label:

* E: A label prefixed with E is how much experience is required to take on the issue
* M: These are miscellaneous issues, typically ones too broad to categorize
* O: These are the operating system/platform that an issue is relevant to
* P: The priority of the issue
* S: The stage in which the issue occurs (note not all issues have stages)
* A: An area in which the issue is relevant to, e.g. A-tests is for issues related to tests

### <a name="pull-requests"></a> Pull Requests
We love Pull Requests! To keep everything in working order, make
sure you follow these guidelines:

* Run `make fmt` on your code before you submit a PR
* Work on a separate branch, named appropriately e.g. `for-loop-codegen-fix`
* If your patch is not up to date with the master branch, please rebase
* If you are implementing a feature, make sure there is a text if there is not an existing one already
* Please add a summary/description of your pull request, the more detail the better.

Once your PR has been submitted, one of the developers will review your code
and merge it if everything is ok!

## <a name="utilities"></a> Utilities
### <a name="make-gen-and-make-fmt"></a> `make gen` and `make fmt`
The targets `gen` and `fmt` are included for the convenience of the developers. 
They run `go generate` and `go fmt` respectively on all the modules in Ark. 
Please run `make fmt` before creating a pull request.

### <a name="testing"></a> Testing
Requires `$GOPATH/bin` to be in your `$PATH` and Python 2.4 or newer.

### <a name="writing-tests"></a> Writing Tests
Writings tests are very important. The tests are located
in the `tests/` directory in this repository. Each test has
an ark module, and its corresponding TOML file.

In the TOML test file, you specify various arguments such as
what sourcefile to target, the compiler arguments to pass, and
so on.

Below is an example test file:

	Name       = "my_test"
	Sourcefile = "my_test.ark"

	CompilerArgs = []
	RunArgs      = []

	CompilerError = 0
	RunError      = 0

	Input = ""

	CompilerOutput = ""
	RunOutput      = ""

A test file should have a name that is meaningful, and shows
the developer at first glance, what is being tested.

You can run these tests by executing the `ark-testrunner` command
in your terminal (assuming `$GOPATH/bin` is in your `$PATH`).