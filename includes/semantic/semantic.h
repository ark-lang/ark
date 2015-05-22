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

	/** for emulating scope */
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
 * Represents a variable type
 */
typedef struct {
	int type;
	bool isArray;
	int arrayLen;
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

/**
 * Create a variable type on the heap
 * @param  type the type to allocate
 * @return      the allocated type
 */
VarType *createVarType(int type);

/**
 * Destroy the given variable type
 * note this isnt called at all, lol 
 * @param type the type to destroy
 */
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
 * Check if an impl exists globally
 * @param  self    the semantic analyzer instance
 * @param  varName the impl name to lookup
 * @return         the impl if it exists
 */
ImplDecl *checkGlobalImplExists(SemanticAnalyzer *self, char *implName);

/**
 * Check if a function exists globally
 * @param  self    the semantic analyzer instance
 * @param  varName the function name to lookup
 * @return         the function if it exists
 */
FunctionDecl *checkFunctionExists(SemanticAnalyzer *self, char *funcName);

/**
 * Check if a local parameter exists
 * @param  self      the semantic analyzer instance
 * @param  paramName the parameter name to lookup
 * @return           the paramter section if it exists, or NULL (false)
 */
ParameterSection *checkLocalParameterExists(SemanticAnalyzer *self, char *paramName);

/**
 * Check if a datatype exists
 * @param self the semantic analyzer instance
 * @param name the datatype name
 */
Type *checkDataTypeExists(SemanticAnalyzer *self, char *name);

/**
 * Pushes a parameter to the scope
 * @param self  the semantic analyzer instance
 * @param param the param to push
 */
void pushParameterSection(SemanticAnalyzer *self, ParameterSection *param);

/**
 * Pushes a variable declaration
 * @param self the semantic analyzer instance
 * @param var  the variable to push
 */
void pushVariableDeclaration(SemanticAnalyzer *self, VariableDecl *var);

/**
 * Pushes a structure declaration
 * @param self the semantic analyzer instance
 * @param structure  the structure to push
 */
void pushStructDeclaration(SemanticAnalyzer *self, StructDecl *structure);

/**
 * Pushes a function declaration
 * @param self the semantic analyzer instance
 * @param func  the function to push
 */
void pushFunctionDeclaration(SemanticAnalyzer *self, FunctionDecl *func);

/**
 * Pushes an implementation declaration
 * @param self the semantic analyzer instance
 * @param impl  the impl to push
 */
void pushImplDeclaration(SemanticAnalyzer *self, ImplDecl *impl);

/** TYPE INFERENCE DEDUCTION STUFF */

/**
 * Deduces a type from the given call
 * @param  self the semantic analyzer instance
 * @param  call the call to deduce
 * @return      the var type we deduced the call to be
 */
VarType *deduceTypeFromFunctionCall(SemanticAnalyzer *self, Call *call);

/**
 * Deduces a type from the given type val
 * @param  self the semantic analyzer instance
 * @param  typeVal the type val to deduce
 * @return      the type val type val we deduced the val type to be
 */
VarType *deduceTypeFromTypeVal(SemanticAnalyzer *self, char *typeVal);

/**
 * Deduces a type from the given lit
 * @param  self the semantic analyzer instance
 * @param  lit the lit to deduce
 * @return      the var type we deduced the lit to be
 */
VarType *deduceTypeFromLiteral(SemanticAnalyzer *self, Literal *lit);

/**
 * Deduces a type from the given binary expression
 * @param  self the semantic analyzer instance
 * @param  expr the binary expression to deduce
 * @return      the var type we deduced the binary expression to be
 */
VarType *deduceTypeFromBinaryExpr(SemanticAnalyzer *self, BinaryExpr *expr);

/**
 * Deduces a type from the given type
 * @param  self the semantic analyzer instance
 * @param  type the type to deduce
 * @return      the var type we deduced the type to be
 */
VarType *deduceTypeFromType(SemanticAnalyzer *self, Type *type);

/**
 * Deduces a type from the given unary expression
 * @param  self the semantic analyzer instance
 * @param  expr the unary expression to deduce
 * @return      the var type we deduced the unary expression to be
 */
VarType *deduceTypeFromUnaryExpr(SemanticAnalyzer *self, UnaryExpr *expr);

/**
 * Deduces a type from the given type
 * @param  self the semantic analyzer instance
 * @param  type the type to deduce
 * @return      the var type we deduced the type to be
 */
VarType *deduceTypeFromTypeNode(SemanticAnalyzer *self, TypeName *type);

/**
 * Deduces a type from the given expression
 * @param  self the semantic analyzer instance
 * @param  expr the expression to deduce
 * @return      the var type we deduced the expression to be
 */
VarType *deduceTypeFromExpression(SemanticAnalyzer *self, Expression *expr);

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

/**
 * Destroy the semantic analyzer instance
 * @param self the semantic analyzer instance to destroy
 */
void destroySemanticAnalyzer(SemanticAnalyzer *self);

#endif // __SEMANTIC_H
