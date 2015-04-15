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

typedef struct {
	/** the AST to semantically analyze */
	Vector *parseTree;

	/** hashmap for functions defined */
	map_t funcSymTable;

	/** hashmap for variables defined */
	map_t varSymTable;
} SemanticAnalysis;

/**
 * Create a new Semantic Analysis instance
 */
SemanticAnalysis *createSemanticAnalysis();

/**
 * Analyzes an unstructured statement top level node
 */
void analyzeUnstructuredStatement(SemanticAnalysis *self, UnstructuredStatement *unstructured);

/**
 * Analyzes a structured statement top level node
 */
void analyzeStructuredStatement(SemanticAnalysis *self, StructuredStatement *unstructured);

/**
 * Analyzes a statement top level node
 */
void analyzeStatement(SemanticAnalysis *self, Statement *stmt);

/**
 * Start semantically analyzing the source files. 
 *
 * NOTE TO SELF FELIKS:
 * I have a feeling the order of source files being analyzed will matter
 * especially if we're doing this recursively... so I should probably think
 * about this.
 */
void startSemanticAnalysis(SemanticAnalysis *self);

/**
 * Destroys the semantic analysis instance given
 */
void destroySemanticAnalysis(SemanticAnalysis *self);

#endif // __SEMANTIC_H