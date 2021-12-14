#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "pause.h"

// Usage: ./pause n_cores duration(ms) priority n_targets target1 target2...
int main(int argc, char *argv[]) {
    int n_cores = atoi(argv[1]);
    time_t duration = (time_t)atoi(argv[2]);
    int priority = atoi(argv[3]);
    int n_targets = atoi(argv[4]);

    int targets[1024];
    for (int i = 0; i < n_targets; i++) {
        targets[i] = atoi(argv[5 + i]);
    }

    return virtual_speedup(n_cores, duration, priority, targets, n_targets);
}