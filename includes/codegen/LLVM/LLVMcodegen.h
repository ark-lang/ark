#ifndef __LLVM_CODEGEN_H
#define __LLVM_CODEGEN_H

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

typedef struct {
	/**
	 * The current abstract syntax tree being
	 * generated for
	 */
	Vector *abstractSyntaxTree;
	
	/**
	 * All of the source files we're generating code
	 * for
	 */
	Vector *sourceFiles;

	Vector *asmFiles;

	/**
	 * The current source file to
	 * generate code for
	 */
	SourceFile *currentSourceFile;

	LLVMBuilderRef builder;

	/**
	 * Our index in the abstract syntax
	 * tree
	 */
	int currentNode;
} LLVMCodeGenerator;

/**
 * Creates an instance of the code generator
 * @param  sourceFiles the source files to codegen for
 * @return             the instance we created
 */
LLVMCodeGenerator *createLLVMCodeGenerator(Vector *sourceFiles);

/**
 * This is pretty much where the magic happens, this will
 * start the code gen
 * @param self the code gen instance
 */
void startLLVMCodeGeneration(LLVMCodeGenerator *self);

/**
 * Destroys the given code gen instance
 * @param self the code gen instance
 */
void destroyLLVMCodeGenerator(LLVMCodeGenerator *self);

static LLVMTypeRef getIntType();

static LLVMTypeRef getLLVMType(DataType type);

LLVMValueRef genFunctionSignature(LLVMCodeGenerator *self, FunctionSignature *decl);

LLVMValueRef genStatement(LLVMCodeGenerator *self, Statement *stmt);

LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl);

LLVMValueRef genDeclaration(LLVMCodeGenerator *self, Declaration *decl);

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt);

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt);

LLVMValueRef genBinaryExpression(LLVMCodeGenerator *self, BinaryExpr *expr);

LLVMValueRef genExpression(LLVMCodeGenerator *self, Expression *expr);

LLVMValueRef genBinaryExpression(LLVMCodeGenerator *self, BinaryExpr *expr);

LLVMValueRef genFunctionCall(LLVMCodeGenerator *self, Call *call);

LLVMValueRef genTypeName(LLVMCodeGenerator *self, TypeName *name);

LLVMValueRef genLiteral(LLVMCodeGenerator *self, Literal *lit);

LLVMValueRef genTypeLit(LLVMCodeGenerator *self, TypeLit *lit);

LLVMValueRef genType(LLVMCodeGenerator *self, Type *type);

LLVMValueRef genExpression(LLVMCodeGenerator *self, Expression *expr);

LLVMValueRef genFunctionSignature(LLVMCodeGenerator *self, FunctionSignature *decl);

LLVMValueRef genStatement(LLVMCodeGenerator *self, Statement *stmt);

LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl);

LLVMValueRef genDeclaration(LLVMCodeGenerator *self, Declaration *decl);

LLVMValueRef genLeaveStatNode(LLVMCodeGenerator *self, LeaveStat *leave);

LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt);

LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt);

char *createBitcode(LLVMCodeGenerator *self);

void convertBitcodeToAsm(LLVMCodeGenerator *self, sds bitcodeName);

void createBinary(LLVMCodeGenerator *self);

/**
 * Jumps ahead in the AST we're parsing
 * @param self   the code gen instance
 * @param amount the amount to consume by
 */
static void consumeAstNodeBy(LLVMCodeGenerator *self, int amount);

/**
 * Run through all the nodes in the AST and
 * generate the code for them!
 * @param self the code gen instance
 */
static void traverseAST(LLVMCodeGenerator *self);

#endif
