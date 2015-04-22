CFLAGS = -g -Wall -std=c99 -Wextra -pedantic

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

all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) $(INCLUDES) ${SOURCES} -o bin/alloyc

clean:
	@rm -f *.o	
	@rm -rf bin/
	@rm -f _gen_*	# remove any output code if it's there
	@rm -rf *.dSYM/
