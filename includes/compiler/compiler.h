#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser/parser.h"
#include "util/vector.h"
#include "util/hashmap.h"

#define SPACE_CHAR " "
#define OPEN_BRACKET "("
#define CLOSE_BRACKET ")"
#define OPEN_BRACE "{"
#define CLOSE_BRACE "}"
#define CONST_KEYWORD "const"
#define ASTERISKS "*"
#define NEWLINE "\n"
#define TAB "\t"
#define EQUAL_SYM "="
#define SEMICOLON ";"
#define COMMA_SYM ","

typedef struct {
	Vector *abstractSyntaxTree;
	Vector *references;

    int sourceFileSize;
    int sourcePosition;
	char *sourceName;
	char *sourceContents;

	int currentNode;

	// timers for benchmarking
	clock_t timer;
	double secondsTaken;
	double msTaken;
} Compiler;

Compiler *createCompiler();

void appendToSource(Compiler *self, char *str);

void emitExpression(Compiler *self, ExpressionAstNode *expr);

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var);

void emitReturnStatement(Compiler *self, FunctionReturnAstNode *ret);

void emitFunctionCall(Compiler *self, FunctionCallAstNode *call);

void emitBlock(Compiler *self, BlockAstNode *block);

void emitArguments(Compiler *self, Vector *args);

void emitFunction(Compiler *self, FunctionAstNode *func);

void consumeAstNode(Compiler *self);

void consumeAstNodeBy(Compiler *self, int amount);

void startCompiler(Compiler *self, Vector *ast);

void destroyCompiler(Compiler *self);

#endif // compiler_H
