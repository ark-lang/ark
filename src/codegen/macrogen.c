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

void emitMacroNode(CodeGenerator *self, Macro *macro) {
	switch (macro->type) {
		case USE_MACRO_NODE: emitUseMacro(self, macro->use); break;
	}
}

void generateMacros(CodeGenerator *self) {
	for (int i = 0; i < self->currentSourceFile->ast->size; i++) {
		Statement *stmt = getVectorItem(self->currentSourceFile->ast, i);

		switch (stmt->type) {
			case MACRO_NODE: 
				emitMacroNode(self, stmt->macro);
				break;
		}
	}
}