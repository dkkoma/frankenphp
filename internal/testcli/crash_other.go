//go:build !linux && !darwin && !freebsd

package main

func crashInCThread() {}

func crashInGoThread() {}
