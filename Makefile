COMPILER = gcc
C_FLAGS = -g -o builds/j4 -Wall -Iincludes/
SOURCES = src/*.c

all: ${SOURCES}
	${COMPILER} ${SOURCES} ${C_FLAGS}

clean: 
	-rm *.o