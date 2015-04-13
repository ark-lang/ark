# [Alloy](http://alloy-lang.org) [![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)
Alloy is a work in progress programming language, read the [reference](docs/REFERENCE.md) to find out more about the plans, syntax, etc. Please note that the language is constantly being changed, so the Reference may be outdated, or a little bit behind the master.

## Disclaimer?
Alloy is a somewhat experimental language, this means it can and probably will change at any point, be it a small change, or a complete re-write. This means we cannot promise any backwards compatibility. 

## TODO
Here are the list of things that need to be done in order for the first iteration to be complete.

* [ ] Member Access: `some_struct.some_array[3].function()`, etc.
* [ ] Unary operators for address of and dereferencing: `x: ^int = &y;`, `do_stuff(^x);`
* [ ] Explicit casting, perhaps we should try and get the compiler to take care of most of this, but in some cases you want the value to be casted to another type
* [ ] File inclusion, this is mostly done but support should be added for files in another directory
* [ ] Option Types: `fn add(a: int, b: int): <int>;`
* [ ] Match statements
* [ ] Memory Model
* [ ] Macro System (?)
* [ ] Macro for whether a file is compiled on "use"

## Status
Alloy is still constantly being worked on. At the moment it compiles,
however some aspects of the language are unimplemented or broken.

## Just show me what it looks like already
Sure, you can either see a small virtual machine implemented in Alloy [here](tests/misc/virtualmachine.aly). Or you can just see a small Hello World example below:

```rust
use "stdio"

fn main(): int {
    println("Hello, World");
    return 0;
}
```

## How can I help out with Alloy?
There are loads of ways you can help out:

* Sending PR's (features, bug fixes, etc)
* Commenting on issues
* Creating issues (proposals, bugs, etc)
* Writing documentation

## License
Alloy is licensed under the [MIT License](/LICENSE.md).
