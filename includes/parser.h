#ifndef PARSER_H
#define PARSER_H

// Class

typedef struct tree {
  char* val;
  tree* right;
  tree* left;
} Tree;

void insert(Tree ** tr, int value) {
  Tree *temp = NULL;

  if(!(*tr)) {
    temp = malloc(sizeof(Tree));
    temp->left = temp->right = NULL;
    temp->data = val;
    *tr = temp;
    return;
  }

  if(value < (*tr)->data)
    insert(&(*tree)->left, value);
  else if(value > (*tree)->data)
    insert(&(*tree)->right, value);
  }
}


void deleteTree(Tree ** tr);

#ifndef PARSER_H
