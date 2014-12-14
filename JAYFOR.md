# JAYFOR
JAYFOR is a compiled programming language

## Frontend
The frontend of JAYFOR is fairly simple, we use a hand written Lexer and Parser. 

### Why a hand written Lexer/Parser?
* I've never used a lexer/parser generator 
* I like writing things from scratch
* Fun
* Learning more about lexers/parsers
* Tailoring the lexer/parser specifically to the J4 language
* Less bloat

## Backend
The backend uses LLVM. We use LLVM because
* More optimisations in there that I could do in a lifetime
* I've never written a Bytecode Compiler
* It's super duper fast
* Easy to use