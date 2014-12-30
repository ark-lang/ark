CC = clang
C_FLAGS = -o j4 -fomit-frame-pointer -Wall -Iincludes/ 
LLVM_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit interpreter native`
SOURCES = src/*.c

all: ${SOURCES}
	${CC} ${SOURCES} ${C_FLAGS}

clean: 
	-rm *.o

.PHONY: clean