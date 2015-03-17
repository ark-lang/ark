#include "compiler.h"

static const char* NODE_TYPE[] = {
	"EXPRESSION_AST_NODE", "VARIABLE_DEF_AST_NODE",
	"VARIABLE_DEC_AST_NODE", "FUNCTION_ARG_AST_NODE",
	"FUNCTION_AST_NODE", "FUNCTION_PROT_AST_NODE",
	"BLOCK_AST_NODE", "FUNCTION_CALLEE_AST_NODE",
	"FUNCTION_RET_AST_NODE", "FOR_LOOP_AST_NODE",
	"VARIABLE_REASSIGN_AST_NODE", "INFINITE_LOOP_AST_NODE",
	"BREAK_AST_NODE", "CONTINUE_AST_NODE", "ENUM_AST_NODE", "STRUCT_AST_NODE",
	"IF_STATEMENT_AST_NODE", "MATCH_STATEMENT_AST_NODE", "WHILE_LOOP_AST_NODE",
	"ANON_AST_NODE", "USE_STATEMENT_AST_NODE"
};

Compiler *createCompiler(Vector *sourceFiles) {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;
	self->sourceFiles = sourceFiles;
	return self;
}

void emitExpression(Compiler *self, ExpressionAstNode *expr) {

}

void emitVariableDeclaration(Compiler *self, VariableDeclarationAstNode *var) {

}

void emitStructure(Compiler *self, StructureAstNode *structure) {

}

void emitFunctionCall(Compiler *self, FunctionCallAstNode *call) {	

}

void emit_if_statement(Compiler *self, IfStatementAstNode *stmt) {
	
}

void emitBlock(Compiler *self, BlockAstNode *block) {

}

void emitArguments(Compiler *self, Vector *args) {

}

void emitReturnStatement(Compiler *self, FunctionReturnAstNode *ret) {

}

void emitFunction(Compiler *self, FunctionAstNode *func) {

}

void consumeAstNode(Compiler *self) {
	self->currentNode += 1;
}

void consumeAstNodeBy(Compiler *self, int amount) {
	self->currentNode += amount;
}

void startCompiler(Compiler *self) {
	int i;
	for (i = 0; i < self->sourceFiles->size; i++) {
		SourceFile *sourceFile = getVectorItem(self->sourceFiles, i);
		self->currentSourceFile = sourceFile;
		self->abstractSyntaxTree = self->currentSourceFile->ast;
		compileAST(self);
	}
}

void compileAST(Compiler *self) {
	int i;
	for (i = 0; i < self->abstractSyntaxTree->size; i++) {
		AstNode *currentAstNode = getVectorItem(self->abstractSyntaxTree, i);

		switch (currentAstNode->type) {
			case FUNCTION_AST_NODE:
				emitFunction(self, currentAstNode->data);
				break;
			case VARIABLE_DEC_AST_NODE:
				emitVariableDeclaration(self, currentAstNode->data);
				break;
			case STRUCT_AST_NODE:
				emitStructure(self, currentAstNode->data);
				break;
			default:
				printf("node not yet supported: %s :(\n", NODE_TYPE[currentAstNode->type]);
				break;
		}
	}
}

void destroyCompiler(Compiler *self) {
	free(self);
}
