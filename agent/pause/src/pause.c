#define _GNU_SOURCE
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/syscall.h>
#include <time.h>
#include <unistd.h>

/*
 * NOTE: Must set 'sudo sysctl -w kernel.sched_rt_runtime_us=1000000'
 * in order to allow obstructor processes to take up 100% of CPU time
 */

// This should be in sched.h, but is not
struct sched_attr {
    unsigned int size;
    unsigned int sched_policy;
    unsigned long sched_flags;
    int sched_nice;
    unsigned int sched_priority;
    unsigned long sched_runtime;
    unsigned long sched_deadline;
    unsigned long sched_period;
};

// Config stored to restore state after pause
typedef struct sched_config {
    int policy;
    int priority;
} sched_config_t;

// Obstructs a CPU for the given duration by setting own scheduling policy
// to round robin (FIFO with time quantum) with given priority
int obstruct(int32_t prio, time_t duration) {
    time_t start;
    time_t now;
    time(&start);
    do {
        time(&now);
    } while ((now - start) * 1000 < duration);
    return 0;
}

// Pause all processes except for those with given process ids
int virtual_speedup(int n_cores, time_t duration, int prio, pid_t targets[],
                    int n_targets) {
    sched_config_t *configs = malloc(sizeof(sched_config_t) * n_targets);
    for (int i = 0; i < n_targets; i++) {
        pid_t pid = targets[i];
        // Save current policy
        int cur_policy = sched_getscheduler(pid);
        if (cur_policy == -1) {
            free(configs);
            perror("sched_getscheduler error");
            return -1;
        }
        configs[i].policy = cur_policy;

        // Save current priority
        char cur_sched_attr[48];
        int err = syscall(SYS_sched_getattr, pid, &cur_sched_attr,
                          sizeof(cur_sched_attr), 0);
        if (err != 0) {
            free(configs);
            perror("sched_getattr error");
            return -1;
        }
        configs[i].priority =
            ((struct sched_attr *)cur_sched_attr)->sched_priority;

        // Set to temp policy and priority
        struct sched_param temp_params;
        memset(&temp_params, 0, sizeof(temp_params));
        temp_params.sched_priority = prio + 1;
        err = sched_setscheduler(pid, SCHED_RR, &temp_params);
        if (err != 0) {
            perror("sched_setscheduler error");
            free(configs);
            return -1;
        }
    }

    // Start obstructor processes
    for (int i = 0; i < n_cores; i++) {
        int f = fork();
        if (f == -1) {
            perror("sched_setscheduler error");
            free(configs);
            return -1;
        } else if (f == 0) {
            obstruct(prio, duration);
            exit(0);
        } else {
            struct sched_param c;
            c.sched_priority = prio;
            int err = sched_setscheduler(f, SCHED_RR, &c);
            if (err != 0) {
                perror("sched_setscheduler error");
                return 1;
            }
        }
    }
    // Agent is lower priority than obstructors, so this will not run until
    // after obstructors finish executing
    for (int i = 0; i < n_targets; i++) {
        pid_t pid = targets[i];
        // Restore old policy and priority
        struct sched_param old_params;
        memset(&old_params, 0, sizeof(old_params));
        old_params.sched_priority = configs[i].priority;
        int err = sched_setscheduler(pid, configs[i].policy, &old_params);
        if (err != 0) {
            free(configs);
            perror("sched_setscheduler error");
            return 1;
        }
    }
    free(configs);
    return 0;
}
