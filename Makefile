# local compiler stuff
LCC = clang
LCXX = clang++

C_FLAGS = `llvm-config --cflags` -Wall -Iincludes/
LLVM_FLAGS = `llvm-config --libs --cflags --ldflags core analysis executionengine jit interpreter native`
SOURCES = src/*.c

# this might just work...?
travis: ${SOURCES}
	${CC} ${C_FLAGS} ${SOURCES} -c ${SOURCES}
	${CC}++ *.o ${LLVM_FLAGS} -o j4
	-rm *.o

all: ${SOURCES}
	${LCC} ${C_FLAGS} ${SOURCES} -c ${SOURCES}
	${LCXX} *.o ${LLVM_FLAGS} -o j4
	-rm *.o

clean: 
	-rm *.o

.PHONY: clean