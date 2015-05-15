#ifndef __SEMANTIC_H
#define __SEMANTIC_H

/**
 * Semantic analysis of our parser, not a high priority as
 * of writing this, so it's partially unimplemented.
 *
 * I decided to have multiple hashmaps so we don't have
 * to create any structs representing the various kinds of
 * nodes created, thus no memory management... etc
 *
 * This is also recursive, like the parser and codegenerator.
 */

#include "ast.h"
#include "vector.h"
#include "hashmap.h"
#include "sourcefile.h"
#include "parser.h"
#include "stack.h"

#define MAIN_FUNC "main"

typedef struct {
	/** The current AST being analyzed */
	Vector *abstractSyntaxTree;

	/** The current source file */
	SourceFile *currentSourceFile;

	/** the source files to semantically analyze */
	Vector *sourceFiles;

	/** data types for type inference */
	map_t dataTypes;

	/** the current node in the ast */
	int currentNode;

	/** if this stage failed or not */
	bool failed;

	Stack *scopes;
} SemanticAnalyzer;

// for the hacky type inference
typedef enum {
	INTEGER_VAR_TYPE,
	DOUBLE_VAR_TYPE,
	STRING_VAR_TYPE,
	CHAR_VAR_TYPE,
	STRUCTURE_VAR_TYPE,
	UNKNOWN_VAR_TYPE
} VariableType;

typedef struct {
	map_t funcSymTable;
	map_t varSymTable;
	map_t paramSymTable;
	map_t structSymTable;
	map_t implSymTable;
	map_t implFuncSymTable;
} Scope;

/*
 * Heap allocation hacky stuff
 */
typedef struct {
	int type;
} VarType; 

/*
 * Create a new Scope
 */
Scope *createScope();

/**
 * Destroys the given Scope, frees up
 * everything defined in the scope, e.g
 * variables, functions, etc.
 */
void destroyScope(Scope *scope);

VarType *createVarType(int type);

void destroyVarType(VarType *type);

/**
 * Create instance of semantic analyzer
 * @return             the instance created
 */
SemanticAnalyzer *createSemanticAnalyzer(Vector *sourceFiles);

/**
 * Analyze the block node given
 * @param self  the semantic analyzer instance
 * @param block the block node to analyze
 */
void analyzeBlock(SemanticAnalyzer *self, Block *block);

/**
 * Analyze a function decl
 * @param self the semantic analyzer instance
 * @param decl the function decl node to analyze
 */
void analyzeFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *decl);

/**
 * Analyze a struct decl
 * @param self the semantic analyzer instance
 * @param decl the struct decl
 */
void analyzeStructDeclaration(SemanticAnalyzer *self, StructDecl *decl);

/**
 * Analyze a variable declaration
 * @param self the semantic analyzer instance
 * @param decl the variable decl node to analyze
 */
void analyzeVariableDeclaration(SemanticAnalyzer *self, VariableDecl *decl);

/**
 * Analyze an assignment
 * @param self   the semantic analyzer instance
 * @param assign the assignment node to analyze
 */
void analyzeAssignment(SemanticAnalyzer *self, Assignment *assign);

/**
 * Analyze a decl parent node
 * @param self the semantic analyzer instance
 * @param decl the parent decl node to analyze
 */
void analyzeDeclaration(SemanticAnalyzer *self, Declaration *decl);

/**
 * Analyze an impl declaration
 * @param self the semantic analyzer instance
 * @param impl the impl declaration
 */
void analyzeImplDeclaration(SemanticAnalyzer *self, ImplDecl *impl);

/**
 * Analyze a function call
 * @param self the semantic analyzer instance
 * @param call the function call node to analyze
 */
void analyzeFunctionCall(SemanticAnalyzer *self, Call *call);

/**
 * Analyze a binary expression
 * @param self the semantic analyzer instance
 * @param expr the binary expr to analyze
 */
void analyzeBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr);

/**
 * Analyze a unary expression
 * @param self the semantic analyzer instance
 * @param expr the unary expr to analyze
 */
void analyzeUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr);

/**
 * Analyze an expression
 * @param self the semantic analyzer instance
 * @param expr the expression to analyze
 */
void analyzeExpression(SemanticAnalyzer *self, Expression *expr);

/**
 * analyze an unstructured statement
 * @param self         the semantic analyzer instance
 * @param unstructured the unstructured statement to analyze
 */
void analyzeUnstructuredStatement(SemanticAnalyzer *self, UnstructuredStatement *unstructured);

/**
 * Analyze a structured statement
 * @param self       the semantic analyzer instance
 * @param structured the structured statement to analyze
 */
void analyzeStructuredStatement(SemanticAnalyzer *self, StructuredStatement *structured);

/**
 * Analyze an if statement
 * @param self	 the semantic analyzer instance
 * @param ifStmt the if statement
 */
void analyzeIfStatement(SemanticAnalyzer *self, IfStat *ifStmt);

/**
 * Analyze an else if statement
 * @param self		 the semantic analyzer instance
 * @param elseIfStmt the else if statement
 */
void analyzeElseIfStatement(SemanticAnalyzer *self, ElseIfStat *elseIfStmt);

/**
 * Analyze an else statement
 * @param self	   the semantic analyzer instance
 * @param elseStmt the else statement
 */
void analyzeElseStatement(SemanticAnalyzer *self, ElseStat *elseStmt);

/**
 * Analyze a top level statement
 * @param self the semantic analyzer instance
 * @param stmt the top level statement node
 */
void analyzeStatement(SemanticAnalyzer *self, Statement *stmt);

/**
 * Checks that the main functions is defined, will also
 * check the variables given, etc.
 * @param self the semantic analyzer instance
 */
void checkMainExists(SemanticAnalyzer *self);

/**
 * Starts semantically analyzing everything, will scan through
 * all the files, and their trees etc
 * @param self the semantic analyzer instance
 */
void startSemanticAnalysis(SemanticAnalyzer *self);

/**
 * Checks if a structure exists globally or locally
 * @param  self       the semantic analyzer instance
 * @param  structName the structure name to lookup
 * @return            the structure if it exists
 */
StructDecl *checkStructExists(SemanticAnalyzer *self, char *structName);

/**
 * Check if a variable exists globally or locally
 * @param  self    the semantic analyzer instance
 * @param  varName the variable name to lookup
 * @return         the variable if it exists
 */
VariableDecl *checkVariableExists(SemanticAnalyzer *self, char *varName);

/**
 * Checks if a structure exists locally
 * @param  self       the semantic analyzer instance
 * @param  structName the structure name to lookup
 * @return            the structure if it exists
 */
StructDecl *checkLocalStructExists(SemanticAnalyzer *self, char *structName);

/**
 * Check if a variable exists locally
 * @param  self    the semantic analyzer instance
 * @param  varName the variable name to lookup
 * @return         the variable if it exists
 */
VariableDecl *checkLocalVariableExists(SemanticAnalyzer *self, char *varName);

/**
 * Checks if a structure exists globally
 * @param  self       the semantic analyzer instance
 * @param  structName the structure name to lookup
 * @return            the structure if it exists
 */
StructDecl *checkGlobalStructExists(SemanticAnalyzer *self, char *structName);

/**
 * Check if a variable exists globally
 * @param  self    the semantic analyzer instance
 * @param  varName the variable name to lookup
 * @return         the variable if it exists
 */
VariableDecl *checkGlobalVariableExists(SemanticAnalyzer *self, char *varName);

/**
 * Check if a function exists globally
 * @param  self    the semantic analyzer instance
 * @param  varName the function name to lookup
 * @return         the function if it exists
 */
FunctionDecl *checkFunctionExists(SemanticAnalyzer *self, char *funcName);

ParameterSection *checkLocalParameterExists(SemanticAnalyzer *self, char *paramName);

ImplDecl *checkImplExists(SemanticAnalyzer *self, char *implName);

void pushParameterSection(SemanticAnalyzer *self, ParameterSection *param);

void pushVariableDeclaration(SemanticAnalyzer *self, VariableDecl *var);

void pushStructDeclaration(SemanticAnalyzer *self, StructDecl *structure);

void pushFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *func);

void pushImplDeclaration(SemanticAnalyzer *self, ImplDecl *impl);

/**
 * Push a new scope to the scope stack
 * @param self  the semantic analyzer instance
 */
void pushScope(SemanticAnalyzer *self);

/**
 * Pop a scope from the scope stack
 * @param self the semantic analyzer instance
 */
void popScope(SemanticAnalyzer *self);

/// TYPE CHECKING

VarType *deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call);

VarType *deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit);

VarType *deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr);

VarType *deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr);

VariableType deduceTypeFromExpression(SemanticAnalyzer *self, Expression *expr);

/**
 * Destroy the semantic analyzer instance
 * @param self the semantic analyzer instance to destroy
 */
void destroySemanticAnalyzer(SemanticAnalyzer *self);

#endif // __SEMANTIC_H
