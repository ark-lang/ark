#ifndef __CODE_GEN_H
#define __CODE_GEN_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
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

#define COMPACT_CODE_GEN false

#if COMPACT_CODE_GEN == true
	#define CC_NEWLINE "\n"
#else
	#define CC_NEWLINE " "
#endif

typedef enum {
	WRITE_HEADER_STATE,
	WRITE_SOURCE_STATE
} WriteState;

typedef struct {
	Vector *abstractSyntaxTree;
	Vector *sourceFiles;

	SourceFile *currentSourceFile;
	map_t symtable;
	WriteState writeState;

	int currentNode;
} CodeGenerator;

CodeGenerator *createCodeGenerator(Vector *sourceFiles);

void emitLiteral(CodeGenerator *self, Literal *lit);

void emitBinaryExpr(CodeGenerator *self, BinaryExpr *expr);

void emitUnaryExpr(CodeGenerator *self, UnaryExpr *expr);

void emitExpression(CodeGenerator *self, Expression *expr);

void emitCode(CodeGenerator *self, char *fmt, ...);

void emitTypeLit(CodeGenerator *self, TypeLit *lit);

void emitType(CodeGenerator *self, Type *type);

void emitParameters(CodeGenerator *self, Parameters *params);

void emitFieldList(CodeGenerator *self, FieldDeclList *list);

void emitStructDecl(CodeGenerator *self, StructDecl *decl);

void emitFunctionDecl(CodeGenerator *self, FunctionDecl *decl);

void emitVariableDecl(CodeGenerator *self, VariableDecl *decl);

void emitDeclaration(CodeGenerator *self, Declaration *decl);

void emitUnstructuredStat(CodeGenerator *self, UnstructuredStatement *stmt);

void emitStructuredStat(CodeGenerator *self, StructuredStatement *stmt);

void consumeAstNode(CodeGenerator *self);

void consumeAstNodeBy(CodeGenerator *self, int amount);

void traverseAST(CodeGenerator *self);

void startCodeGeneration(CodeGenerator *self);

void destroyCodeGenerator(CodeGenerator *self);

#endif // __CODE_GEN_H
