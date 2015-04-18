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
 * Create a new Semantic Analysis instance
 */
SemanticAnalyzer *createSemanticAnalyzer();

/**
 * Analyzes an unstructured statement top level node
 */
void analyzeUnstructuredStatement(SemanticAnalyzer *self, UnstructuredStatement *unstructured);

/**
 * Analyzes a structured statement top level node
 */
void analyzeStructuredStatement(SemanticAnalyzer *self, StructuredStatement *unstructured);

/**
 * Analyzes a statement top level node
 */
void analyzeStatement(SemanticAnalyzer *self, Statement *stmt);

/**
 * Start semantically analyzing the source files. 
 *
 * NOTE TO SELF FELIKS:
 * I have a feeling the order of source files being analyzed will matter
 * especially if we're doing this recursively... so I should probably think
 * about this.
 */
void startSemanticAnalysis(SemanticAnalyzer *self);

/**
 * Destroys the semantic analysis instance given
 */
void destroySemanticAnalyzer(SemanticAnalyzer *self);

#endif // __SEMANTIC_H