package pidfile

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
)

// This function will exit if the pidfile has a
// running process already this is useful to just
// place your process in the crontab and leave it.
// It tries to keep only 1 running at a time
//
func IfRunning(file string, running_callback func(int)) {

	running, _ := IsPidfileRunning(file)
	if running {
		pid, _ := GetPidFileValue(file)
		running_callback(pid)
	} else {

		err := SetPidfile(file)
		if err != nil {
			fmt.Printf("Setting Pid Fialed: %v", err)
			os.Exit(1)
		}
		val, _ := GetPidFileValue(file)

		// I'm sure there are many race conditions this does not catch
		if val != os.Getpid() {
			fmt.Printf("Pid file value cheanged: %v != %v\n", val, os.Getpid())
			os.Exit(1)
		}
	}
}

// checks to see if the pid in the file you passed in is still running
// if the file doesnt exists it will return false
func IsPidfileRunning(file string) (running bool, err error) {

	pid, err := GetPidFileValue(file)
	// some sort of error
	if err != nil {
		return false, err
	}

	// no value with no error just means the file didnt exist
	// so it's not running
	if pid == 0 {
		return false, nil
	}

	// this only works with link currently
	if runtime.GOOS != "linux" {
		panic("pidfile: Your OS doesnt work yet. Sry..")
	}
	_, err = os.Stat(path.Join("/proc", strconv.Itoa(int(pid))))
	if err != nil {
		return false, nil
	} else {
		return true, nil
	}

	panic("Should not Reach")
}

func GetPidFileValue(file string) (pid int, err error) {

	fstat, _ := os.Stat(file)
	if fstat == nil {
		return 0, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("Can not Open (%s): %s", file, err)
	}
	// could it really ever be larger than this?
	buf := make([]byte, 30)
	read, err := f.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("Can not Read (%s): %s", file, err)
	}

	bigPid, err := strconv.ParseInt(string(buf[0:read]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Can not Parse (%s): %s", string(buf[0:read]), err)
	}

	return int(bigPid), nil
}

// Sets a pid file without checking anything
// will over write a file if it exists
func SetPidfile(file string) (err error) {

	pid := os.Getpid()
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("Can not create (%s): %s", file, err)
	}

	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		return fmt.Errorf("Can not write (%s): %s", file, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("Can not close (%s): %s", file, err)
	}

	return nil
}
