#define _GNU_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <linux/limits.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

const int N_NAMESPACES = 5;
const static struct ns_file {
    int ns_type;
    char *ns_path;
} NS_FILES[] = {
    {.ns_type = CLONE_NEWCGROUP, .ns_path = "ns/cgroup"},
    {
        .ns_type = CLONE_NEWIPC,
        .ns_path = "ns/ipc",
    },
    {
        .ns_type = CLONE_NEWUTS,
        .ns_path = "ns/uts",
    },
    {
        .ns_type = CLONE_NEWNET,
        .ns_path = "ns/net",
    },
    {
        .ns_type = CLONE_NEWPID,
        .ns_path = "ns/pid",
    },
    {
        .ns_type = CLONE_NEWNS,
        .ns_path = "ns/mnt",
    },
};

void enter_target_ns(char *target_ns_proc_path) {
    int ns_fds[N_NAMESPACES];
    char path_buf[PATH_MAX];
    for (int ns_idx = 0; ns_idx < N_NAMESPACES; ns_idx++) {
        char *ns_path = NS_FILES[ns_idx].ns_path;
        snprintf(path_buf, sizeof(path_buf), "%s/%s", target_ns_proc_path,
                 ns_path);
        ns_fds[ns_idx] = open(path_buf, O_RDONLY);
        if (ns_fds[ns_idx] == -1) {
            char buf[1024];
            snprintf(buf, sizeof(buf), "error opening ns file (%s)", ns_path);
            perror(buf);
            exit(1);
        }
    }
    for (int ns_idx = 0; ns_idx < N_NAMESPACES; ns_idx++) {
        char *ns_path = NS_FILES[ns_idx].ns_path;
        int ns_type = NS_FILES[ns_idx].ns_type;
        int err = setns(ns_fds[ns_idx], ns_type);
        if (err == -1) {
            char buf[1024];
            snprintf(buf, sizeof(buf), "error entering ns (%s)", ns_path);
            perror(buf);
            exit(1);
        }
        close(ns_fds[ns_idx]);
    }
}