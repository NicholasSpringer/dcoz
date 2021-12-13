#ifndef PAUSE_H
#define PAUSE_H

#include <time.h>

int virtual_speedup(int n_cores, unsigned long duration, int prio,
                    pid_t targets[], int n_targets);

#endif