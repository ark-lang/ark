.PHONY: all clean

# Source/Header Files
INCLUDES = -Iincludes/ \
		   -Iincludes/codegen \
		   -Iincludes/lexer \
		   -Iincludes/parser \
		   -Iincludes/util \
		   -Iincludes/semantic \

SOURCES = $(wildcard src/*.c) \
		  $(wildcard src/codegen/*.c) \
		  $(wildcard src/lexer/*.c) \
		  $(wildcard src/parser/*.c) \
		  $(wildcard src/util/*.c) \
		  $(wildcard src/semantic/*.c) \

CC = clang
CFLAGS = -g -Wall -std=c99 -Wextra

all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) $(INCLUDES) ${SOURCES} -o bin/alloyc

clean:
	@rm *.o	
	@rm bin/
	@rm _gen_* 		# remove any output code if it's there
