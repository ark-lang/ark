#include "compiler.h"

Compiler *createCompiler() {
	Compiler *self = safeMalloc(sizeof(*self));
	self->abstractSyntaxTree = NULL;
	self->currentNode = 0;

	self->sourceFileSize = 128;
	self->sourcePosition = 0;
	self->sourceName = "test.c";
	self->sourceContents = malloc(sizeof(char) * (self->sourceFileSize + 1));
	self->sourceContents[0] = '\0';
	self->timer = clock();

	return self;
}

void emitExpression(Compiler *self, ExpressionAstNode *expr) {
	// TODO: emit expressions
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

void write_file(Compiler *self) {
	
}

void startCompiler(Compiler *self, Vector *ast) {
	self->abstractSyntaxTree = ast;

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
				break;
		}
	}

	printf("%s\n", self->sourceContents);
	write_file(self);
}

void destroyCompiler(Compiler *self) {
	free(self);
}
