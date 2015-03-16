CC = clang
C_FLAGS = -Wall -Iincludes/ -g
SOURCES = src/*.c

all: ${SOURCES}
	@mkdir -p bin
	${CC} ${C_FLAGS} ${SOURCES} -o bin/alloyc
	@rm -rf bin/alloyc.dSYM/
clean:
	@rm -rf bin

.PHONY: clean
