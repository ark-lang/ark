#include "codegen.h"

void emitUseMacro(CodeGenerator *self, UseMacro *use) {
	// cuts out the extension from the file
	// name of the current file.
	size_t len = strlen(use->file);
	char temp[len - 2];
	memcpy(temp, &use->file[1], len - 2);
	temp[len - 2] = '\0';

	emitCode(self, "#include \"_gen_%s.h\"\n", temp);
}

void emitLinkerFlagMacro(CodeGenerator *self, LinkerFlagMacro *linker) {
	ALLOY_UNUSED_OBJ(self);

	size_t len = strlen(linker->flag);
	char temp[len - 2];
	memcpy(temp, &linker->flag[1], len - 2);
	temp[len - 2] = '\0';

	self->LINKER_FLAGS = sdscat(self->LINKER_FLAGS, " -l");
	self->LINKER_FLAGS = sdscat(self->LINKER_FLAGS, temp);
	self->LINKER_FLAGS = sdscat(self->LINKER_FLAGS, " ");
}

void emitMacroNode(CodeGenerator *self, Macro *macro) {
	switch (macro->type) {
		case USE_MACRO_NODE: emitUseMacro(self, macro->use); break;
		case LINKER_FLAG_MACRO_NODE: emitLinkerFlagMacro(self, macro->linker); break;
	}
}

void generateMacros(CodeGenerator *self) {
	for (int i = 0; i < self->currentSourceFile->ast->size; i++) {
		Statement *stmt = getVectorItem(self->currentSourceFile->ast, i);

		if (stmt->type == MACRO_NODE) {
			emitMacroNode(self, stmt->macro);
		}
	}
}
