CC = clang
C_FLAGS = -Wall -Iincludes/ -g
SOURCES = src/*.c


all: ${SOURCES}
	${CC} ${C_FLAGS} ${SOURCES} -o bin/alloyc
	mkdir -p bin

clean:
	rm -f bin/alloyc

.PHONY: clean