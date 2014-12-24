CC = clang
C_FLAGS = -o j4 -fomit-frame-pointer -Wall -Iincludes/ 
SOURCES = src/*.c

all: ${SOURCES}
	${CC} ${SOURCES} ${C_FLAGS}

clean: 
	-rm *.o