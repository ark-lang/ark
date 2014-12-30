CC = clang
CPPC = clang++

C_FLAGS   = -o j4 -Wall -Iincludes/ 
LLVM_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit interpreter native`

C_SOURCES = src/*.c
CPP_SOURCES = src/*.cpp

all: ${SOURCES}
	${CC} ${C_FLAGS} ${C_SOURCES}

clean: 
	-rm *.o

.PHONY: clean