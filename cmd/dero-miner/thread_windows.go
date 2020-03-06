package main

import "runtime"
import "sync/atomic"
import "syscall"
import "unsafe"
import "math/bits"

var libkernel32 uintptr
var setThreadAffinityMask uintptr

func doLoadLibrary(name string) uintptr {
	lib, _ := syscall.LoadLibrary(name)
	return uintptr(lib)
}

func doGetProcAddress(lib uintptr, name string) uintptr {
	addr, _ := syscall.GetProcAddress(syscall.Handle(lib), name)
	return uintptr(addr)
}

func syscall3(trap, nargs, a1, a2, a3 uintptr) uintptr {
	ret, _, _ := syscall.Syscall(trap, nargs, a1, a2, a3)
	return ret
}

func init() {
	libkernel32 = doLoadLibrary("kernel32.dll")
	setThreadAffinityMask = doGetProcAddress(libkernel32, "SetThreadAffinityMask")
}

var processor int32

// currently we suppport upto 64 cores
func SetThreadAffinityMask(hThread syscall.Handle, dwThreadAffinityMask uint) *uint32 {
	ret1 := syscall3(setThreadAffinityMask, 2,
		uintptr(hThread),
		uintptr(dwThreadAffinityMask),
		0)
	return (*uint32)(unsafe.Pointer(ret1))
}

// CurrentThread returns the handle for the current thread.
// It is a pseudo handle that does not need to be closed.
func CurrentThread() syscall.Handle { return syscall.Handle(^uintptr(2 - 1)) }

// sets thread affinity to avoid cache collision and thread migration
func threadaffinity() {
	lock_on_cpu := atomic.AddInt32(&processor, 1)
	if lock_on_cpu >= int32(runtime.GOMAXPROCS(0)) { // threads are more than cpu, we do not know what to do
		return
	}

	if lock_on_cpu >= bits.UintSize {
		return
	}
	var cpuset uint
	cpuset = 1 << uint(avoidHT(int(lock_on_cpu)))
	SetThreadAffinityMask(CurrentThread(), cpuset)
}

func avoidHT(i int) int {
	count := runtime.GOMAXPROCS(0)
	if i < count/2 {
		return i * 2
	} else {
		return (i-count/2)*2 + 1
	}
}
