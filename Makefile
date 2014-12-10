j4mingw: src/*.c
	gcc -o j4 -Iincludes/ src/*.c -Wall

j4clang: src/*.c
	clang -o j4 -Iincludes/ src/*.c -Wall
