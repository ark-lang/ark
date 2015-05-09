#include "codegen.h"

void emitUseMacro(CodeGenerator *self, UseMacro *use) {
	for (int i = 0; i < use->files->size; i++) {
		char *fileName = getVectorItem(use->files, i);

		size_t len = strlen(fileName);
		char temp[len - 2];
		memcpy(temp, &fileName[1], len - 2);
		temp[len - 2] = '\0';

		emitCode(self, "#include \"_gen_%s.h\"\n", temp);
	}
}

void emitLinkerFlagMacro(CodeGenerator *self, LinkerFlagMacro *linker) {
	ALLOY_UNUSED_OBJ(self);

	size_t len = strlen(linker->flag);
	char temp[len - 2];
	memcpy(temp, &linker->flag[1], len - 2);
	temp[len - 2] = '\0';

	self->linkerFlags = sdscat(self->linkerFlags, " -l");
	self->linkerFlags = sdscat(self->linkerFlags, temp);
	self->linkerFlags = sdscat(self->linkerFlags, " ");
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
