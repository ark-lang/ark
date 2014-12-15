COMPILER = clang
C_FLAGS = -c -g -Wall -Iincludes/
SOURCES = src/*.c
LLVM_C_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit native`

all: ${SOURCES}
	${COMPILER} ${SOURCES} ${LLVM_C_FLAGS} ${C_FLAGS}
