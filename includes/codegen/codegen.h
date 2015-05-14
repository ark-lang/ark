#ifndef __CODE_GEN_H
#define __CODE_GEN_H

/**
 * The code generator! This works in 2 stages, the first stage it
 * will try and emit code for all of the macros we're given. The second
 * stage is where we generate the "meat" of the program, all of the statements,
 * structures, declarations, etc are generated.
 */

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
#ifdef WINDOWS
	#define NEWLINE "\r\n"
#else
	#define NEWLINE "\n"
#endif
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
	#define CC_NEWLINE NEWLINE
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

	sds linkerFlags;

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
void emitCode(CodeGenerator *self, const char *fmt, ...);

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
 * Emits an array index
 * @param self the code gen instance
 * @param arrayIndex the array index
 */
void emitArrayIndex(CodeGenerator *self, ArrayIndex *arrayIndex);

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

/**
 * This will emit a declaration top level node
 * @param self the code gen instance
 * @param decl the decl node to emit
 */
void emitDeclaration(CodeGenerator *self, Declaration *decl);

/**
 * This will emit a return statement
 * @param self the code gen instance
 * @param ret  the return statement node to emit
 */
void emitReturnStat(CodeGenerator *self, ReturnStat *ret);

/**
 * This will emit a leave statement, which is
 * a top level node for Return, Break and Continue
 * @param self  the code gen instance
 * @param leave the leave statement to emit
 */
void emitLeaveStat(CodeGenerator *self, LeaveStat *leave);

/**
 * This will emit the top level node for a statement
 * @param self the code gen instance
 * @param stmt the statement to emit
 */
void emitStatement(CodeGenerator *self, Statement *stmt);

/**
 * This will emit a top level unstructed node
 * @param self the code gen instance
 * @param stmt the unstructured node to emit
 */
void emitUnstructuredStat(CodeGenerator *self, UnstructuredStatement *stmt);

/**
 * This will emit a structured statement, a structured
 * statement is something with a block or some form of structure,
 * for instance a for loop, or an if statement.
 * @param self the code gen instance
 * @param stmt the structured statement to emit
 */
void emitStructuredStat(CodeGenerator *self, StructuredStatement *stmt);

/**
 * Consumes a node in the AST that we're parsing
 * @param self [description]
 */
void consumeAstNode(CodeGenerator *self);

/**
 * Jumps ahead in the AST we're parsing
 * @param self   the code gen instance
 * @param amount the amount to consume by
 */
void consumeAstNodeBy(CodeGenerator *self, int amount);

/**
 * Run through all the nodes in the AST and
 * generate the code for them!
 * @param self the code gen instance
 */
void traverseAST(CodeGenerator *self);

/**
 * This is pretty much where the magic happens, this will
 * start the code gen
 * @param self the code gen instance
 */
void startCodeGeneration(CodeGenerator *self);

/** MACROS */

/**
 * Emit a macro for a file inclusion
 * @param self the code gen instance
 * @param use  the use macro
 */
void emitUseMacro(CodeGenerator *self, UseMacro *use);

/**
 * Emit a top level node for macros
 * @param self  the code gen instance
 * @param macro the macro to emit
 */
void emitMacroNode(CodeGenerator *self, Macro *macro);

/**
 * Generates the code for all the macros, the code generator
 * is currently in 2 passes, the first will generate the code
 * for all of the macros, the second will generate code for other
 * statements
 * @param self the code gen instance
 */
void generateMacros(CodeGenerator *self);

/**
 * Destroys the given code gen instance
 * @param self the code gen instance
 */
void destroyCodeGenerator(CodeGenerator *self);

#endif // __CODE_GEN_H
