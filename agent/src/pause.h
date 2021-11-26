#ifndef PAUSE_H
#define PAUSE_H

#include <time.h>

int virtual_speedup(int n_cores, int prio, time_t duration, pid_t targets[], int n_targets);

#endif 