#include "compiler.h"

/** List of data types */
static const char* DATA_TYPES[] = {
	"int", "str", "double", "float", "bool",
	"void", "char", "tup"
};

static const char* INSTRUCTION_SET[] = {
	"ADD", "SUB", "MUL", "RET", "ICONST", "LOAD", "GLOAD", "STORE", "GSTORE", "POP", "HALT"
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

int evaluateExpressionNode(Compiler *self, ExpressionNode *expr) {
	int result = 0;
	if (expr->value != NULL) {
		result = atoi(expr->value->content);
	}
	else {
		int left = evaluateExpressionNode(self, expr->lhand);
		int right = evaluateExpressionNode(self, expr->rhand);

		switch (expr->operand) {
			case '+': result = left + right; break;
			case '-': result = left - right; break;
			case '*': result = left * right; break;
			case '/': result = left / right; break;
			case '%': result = left % right; break;
			case '>': result = left > right; break;
			case '<': result = left < right; break;
			case '^': result = left ^ right; break;
		}
	}
	return result;
}

void generateVariableDeclarationCode(Compiler *self, VariableDeclareNode *vdn) {
	DataType type = vdn->vdn->type;
	Token *name = vdn->vdn->name;
	ExpressionNode *expr = vdn->expr;
	int result = evaluateExpressionNode(self, expr);

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
	for (i = 0; i < self->currentInstruction; i += 2) {
		printf("%s %d\n", INSTRUCTION_SET[self->bytecode[i]], self->bytecode[i + 1]);
	}
}

void destroyCompiler(Compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
