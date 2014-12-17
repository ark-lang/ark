#include "compiler.h"

Compiler *createCompiler() {
	Compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	self->currentInstruction = 0;
	self->maxBytecodeSize = 32;
	self->bytecode = malloc(sizeof(*self->bytecode) * self->maxBytecodeSize);
	self->currentNode = 0;
	self->globalCount = -1;
	self->mainEntryPoint = -1;
	self->functions = createHashmap(128);
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
		appendInstruction(self, ICONST);
		appendInstruction(self, left);
		appendInstruction(self, ICONST);
		appendInstruction(self, right);

		switch (expr->operand) {
			case '+': appendInstruction(self, ADD); break;
			case '-': appendInstruction(self, SUB); break;
			case '*': appendInstruction(self, MUL); break;
			case '/': appendInstruction(self, DIV); break;
			case '%': appendInstruction(self, MOD); break;
			case '^': appendInstruction(self, POW); break;
		}
	}
	return result;
}

void generateVariableDeclarationCode(Compiler *self, VariableDeclareNode *vdn) {
	DataType type = vdn->vdn->type;
	Token *name = vdn->vdn->name;
	ExpressionNode *expr = vdn->expr;
	evaluateExpressionNode(self, expr);
	consumeNode(self);
}

void generateFunctionCode(Compiler *self, FunctionNode *func) {
	// easier than a shit ton of arrows
	BlockNode *body = func->body;
	Vector *ret = func->ret;
	Vector *args = func->fpn->args;
	Token *name = func->fpn->name;

	// add function name and address to the table
	int functionAddress = self->currentInstruction;
	setValueAtKey(self->functions, name->content, &functionAddress, sizeof(functionAddress));

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
		case FUNCTION_NODE:
			generateFunctionCode(self, currentNode->data);
			break;
		default:
			printf("unrecognized node\n");
			break;
		}
	}

	int *value = getValueAtKey(self->functions, "main");
	if (value) {
		self->mainEntryPoint = *value;
	}

	// stop
	appendInstruction(self, HALT);
}

void destroyCompiler(Compiler *self) {
	if (self != NULL) {
		free(self);
		self = NULL;
	}
}
