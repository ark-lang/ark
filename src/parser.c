#include "parser.h"

void treeInsert(Tree **tr, char *value) {
    Tree *temp = NULL;

    if(!(*tr)) {
        temp = malloc(sizeof(*temp));
        if (!temp) {
        	perror("malloc: failed to allocate memory for tree");
        }
        temp->left = temp->right = NULL;
        temp->data = value;
        *tr = temp;
        return;
    }

    if(value < (*tr)->data)
        treeInsert(&(*tr)->left, value);
    else if(value > (*tr)->data)
        treeInsert(&(*tr)->right, value);
    }
}