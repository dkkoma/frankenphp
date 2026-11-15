//go:build linux || darwin || freebsd

package main

/*
#include <pthread.h>
#include <stddef.h>
#include <stdlib.h>

static void *write_to_null(void *arg) {
	*(volatile int *)arg = 1;
	return NULL;
}

static void *call_abort(void *arg) {
	abort();
	return NULL;
}

static void run_in_c_thread(void *(*crash)(void *)) {
	pthread_t thread;
	if (pthread_create(&thread, NULL, crash, NULL) == 0) {
		pthread_join(thread, NULL);
	}
}

static void crash_in_c_thread(void) {
	run_in_c_thread(write_to_null);
}

static void abort_in_c_thread(void) {
	run_in_c_thread(call_abort);
}
*/
import "C"

func crashInCThread() {
	C.crash_in_c_thread()
}

func abortInCThread() {
	C.abort_in_c_thread()
}

func crashInGoThread() {
	C.write_to_null(nil)
}
