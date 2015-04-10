#ifndef __CODE_GEN_H
#define __CODE_GEN_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "parser.h"
#include "vector.h"
#include "hashmap.h"

/**
 * Constants that are generated in the
 * code should be defined here so we can
 * change them in the future if we need to
 * etc...
 */
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

/**
 * This defined whether the code we 
 * produce is somewhat readable, or
 * all minified
 */
#define COMPACT_CODE_GEN false

#if COMPACT_CODE_GEN == false
	#define CC_NEWLINE "\n"
#else
	#define CC_NEWLINE " "
#endif

/**
 * The two types of state that
 * we can be in, this makes it cleaner
 * than switching between files or writing
 * one after the other.
 */
typedef enum {
	WRITE_HEADER_STATE,
	WRITE_SOURCE_STATE
} WriteState;

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
	 * The current source file to
	 * generate code for
	 */
	SourceFile *currentSourceFile;
	
	/**
	 * The current write state, i.e
	 * are we writing to the header
	 * or source file
	 */
	WriteState writeState;

	/**
	 * Our index in the abstract syntax
	 * tree
	 */
	int currentNode;
} CodeGenerator;

/**
 * Creates an instance of the code generator
 * @param  sourceFiles the source files to codegen for
 * @return             the instance we created
 */
CodeGenerator *createCodeGenerator(Vector *sourceFiles);

/**
 * This will emit a literal to the source code
 * @param self the code gen instance
 * @param lit  the literal node
 */
void emitLiteral(CodeGenerator *self, Literal *lit);

/**
 * Emits a binary expression into our source code
 * @param self the code gen instance
 * @param expr the binary expr node to emit
 */
void emitBinaryExpr(CodeGenerator *self, BinaryExpr *expr);

/**
 * Emits a unary expression to the source code
 * @param self the code gen instance
 * @param expr the expression node to emit
 */
void emitUnaryExpr(CodeGenerator *self, UnaryExpr *expr);

/**
 * This emits an expression, which could be a function
 * call, unary expr, binary expr, etc
 * @param self the code gen instance
 * @param expr the expression node to emit
 */
void emitExpression(CodeGenerator *self, Expression *expr);

/**
 * This emits code to the current file declared
 * by the "WriteState".
 * @param self the code gen instance
 * @param fmt  the string to print
 * @param ...  additional parameters
 */
void emitCode(CodeGenerator *self, char *fmt, ...);

/**
 * Emits a type literal
 * @param self the code gen instance
 * @param lit  the literal node to emit
 */
void emitTypeLit(CodeGenerator *self, TypeLit *lit);

/**
 * Emits a function call, note that it does not 
 * emit the semi colon at the end, as this function
 * can be inside of an expression!
 * @param self the code gen instance
 * @param call the call to emit
 */
void emitFunctionCall(CodeGenerator *self, Call *call);

/**
 * Emits a type of any kind
 * @param self the code gen instance
 * @param type the type to emit
 */
void emitType(CodeGenerator *self, Type *type);

/**
 * This will emit the function parameters to the
 * source files
 * @param self   the code gen instance
 * @param params the parameters to emit
 */
void emitParameters(CodeGenerator *self, Parameters *params);

/**
 * This will emit a field list, a field list
 * is the declarations inside of a structure, i.e
 * 	int x;
 *  int y;
 * etc
 * 
 * @param self the code gen instance
 * @param list the list to emit
 */
void emitFieldList(CodeGenerator *self, FieldDeclList *list);

/**
 * This will emit a structure declaration, eventually
 * we should make this support declaration of values in
 * a structure.
 * @param self the code gen instance
 * @param decl the declaration to emit
 */
void emitStructDecl(CodeGenerator *self, StructDecl *decl);

/**
 * This will emit a function declaration
 * @param self the code gen instance
 * @param decl the function decl node to emit
 */
void emitFunctionDecl(CodeGenerator *self, FunctionDecl *decl);

/**
 * This will emit a for loop that has a single expression
 * or an "index", for instance:
 *
 * for true {
 * 
 * }
 * 
 * @param self the code gen instance
 * @param stmt the for loop to emit
 */
void emitWhileForLoop(CodeGenerator *self, ForStat *stmt);

/**
 * This will emit a block
 * @param self the code gen instance
 * @param block the block the emit
 */
void emitBlock(CodeGenerator *self, Block *block);

/**
 * This will emit a reassignment, i.e
 * x = 10;
 * 
 * @param self   the code gen instance
 * @param assign the assignment node to emit
 */
void emitAssignment(CodeGenerator *self, Assignment *assign);

/**
 * This will emit an infinite for loop, i.e a for loop
 * with no conditions
 * 
 * @param self the code gen instance
 * @param stmt the for loop to emit
 */
void emitInfiniteForLoop(CodeGenerator *self, ForStat *stmt);

/**
 * This will emit a for loop with two conditions
 * @param self the code gen instance
 * @param stmt for loop to emit code for
 */
void emitForStat(CodeGenerator *self, ForStat *stmt);

/**
 * This will emit a variable decl
 * @param self the code gen instance
 * @param decl the variable decl to emit
 */
void emitVariableDecl(CodeGenerator *self, VariableDecl *decl);

void emitDeclaration(CodeGenerator *self, Declaration *decl);

void emitReturnStat(CodeGenerator *self, ReturnStat *ret);

void emitLeaveStat(CodeGenerator *self, LeaveStat *leave);

void emitUnstructuredStat(CodeGenerator *self, UnstructuredStatement *stmt);

void emitStructuredStat(CodeGenerator *self, StructuredStatement *stmt);

void consumeAstNode(CodeGenerator *self);

void consumeAstNodeBy(CodeGenerator *self, int amount);

void traverseAST(CodeGenerator *self);

void startCodeGeneration(CodeGenerator *self);

void destroyCodeGenerator(CodeGenerator *self);

#endif // __CODE_GEN_H
