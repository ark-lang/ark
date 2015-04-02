# Pretty meh Makefile, we could probably clean this up a lot

# Source/Header Files
INCLUDES = -Iincludes/ -Iincludes/codegen -Iincludes/lexer -Iincludes/parser -Iincludes/util
SOURCES = $(wildcard src/*.c) \
		  $(wildcard src/codegen/*.c) \
		  $(wildcard src/lexer/*.c) \
		  $(wildcard src/parser/*.c) \
		  $(wildcard src/util/*.c) \

# Flags n stuff
CC = clang
CFLAGS = -g -Wall `llvm-config --cflags` -I`llvm-config --includedir`
LD=clang++
LDFLAGS=`llvm-config --cxxflags --ldflags --libs core executionengine jit interpreter analysis native bitwriter --system-libs`

all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) $(INCLUDES) ${SOURCES} -c ${SOURCES}
	$(LD) *.o $(LDFLAGS) -o bin/alloyc
	@rm *.o

clean:
	@rm *.o
	@rm -rf bin

.PHONY: clean
