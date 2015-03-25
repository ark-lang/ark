# Alloy Reference
This document is an informal specification for Alloy, a systems programming language.

## Guiding Principles
Alloy is a systems programming language, intended as an alternative to C. It's main purpose is to modernize C, without
deviating from C's original goal of simplicity. 

The design is motivated by the following:

* Fast
* Multi-Paradigm
* Strongly Typed
* Concise, clean and semantic syntax
* **No** garbage collection
* Efficient

The language should be strong enough that it can be self-hosted.

## Source Code Representation
The source code is in unicode text, encoded in utf-8. Source text is case-sensitive. Whitespace is blanks,
newlines, carriage returns, or tabs. Comments are denoted with `//` for single line, or `/* */` without nesting.
For simplicity, identifiers are treated as ASCII, however unicode may be supported in the future, but aren't a priority.