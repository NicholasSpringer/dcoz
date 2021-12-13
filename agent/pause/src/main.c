#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

//#include "ns.h"
#include "pause.h"

int N_CORES = 8;
time_t DURATION = 5;
int PRIORITY = 4;

// Usage: ./pause target_ns_proc_path n_cores duration(ms) priority n_targets target1 target2...
int main(int argc, char *argv[]) {
    //char *target_ns_proc_path = argv[1];
    int n_cores = atoi(argv[2]);
    time_t duration = (time_t)atoi(argv[3]);
    int priority = atoi(argv[4]);
    int n_targets = atoi(argv[5]);

    

    int targets[1024];
    for (int i=0; i<n_targets; i++) {
        targets[i] = atoi(argv[6 + i]);
    }

    /*enter_target_ns(target_ns_proc_path);

    // We must replace original process with new one to have PID namespace
    // change take effect
    int f = fork();
    if (f == -1) {
        perror("error forking");
        return 1;
    } else if (f != 0) {
        return 0;
    }*/

    
    return virtual_speedup(n_cores, duration, priority, targets, n_targets);
}