// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package perf

import (
	"fmt"
	. "logger"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	rpprof "runtime/pprof"
	"strings"
	"time"
)

var (
	pid      int
	progname string
	// 监听端口
	pprof_listener net.Listener
)

func init() {
	pid = os.Getpid()
	paths := strings.Split(os.Args[0], "/")
	paths = strings.Split(paths[len(paths)-1], string(os.PathSeparator))
	progname = paths[len(paths)-1]
}

// StartPprof start http pprof.
func StartPprof(addr string) error {

	if nil != pprof_listener {
		return fmt.Errorf("listening: %v", pprof_listener.Addr())
	}

	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofServeMux.HandleFunc("/pid", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "PID: %d", pid)
	})

	var err error

	// 监听端口
	if pprof_listener, err = net.Listen("tcp", addr); err != nil {
		ERROR("StartPprof: %s", err.Error())
		return err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				LOG_ERROR("StartPprof panic: %v", err)
			}
		}()

		if err := http.Serve(pprof_listener, pprofServeMux); err != nil {
			if nil != pprof_listener {
				pprof_listener.Close()
				pprof_listener = nil
			}
		}
	}()

	INFO("pprof: http://%s/debug/pprof/", addr)

	// gc.
	runtime.GC()
	return nil
}

func StopPprof() {

	defer func() {
		if err := recover(); err != nil {
			LOG_ERROR("StopPprof: panic: %v", err)
		}
	}()

	if nil != pprof_listener {
		INFO("stop pprof.")
		pprof_listener.Close()
		pprof_listener = nil
	}
}

func SaveHeapProfile() {
	runtime.GC()

	f, err := os.Create(fmt.Sprintf("./heap_%s_%d_%s.prof", progname, pid, time.Now().Format("2006_01_02_03_04_05")))
	if err != nil {
		return
	}
	defer f.Close()
	rpprof.Lookup("heap").WriteTo(f, 1)
}

func StartCPUProfile() func() {

	f, err := os.Create(fmt.Sprintf("./cpu_%s_%d_%s.prof", progname, pid, time.Now().Format("2006_01_02_03_04_05")))
	if err != nil {
		return func() {}
	}

	rpprof.StartCPUProfile(f)

	return func() {
		rpprof.StopCPUProfile()
		f.Close()
	}
}
