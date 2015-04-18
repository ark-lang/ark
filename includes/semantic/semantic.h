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

#define MAIN_FUNC "main"

typedef struct {
	/** The current AST being analyzed */
	Vector *abstractSyntaxTree;

	/**  */
	SourceFile *currentSourceFile;

	/** the source files to semantically analyze */
	Vector *sourceFiles;

	/** the current node in the ast */
	int currentNode;

	/** hashmap for functions defined */
	map_t funcSymTable;

	/** hashmap for variables defined */
	map_t varSymTable;

	/** if this stage failed or not */
	bool failed;
} SemanticAnalyzer;

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
 * Analyze a function call
 * @param self the semantic analyzer instance
 * @param call the function call node to analyze
 */
void analyzeFunctionCall(SemanticAnalyzer *self, Call *call);

/**
 * Analyze a literal
 * @param self the semantic analyzer instance
 * @param lit  the literal node to analyze
 */
void analyzeLiteral(SemanticAnalyzer *self, Literal *lit);

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
 * Destroy the semantic analyzer instance
 * @param self the semantic analyzer instance to destroy
 */
void destroySemanticAnalyzer(SemanticAnalyzer *self);

#endif // __SEMANTIC_H