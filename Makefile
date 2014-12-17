COMPILER = gcc
C_FLAGS = -g -o j4 -Wall -Iincludes/
SOURCES = src/*.c
LLVM_C_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit native`

all: ${SOURCES}
	${COMPILER} ${SOURCES} ${C_FLAGS}

j4: ${SOURCES}
	${COMPILER} ${SOURCES} ${LLVM_C_FLAGS} ${C_FLAGS}

clean: 
	-rm *.o