#include "pause.h"
#include <stdlib.h>
#include <stdio.h>

int N_CORES = 8;
time_t DURATION = 5;
int PRIORITY = 4;

// Usage: ./pause n_cores duration(ms) priority n_targets target1 target2...
int main(int argc, char *argv[]) {
    int n_cores = atoi(argv[1]);
    time_t duration = (time_t)atoi(argv[2]);
    int priority = atoi(argv[3]);
    int n_targets = atoi(argv[4]);

    int targets[1024];
    for (int i=0; i<n_targets; i++) {
        targets[i] = atoi(argv[5 + i]);
    }
    virtual_speedup(n_cores, duration, priority, targets, n_targets);
    return 0;
}