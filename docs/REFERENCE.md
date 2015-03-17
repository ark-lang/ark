# Alloy Reference
This is a reference for the Alloy programming language. Please note that the language is in constant development,
therefore this reference can become outdated at any time, and it can be days before it is updated to the current stage
of the compiler.

## Introduction
The aim of this compiler is to reinvent C. We love C, it's simple, fast, expressive, and cross-platform. However, it
has it's flaws:

* buffer overflows
* no string type
* no booleans without stdbool
* 30 years old
* inconsistencies, especially in error handling

Our goal with Alloy is to fix these errors, yet maintaining a cleaner, simpler syntax.
A lot of the syntax for Alloy is inspired by existing languages, such as Rust, Go, and Java. The language itself is
also heavily inspired by the simplicity of C. We don't want to add too much that it constrains the developer to a single paradigm.

## Memory Model
