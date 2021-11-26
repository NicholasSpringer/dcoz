#include "pause.h"
#include <stdlib.h>
#include <stdio.h>

int N_CORES = 8;
time_t DURATION = 5;
int PRIORITY = 4;

int main(int argc, char *argv[]) {
    int targets[] = {atoi(argv[1])};
    virtual_speedup(N_CORES, PRIORITY, DURATION, targets, 1);
    return 0;
}