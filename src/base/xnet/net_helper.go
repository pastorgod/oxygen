package xnet

import "net"
import "syscall"
import "runtime"

import . "logger"

//var listenFD *int = flag.Int("listenFD", 0, "the already-open fd to listen on (internal use only)")

// 网卡地址列表
func FetchInterfaces() []string {

	ifs, err := net.Interfaces()
	Assert(err == nil, err)

	addrs, err := net.InterfaceAddrs()
	Assert(err == nil, err)

	list := make([]string, 0, len(ifs))

	for index, ifa := range ifs {
		if 0 != (ifa.Flags&net.FlagUp) && 0 == (ifa.Flags&net.FlagBroadcast) {
			list = append(list, addrs[index].String())
		}
	}

	return list
}

func fcntl(fd, cmd, arg int) (val int, err error) {

	if runtime.GOOS != "linux" {
		LOG_FATAL("only for linux.")
	}

	r0, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), uintptr(arg))

	val = int(r0)

	if e1 != 0 {
		err = e1
	}

	return
}

func NoCloseOnExec(fd uintptr) bool {
	if _, err := fcntl(int(fd), syscall.F_SETFD, ^syscall.FD_CLOEXEC); err != nil {
		LOG_ERROR("NoCloseOnExec Fail. %d", int(fd))
		return false
	}

	return true
}
