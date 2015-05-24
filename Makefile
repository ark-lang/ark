CFLAGS = -g -Wall -std=c99 -Wextra -pedantic -Wno-unused-function -Wno-unused-parameter `llvm-config --cflags` -DENABLE_LLVM -w
LDFLAGS = `llvm-config --cxxflags --ldflags --libs core analysis native bitwriter --system-libs`

.PHONY: all clean

# Source/Header Files
INCLUDES = -Iincludes/ \
		   -Iincludes/codegen \
		   -Iincludes/lexer \
		   -Iincludes/parser \
		   -Iincludes/util \
		   -Iincludes/semantic \

SOURCES = $(wildcard src/*.c) \
		  $(wildcard src/lexer/*.c) \
		  $(wildcard src/parser/*.c) \
		  $(wildcard src/util/*.c) \
		  $(wildcard src/semantic/*.c) \
 		  $(wildcard src/codegen/LLVM/*.c) \

all: ${SOURCES}
	@mkdir -p bin/
	@${CC} ${CFLAGS} $(INCLUDES) ${SOURCES} -c ${SOURCES}
	@${CXX} *.o ${LDFLAGS} -o bin/arkc 
	@-rm *.o

clean:
	@rm -f *.o	
	@rm -rf bin/
	@rm -f _gen_*	# remove any output code if it's there
	@rm -rf *.dSYM/
	@rm -rf tests/*.test
	@rm -rf tests/*.dSYM/
