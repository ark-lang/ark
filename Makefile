CC = clang
CPPC = clang++

C_FLAGS = `llvm-config --cflags` -Wall -Iincludes/
LLVM_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit interpreter native`
SOURCES = src/*.c

all: ${SOURCES}
	${CC} ${C_FLAGS} ${SOURCES} -c ${SOURCES}
	${CPPC} *.o ${LLVM_FLAGS} -o j4
	-rm *.o

clean: 
	-rm *.o

.PHONY: clean