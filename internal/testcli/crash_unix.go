//go:build linux || darwin || freebsd

package main

/*
#include <pthread.h>
#include <stddef.h>

static void *write_to_null(void *arg) {
	*(volatile int *)arg = 1;
	return NULL;
}

static void crash_in_c_thread(void) {
	pthread_t thread;
	if (pthread_create(&thread, NULL, write_to_null, NULL) == 0) {
		pthread_join(thread, NULL);
	}
}
*/
import "C"

func crashInCThread() {
	C.crash_in_c_thread()
}

func crashInGoThread() {
	C.write_to_null(nil)
}
