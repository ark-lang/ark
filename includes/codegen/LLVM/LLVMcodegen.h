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

	/**
	 * Asm files that are generated from the LLVM IR
	 */
	Vector *asmFiles;

	/**
	 * The current source file to
	 * generate code for
	 */
	SourceFile *currentSourceFile;

	/**
	 * IR Builder Reference
	 */
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

/**
 * Gets the integer type based on the
 * computers architecture.
 *
 * @return the type as a llvm ref
 */
LLVMTypeRef getIntType();

/**
 * Converts a DataType to an LLVM Type
 * @param  type the type to convert
 * @return      the reference to the converted type
 */
LLVMTypeRef getLLVMType(DataType type);

/**
 * Generates code for a function signature
 * @param  self the codegen instance
 * @param  decl functionsignature node to generate
 * @return      the reference of the function signature
 */
LLVMValueRef genFunctionSignature(LLVMCodeGenerator *self, FunctionDecl *decl);

/**
 * Generates a top level statement unstructured or structured
 * @param  self the codegen instance
 * @param  stmt the statement node to generate
 * @return      the reference of the statement node generated
 */
LLVMValueRef genStatement(LLVMCodeGenerator *self, Statement *stmt);

/**
 * Generates a function declaration
 * @param  self the codegen instance
 * @param  decl the decl to generate for
 * @return      the reference of the function decl node generated
 */
LLVMValueRef genFunctionDecl(LLVMCodeGenerator *self, FunctionDecl *decl);

/**
 * Generates a top level declaration node
 * @param  self the codegen instance
 * @param  decl the decl node to generate for
 * @return      the reference to the generated node
 */
LLVMValueRef genDeclaration(LLVMCodeGenerator *self, Declaration *decl);

/**
 * Generates an unstructured statement node
 * @param  self the codegen instance
 * @param  stmt the statement node to generate for
 * @return      the reference to the generated node
 */
LLVMValueRef genUnstructuredStatementNode(LLVMCodeGenerator *self, UnstructuredStatement *stmt);

/**
 * Generates a structured statement node
 * @param  self the codegen instance
 * @param  stmt the statement node to generate for
 * @return      the reference to the generated node
 */
LLVMValueRef genStructuredStatementNode(LLVMCodeGenerator *self, StructuredStatement *stmt);

/**
 * Generates a binary expression
 * @param  self the codegen instance
 * @param  expr the expression to generate
 * @return      the reference for the binary expression
 */
LLVMValueRef genBinaryExpression(LLVMCodeGenerator *self, BinaryExpr *expr);

/**
 * Generates an expression
 * @param  self the codegen instance
 * @param  expr the expression to generate
 * @return      the reference for the expression
 */
LLVMValueRef genExpression(LLVMCodeGenerator *self, Expression *expr);

/**
 * Generates a function call
 * @param  self the codegen instance
 * @param  call the call node to generate
 * @return      the reference for the call
 */
LLVMValueRef genFunctionCall(LLVMCodeGenerator *self, Call *call);

/**
 * Generates a Type Name
 * @param  self the codegen instance
 * @param  name the node to generate
 * @return      the reference for the typename
 */
LLVMValueRef genTypeName(LLVMCodeGenerator *self, TypeName *name);

/**
 * Generates a literal node
 * @param  self the codegen instance
 * @param  lit  the literal node to generate for
 * @return      the reference for the literal
 */
LLVMValueRef genLiteral(LLVMCodeGenerator *self, Literal *lit);

/**
 * Generates a type lit
 * @param  self the codegen instance
 * @param  lit  the literal node to generate
 * @return      the reference for the type literal
 */
LLVMValueRef genTypeLit(LLVMCodeGenerator *self, TypeLit *lit);

/**
 * Generates a type parent node
 * @param  self the codegen instance
 * @param  type the type to generate for
 * @return      the reference to the generated node
 */
LLVMValueRef genType(LLVMCodeGenerator *self, Type *type);

/**
 * Generates a leave statement node, for instance
 * a return type, break or continue
 * @param  self  the codegen instance
 * @param  leave the node to generate for
 * @return       the reference to the generated node
 */
LLVMValueRef genLeaveStatNode(LLVMCodeGenerator *self, LeaveStat *leave);

/**
 * Creates bitcode from the sourcefiles modules
 * @param  self the codegen instance
 * @return      the filename of the generated bitcode (i.e. file.bc)
 */
char *createBitcode(LLVMCodeGenerator *self);

/**
 * Finds the bitcode file and converts it to assembly
 * @param self        the codegen instance
 * @param bitcodeName the name of the bitcode file to convert
 */
void convertBitcodeToAsm(LLVMCodeGenerator *self, sds bitcodeName);

/**
 * Creates a binary file from the assembly created
 * @param self the codegen instance
 */
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
