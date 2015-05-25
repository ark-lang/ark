#include "arguments.h"

Argument *createArgument(char *argName, char *argDescription, size_t arguments, void (*action)(void)) {
    Argument *arg = safeMalloc(sizeof(*arg));
    arg->argName = argName;
    arg->argDescription = argDescription;
    arg->action = action;
    arg->arguments = arguments;
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

void help() {
    // eventually this should iterate through the hasmap
    // and print out all of the help description shit
    printf("Usage:\n\n  ark command [arguments]\n\n");
    printf("Commands:\n\n");
    printf("  help                      Shows this help menu\n");
    printf("  verbose                   Verbose compilation\n");
    printf("  debug                     Logs extra debug information\n");
    printf("  out <file>                Place the output into <file>\n");
    printf("  version                   Shows current version\n");
    printf("  compiler <name>           Sets the C compiler to <name> (default: cc)\n");
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
}

Vector *setup_arguments(int argc, char** argv) {
    arguments = hashmap_new();
    Vector *result = createVector(VECTOR_EXPONENTIAL);

    arg_count = argc;
    arg_value = argv;

    // memory leaks will fix when its not 4am in the morning
    hashmap_put(arguments, "help", createArgument("help", "Shows this help menu", 0, &help));
    hashmap_put(arguments, "verbose", createArgument("verbose", "Verbose compilation", 0, &verbose));
    hashmap_put(arguments, "debug", createArgument("debug", "Logs extra debug information", 0, &debug));
    hashmap_put(arguments, "out", createArgument("out", "Place the output into <file>", 1, &out));
    hashmap_put(arguments, "version", createArgument("version", "Shows current version", 0, &version));
    hashmap_put(arguments, "compiler", createArgument("compiler", "Sets the C compiler to <name> (default: cc)", 1, &compiler));

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

    return result;
}