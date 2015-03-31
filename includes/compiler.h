#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "parser.h"
#include "vector.h"
#include "hashmap.h"

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
	Vector *sourceFiles;
	SourceFile *currentSourceFile;

	map_t functions;
	map_t structures;
	map_t variables;

	int currentNode;
} Compiler;

Compiler *createCompiler(Vector *sourceFiles);

void emit(Compiler *self, char *fmt, ...);

void emitExpression(Compiler *self, Expression *expr);

void emitType(Compiler *self, Type *type);

void emitParameters(Compiler *self, Parameters *params);

void emitFunctionSignature(Compiler *self, FunctionSignature *func);

void emitStructuredStatement(Compiler *self, StructuredStatement *stmt);

void emitUnstructuredStatement(Compiler *self, UnstructuredStatement *stmt);

void emitBlock(Compiler *self, Block *block);

void emitForStat(Compiler *self, ForStat *stmt);

void emitIfStat(Compiler *self, IfStat *stmt);

void emitMatchStat(Compiler *self, MatchStat *stmt);

void emitStatementList(Compiler *self, StatementList *stmtList);

void emitFunctionDecl(Compiler *self, FunctionDecl *decl);

void emitStructDecl(Compiler *self, StructDecl *decl);

void emitDeclaration(Compiler *self, Declaration *decl);

void consumeAstNode(Compiler *self);

void consumeAstNodeBy(Compiler *self, int amount);

void startCompiler(Compiler *self);

void compileAST(Compiler *self);

void destroyCompiler(Compiler *self);

#endif // compiler_H
