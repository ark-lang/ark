CFLAGS = -g -Wall -std=c99 -Wextra -pedantic -Wno-unused-function -Wno-unused-parameter

ifdef ENABLE_LLVM
CFLAGS += `llvm-config --cflags` -DENABLE_LLVM
LDFLAGS = `llvm-config --cxxflags --ldflags --libs core analysis native bitwriter --system-libs`
endif

.PHONY: all clean

# Source/Header Files
INCLUDES = -Iincludes/ \
		   -Iincludes/codegen \
		   -Iincludes/lexer \
		   -Iincludes/parser \
		   -Iincludes/util \
		   -Iincludes/semantic \

SOURCES = $(wildcard src/*.c) \
		  $(wildcard src/codegen/C/*.c) \
		  $(wildcard src/lexer/*.c) \
		  $(wildcard src/parser/*.c) \
		  $(wildcard src/util/*.c) \
		  $(wildcard src/semantic/*.c) \

ifdef ENABLE_LLVM
SOURCES += $(wildcard src/codegen/LLVM/*.c) 
endif

ifndef ENABLE_LLVM
all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) $(INCLUDES) ${SOURCES} -o bin/alloyc

else
all: ${SOURCES}
	@mkdir -p bin/
	@${CC} ${CFLAGS} $(INCLUDES) ${SOURCES} -c ${SOURCES}
	@${CXX} *.o ${LDFLAGS} -o bin/alloyc 
	@-rm *.o
endif

clean:
	@rm -f *.o	
	@rm -rf bin/
	@rm -f _gen_*	# remove any output code if it's there
	@rm -rf *.dSYM/
	@rm -rf tests/*.test
	@rm -rf tests/*.dSYM/
