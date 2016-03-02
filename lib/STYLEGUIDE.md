# Introduction
This is the style guide for writing Ark code, you should follow this when writing code for the standard library, though we suggest you follow it outside of the standard library too.

## Comments
### Single Line Comments

    // This is an ordinary comment.
    // Nothing to see here, though
    // preferably you want to keep
    // them relatively short for
    // easier reading

### Multiple Line Comments

    /*
        One asterisks preferably,
        the content of the comment
        is indented.
    */

### Documentation Comments

    /// Documentation comments are
    /// denoted with three slashes.
    ///
    /// They contain markdown, and when
    /// writing code we suggest you use
    /// it for fancy documentation!
    ///
    /// ## Examples
    /// 
    /// ```
    /// type stuff int;
    /// please := int(64);
    /// do := stuff(please);
    /// ```

## Modules
A module name must be lowercase, each word separated by underscores, though prefer keeping
everything to a single word unless necessary:

    - data_structures/
        - array.ark
        - linked_list.ark

## Variables
Use camelCase for variables, global constants in 
capitalized snake_case:

    // global constant
    SNAKE_CASE_CONSTANT := 12; // const

    // mutable local
    mut fooBar := SNAKE_CASE_CONSTANT;
    
    // constnat local
    barFoo := SNAKE_CAST_CONSTANT:

## Bracing
Bracing should be on the same line as the
declaration.

    if foo == 12 {
        ...
    } else if bar == 32 {
        ...
    } else {
        ...
    }

    match foo {
        blah -> {
            ...
        },
        foo -> {
            ...
        }, 
    }

    type Foo T {
        ...
    };
  
## Commas
Prefer to add in a trailing comma to structures,
interfaces, etc:

    match foo {
        5 -> _,
        6 -> _,
        _ -> _, // trail...
    };
  
    type Blah interface {
        func foo() -> int,
        func bar() -> int, // trail...
    }
  
## Functions
Functions are in camelCase:

    func fooBar() -> int {
    
    }

## Named Types
Any type that you introduce should be in PascalCase:

    type Foo struct {
    
    };
    
    type Person struct {
    
    };
    
    type JonathanBlow []int;
