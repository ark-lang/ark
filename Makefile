# local compiler stuff
LCC = clang
LCXX = clang++

C_FLAGS = `llvm-config --cflags` -Wall -Iincludes/ -std=c99
LLVM_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit interpreter native`
SOURCES = src/*.c

BUILD_COMMAND = 

ifeq ($(CC),gcc)
	BUILD_COMMAND = g++ *.o ${LLVM_FLAGS} -o j4 -ldl -ltinfo -pthread
else
	BUILD_COMMAND = ${CC}++ *.o ${LLVM_FLAGS} -o j4 
endif

all: ${SOURCES}
	${LCC} ${C_FLAGS} ${SOURCES} -c ${SOURCES}
	${LCXX} *.o ${LLVM_FLAGS} -o j4
	-rm *.o

travis: ${SOURCES}
	${CC} ${C_FLAGS} ${SOURCES} -c ${SOURCES}
	${BUILD_COMMAND}
	-rm *.o

clean: 
	-rm *.o

.PHONY: clean
