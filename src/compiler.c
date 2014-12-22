#include "compiler.h"

Compiler *createCompiler() {
	Compiler *self = malloc(sizeof(*self));
	self->ast = NULL;
	self->currentInstruction = 0;
	self->maxBytecodeSize = 32;
	self->bytecode = malloc(sizeof(*self->bytecode) * self->maxBytecodeSize);
	self->currentNode = 0;
	self->globalCount = -1;
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
	if (expr->value != NULL) {
		appendInstruction(self, ICONST);
		appendInstruction(self, atoi(expr->value->content));
	}
	else {
		evaluateExpressionNode(self, expr->lhand);
		evaluateExpressionNode(self, expr->rhand);

		switch (expr->operand) {
			case '+': appendInstruction(self, ADD); break;
			case '-': appendInstruction(self, SUB); break;
			case '*': appendInstruction(self, MUL); break;
			case '/': appendInstruction(self, DIV); break;
			case '%': appendInstruction(self, MOD); break;
			case '^': appendInstruction(self, POW); break;
		}
	}
	return 1337;
}

void generateVariableDeclarationCode(Compiler *self, VariableDeclareNode *vdn) {
/*
	DataType type = vdn->vdn->type;
	Token *name = vdn->vdn->name;
	ExpressionNode *expr = vdn->expr;
	evaluateExpressionNode(self, expr);
*/	

	// todo

	consumeNode(self);
}

void generateFunctionCalleeCode(Compiler *self, FunctionCalleeNode *fcn) {
	char *name = fcn->callee->content;
	int *address = getValueAtKey(self->functions, name);
	int numOfArgs = fcn->args->size;

	int i;
	for (i = 0; i < numOfArgs; i++) {
		FunctionArgumentNode *fan = getItemFromVector(fcn->args, i);
		evaluateExpressionNode(self, fan->value);
	}

	appendInstruction(self, CALL);
	appendInstruction(self, *address);
	appendInstruction(self, numOfArgs);

	consumeNode(self);
}

void generateFunctionReturnCode(Compiler *self, FunctionReturnNode *frn) {
	if (frn->numOfReturnValues > 1) {
		printf("tuples not yet supported.\n");
		exit(1);
	}
	// no tuple support, just use first return value for now.
	ExpressionNode *expr = getItemFromVector(frn->returnVals, 0);
	evaluateExpressionNode(self, expr);
}

void generateFunctionCode(Compiler *self, FunctionNode *func) {
	int address = self->currentInstruction;
	setValueAtKey(self->functions, func->fpn->name->content, &address, sizeof(int));
		
	Vector *statements = func->body->statements;

	// return stuff
	int i;
	for (i = 0; i < statements->size; i++) {
		StatementNode *sn = getItemFromVector(statements, i);
		switch (sn->type) {
			case FUNCTION_RET_NODE:
				generateFunctionReturnCode(self, sn->data);
				break;
			default:
				printf("WHAT NODES YA GIVIN ME SON?\n");
				break;
		}
	}

	appendInstruction(self, RET);

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
		case FUNCTION_CALLEE_NODE:
			generateFunctionCalleeCode(self, currentNode->data);
			break;
		default:
			printf("unrecognized node\n");
			break;
		}
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
