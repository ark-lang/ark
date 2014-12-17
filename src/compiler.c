#include "compiler.h"

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void", "char", "tup"
};

Compiler *createCompiler() {
	Compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	self->currentInstruction = 0;
	self->maxBytecodeSize = 32;
	self->bytecode = malloc(sizeof(*self->bytecode) * self->maxBytecodeSize);
	self->currentNode = 0;
	self->globalCount = -1;
	return self;
}

void appendInstruction(Compiler *self, int instr) {
	if (self->currentInstruction >= self->maxBytecodeSize) {
		self->maxBytecodeSize *= 2;
		self->bytecode = realloc(self->bytecode, sizeof(*self->bytecode) * self->maxBytecodeSize);
	}
	self->bytecode[self->currentInstruction++] = instr;
}

void consumeNode(Compiler *self) {
	self->currentNode += 1;
}

void generateVariableDeclarationCode(Compiler *self, VariableDeclareNode *vdn) {
	DataType type = vdn->vdn->type;
	Token *name = vdn->vdn->name;
	ExpressionNode *expr = vdn->expr;

	if (expr->type != EXPR_NUMBER) {
		printf("unsupported expression.\n");
		exit(1);
	}

	int result = atoi(expr->value->content);

	printf("%s %s = %d\n", DATA_TYPES[type], name->content, result);

	appendInstruction(self, ICONST);
	appendInstruction(self, result);
	appendInstruction(self, GSTORE);
	appendInstruction(self, ++self->globalCount);

	consumeNode(self);
}

void startCompiler(Compiler *self, Vector *ast) {
	self->ast = ast;

	while (self->currentNode < self->ast->size) {
		Node *currentNode = getItemFromVector(self->ast, self->currentNode);

		switch (currentNode->type) {
			case VARIABLE_DEC_NODE:
				generateVariableDeclarationCode(self, currentNode->data);
				break;
			default:
				printf("unrecognized node\n");
				break;
		}
	}

	// stop
	appendInstruction(self, HALT);

	int i;
	for (i = 0; i < self->currentInstruction; i++) {
		printf("%d\n", self->bytecode[i]);
	}
}

void destroyCompiler(Compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
