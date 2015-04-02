#ifndef compiler_H
#define compiler_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/BitWriter.h>

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

typedef struct {
	Vector *abstractSyntaxTree;
	Vector *sourceFiles;
	
	// llvm stuff
	LLVMModuleRef module;
	LLVMExecutionEngineRef engine;
	LLVMBuilderRef builder;

	SourceFile *currentSourceFile;
	map_t symtable;
	int currentNode;
} Compiler;

Compiler *createCompiler(Vector *sourceFiles);

LLVMTypeRef getTypeRef(Type *type);

void generateFunctionCode(Compiler *self, FunctionDecl *decl);

void consumeAstNode(Compiler *self);

void consumeAstNodeBy(Compiler *self, int amount);

void startCompiler(Compiler *self);

void declarationDriver(Compiler *self, Declaration *decl);

void unstructuredStatementDriver(Compiler *self, UnstructuredStatement *stmt);

void structuredStatementDriver(Compiler *self, StructuredStatement *stmt);

void statementDriver(Compiler *self, Statement *stmt);

void destroyCompiler(Compiler *self);

#endif // compiler_H
