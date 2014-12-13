COMPILER = clang
C_FLAGS = -Wall -std=c99 -Iincludes/
SOURCES = src/*.c

LLVM_C_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit native`

all: ${SOURCES}
	${COMPILER} -o j4 ${SOURCES} ${LLVM_C_FLAGS} ${C_FLAGS}
