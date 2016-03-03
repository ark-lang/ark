# Standard Library
The standard library for Ark should contain the bare
minimum for a programmer to accomplish their daily
tasks. The library should adopt a consistent style
guide and paradigm for the implementation.

At least, the library should contain libraries for:
* Input and output (buffered)
  * Reader
  * Writer
  * File IO wrappers on top of the basic reader/writer?
* Basic data structures
  * HashMap
  * List [x]
  * Linked List
  * Stack [x]
  * Ordered binary tree?
* Formatting for I/O
* Math
  * Big math
  * Psuedo-random generators
  * Complex (128 bit stuff)
* OS
  * Processes
  * Signals
  * System calls?
* Path
* Regex
* Sorting
* Compression
  * zlib
  * gzip
  * lzw
  * DEFLATE
* Concurrency
  * Threads
  * Mutexes
  * Channels
* Date/time
* Strings
* UTF-8 [-]
* Iterators

Probably an official external library"
* Cryptography
  * AES 128/256
  * SHA 1/256/512
  * MD4
  * RSA
  * Blowfish & Twofish?
  
### Notes
Cryptography will likely be done via a C library wrapper as an external library.

I'm open to adding more data structures, but right now they seem like the most
common data structures that we should implement.

For the HashMap I'm thinking we could has SipHash, though it is known to be
quite slow it's cryptographically secure. We could probably implement an
alternative HashMap that is faster and is used for cases where you don't
care about collisions, etc.

The Linked List would be a doubly linked list, which means it can be used
as a queue too and is quite fast.

The formatting library for IO would eventually take advantage of the macro
system that doesn't yet exist. 

With regards to paths, I'm not sure if this should go under OS, or IO or maybe
an IO utility module?
