# [Alloy](http://alloy-lang.org) [![BuildStatus](https://travis-ci.org/alloy-lang/alloy.svg?branch=master)](https://travis-ci.org/alloy-lang/alloy)
Alloy is a work in progress programming language, read the [reference](docs/REFERENCE.md) to find out more about the plans, syntax, etc. Please note that the language is constantly being changed, so the Reference may be out-dated, or a little bit behind the master.

## Disclaimer?
Alloy is a somewhat experimental language, this means it can and probably will change at any point, be it a small change, or a complete re-write. This means we cannot promise any backwards compatibility. 

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
