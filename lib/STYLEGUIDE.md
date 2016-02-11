# Introduction
This is the style guide for writing Ark code, you should follow this when writing code for the standard library, though we suggest you follow it outside of the standard library too.

## Modules
A module name must be lowercase, each word separated by underscores, though prefer keeping
everything to a single word unless necessary:

    - data_structures/
        - array.ark
        - linked_list.ark

## Variables
Use camelCase for variables, constants in capitalized snake_case:

    SNAKE_CASE_CONSTANT := 12; // const
    mut fooBar := SNAKE_CASE_CONSTANT;
  
## Bracing
Bracing should be on the same line as the
declaration.

    if foo == 12 {
    
    } else if bar == 32 {
    
    } else {
        
    }
  
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
Any type that you introduce should be in CamelCase:

    type Foo struct {
    
    };
    
    type Person struct {
    
    };
    
    type JonathanBlow []int;
