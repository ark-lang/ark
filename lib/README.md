# stdlib
The standard library for the Ark programming language.

The standard library should be bundled with
the compiler when you install, but in the
mean time you can clone the stdlib and place
it in any projects you are working on.

Note there is no testing framework yet, but
for now I'm using a simple test project to
check that features work...

GNU Make is required to run the test case,
it will pass in the base directory (tests),
pass in the entry source code file (test.ark),
and it will also include `-I` the standard library
folder (`std/`).

## TODO
Some of these are blocked by language features
being incomplete, or not implemented at all.

* [ ] Memory allocation/destruction with generics
* [ ] File reading/writing
* [ ] Concatenation in println
	* [ ] Print, Print to fd, etc..
* [ ] Option types

## Contributing
Great! To contribute to the standard library, 
please see the [contributing](/CONTRIBUTING.md) file.

## License
Licensed under [MIT](/LICENSE)
