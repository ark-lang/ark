CC = clang
C_FLAGS = -Wall -Iincludes/ -g
SOURCES = src/*.c

all: ${SOURCES}
	${CC} ${C_FLAGS} ${SOURCES}
	mkdir -p bin
	-rm *.o


# clean stuff up
clean:
	-rm -f bin/alloyc
	rm *.o

.PHONY: clean
