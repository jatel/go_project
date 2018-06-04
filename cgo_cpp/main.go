package main

// #cgo CFLAGS: -I./object
// #cgo LDFLAGS: -L./object -lstdc++ -ltest
// #include "test.h"
import "C"

func main() {
	C.wht_print()
}
