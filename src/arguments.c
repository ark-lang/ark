#include "arguments.h"

Argument *createArgument(char *argName, char *argDescription, size_t arguments, void (*action)(void), int type) {
    Argument *arg = safeMalloc(sizeof(*arg));
    arg->argName = argName;
    arg->argDescription = argDescription;
    arg->action = action;
    arg->arguments = arguments;
    arg->type = type;
    return arg;
}

void destroyArgument(Argument *arg) {
    free(arg);
}

Vector *eatArguments() {
    Vector *vec = createVector(VECTOR_EXPONENTIAL);
    for (int i = currentArgumentIndex; i < currentArgument->arguments + currentArgumentIndex; i++) {
        char *value = arg_value[i + 1];
        pushBackItem(vec, value);
    }
    return vec;
}

int printCommand(any_t __attribute__((unused)) passedData, any_t item) {
    Argument *arg = item;
    if (arg->type != ARG_COMMAND) return MAP_OK;
    int width = 20;
    printf("  %-*s %s\n", width, boldText(arg->argName), arg->argDescription);
    return MAP_OK;
}

int printOption(any_t __attribute__((unused)) passedData, any_t item) {
    Argument *arg = item;
    if (arg->type != ARG_OPTION) return MAP_OK;
    int width = 20;
    printf("  %-*s %s\n", width, boldText(arg->argName), arg->argDescription);
    return MAP_OK;
}

void help() {
    printf("Usage:\n\n  ark command [arguments]\n\n");

    printf("Commands:\n\n");
    hashmap_iterate(arguments, printCommand, NULL);
    printf("\n");

    printf("Options:\n\n");
    hashmap_iterate(arguments, printOption, NULL);
    printf("\n");
}

void verbose() {
    VERBOSE_MODE = true;
}

void debug() {
    DEBUG_MODE = true;
}

void out() {
    Vector *args = eatArguments();
    char *topItem = getVectorTop(args);
    if (!topItem) {
        errorMessage("Missing argument, expected %d arguments for command `%s`", currentArgument->arguments, currentArgument->argName);
        return;
    }
    OUTPUT_EXECUTABLE_NAME = topItem;
    printf("setting OUTPUT_EXECUTABLE_NAME to %s\n", topItem);
    destroyVector(args);
}

void version() {
    printf("%s %s\n", COMPILER_NAME, COMPILER_VERSION);
}

void compiler() {
    Vector *args = eatArguments();
    char *topItem = getVectorTop(args);
    if (!topItem) {
        errorMessage("Missing argument, expected %d arguments for command `%s`", currentArgument->arguments, currentArgument->argName);
        return;
    }
    printf("setting compiler to %s\n", topItem);
    COMPILER = topItem;
    destroyVector(args);
}

int destroyArguments(any_t __attribute__((unused)) passedData, any_t item) {
    destroyArgument(item);
    return MAP_OK;
}

Vector *setup_arguments(int argc, char** argv) {
    arguments = hashmap_new();
    Vector *result = createVector(VECTOR_EXPONENTIAL);

    arg_count = argc;
    arg_value = argv;

    // memory leaks will fix when its not 4am in the morning
    hashmap_put(arguments, "help", createArgument("help", "Shows this help menu", 0, &help, ARG_COMMAND));
    hashmap_put(arguments, "version", createArgument("version", "Shows current version", 0, &version, ARG_COMMAND));
    hashmap_put(arguments, "-v", createArgument("-v", "Verbose compilation", 0, &verbose, ARG_OPTION));
    hashmap_put(arguments, "-d", createArgument("-d", "Logs extra debug information", 0, &debug, ARG_OPTION));
    hashmap_put(arguments, "-o", createArgument("-o", "Place the output into <file>", 1, &out, ARG_OPTION));
    hashmap_put(arguments, "-ac", createArgument("-ac", "Sets the C compiler to <name> (default: cc)", 1, &compiler, ARG_OPTION));

    // i = 1, ignores first argument
    for (int i = 1; i < argc; i++) {
        if (strstr(argv[i], COMPILER_EXTENSION)) {
            SourceFile *file = createSourceFile(sdsnew(argv[i]));
            if (!file) {
                verboseModeMessage("Error when attempting to create a source file");
                return false;
            }
            pushBackItem(result, file);
        }
        else {
            Argument *arg = NULL;
            if (hashmap_get(arguments, argv[i], (void**) &arg) == MAP_OK) {
                currentArgument = arg;
                currentArgumentIndex = i;
                i += arg->arguments;
                arg->action();
            }
            else {
                errorMessage("unknown argument `%s`\n Run `ark help` for usage.");
            }
        }
    }

    // memory leak here
    hashmap_iterate(arguments, destroyArguments, NULL);
    hashmap_free(arguments);

    return result;
}