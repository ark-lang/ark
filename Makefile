CC = clang
C_FLAGS = -Wall -Iincludes/ -g
SOURCES = src/*.c

all: ${SOURCES}
	@mkdir -p bin
	@${CC} ${C_FLAGS} ${SOURCES} -o bin/alloyc
	@rm -rf bin/alloyc.dSYM/
	@echo "the alloyc executable can be found in the newly created bin/ directory."
clean:
	@rm -rf bin

.PHONY: clean
