CC = clang
C_FLAGS = -g -o j4 -Wall -Iincludes/
SOURCES = src/*.c

all: ${SOURCES}
	${CC} ${SOURCES} ${C_FLAGS}

clean: 
	-rm *.o