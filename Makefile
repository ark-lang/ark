.PHONY: all clean

# Source/Header Files
INCLUDES = -Iincludes/ -Iincludes/codegen -Iincludes/lexer -Iincludes/parser -Iincludes/util
SOURCES = $(wildcard src/*.c) \
		  $(wildcard src/codegen/*.c) \
		  $(wildcard src/lexer/*.c) \
		  $(wildcard src/parser/*.c) \
		  $(wildcard src/util/*.c) \

CC = clang
CFLAGS = -g -Wall 

all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) $(INCLUDES) ${SOURCES} -o bin/alloyc

clean:
	@rm *.o

.PHONY: clean
