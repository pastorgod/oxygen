package base

import (
	"bytes"
	"errors"
	"io/ioutil"
	. "logger"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func RunCommand(timeout int, cmdStr string, args ...string) (string, error) {

	defer func() {
		if err := recover(); err != nil {
			LOG_ERROR("RunCommand: %v", err)
		}
	}()

	output := bytes.NewBufferString("")

	cmd := exec.Command(cmdStr, args...)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	LOG_DEBUG("执行命令: %v", cmd.Args)

	runfunc := func() error {

		defer func() {
			if err := recover(); err != nil {
				LOG_ERROR("RunCommand: %v", err)
			}
		}()

		if err := cmd.Run(); err != nil {
			LOG_ERROR("ExcuteCommand Fail. %v", err.Error())
			return err
		}
		/*
			if err := cmd.Wait(); err != nil {
				LOG_ERROR( "Command.Wait: %s", err.Error() )
				return err
			}
		*/
		return nil
	}

	if timeout > 0 {
		var err error
		timeoutchan := make(chan bool, 1)

		go func() {
			err = runfunc()
			timeoutchan <- true
		}()

		select {
		case <-time.After(time.Second * time.Duration(timeout)):
			cmd.Process.Kill()
			err = errors.New("RunCommand Timeout")
			break
		case <-timeoutchan:
			break
		}

		return output.String(), err
	}

	err := runfunc()
	return output.String(), err
}

// 异步执行命令
func AsynRunCommand(callback func(string, error), timeout int, cmdStr string, args ...string) {

	go func() {
		output, err := RunCommand(timeout, cmdStr, args...)
		callback(output, err)
	}()
}

// 解压缩文件
func ExtractTo(tarFile, folder string) error {

	if err := os.MkdirAll(folder, 0777); err != nil {

		if !os.IsExist(err) {
			LOG_ERROR("创建文件夹失败! %s", err.Error())
			return err
		}
	}

	_, err := RunCommand(-1, "tar", "-xvf", tarFile, "-C", folder)
	return err
}

// ./rsync_script.sh src user ip dst
func RsyncScript(script, src, deploy_addr string) error {

	u, err := url.Parse(deploy_addr)

	if err != nil {
		return err
	}

	user := u.User.Username()

	ip, _, err := net.SplitHostPort(u.Host)

	if err != nil {
		return err
	}

	dest := u.Path

	if dest != "" && (!strings.HasPrefix(dest, ".") || !strings.HasPrefix(dest, "~")) {
		dest = dest[1:]
	}

	if "" == dest {
		dest = "."
	}

	output, err := RunCommand(-1, "/bin/sh", script, src, user, ip, dest)
	LOG_DEBUG("RsyncScript: %s %s %s, [%s, %v]", script, src, deploy_addr, output, err)
	return err
}

// filename ssh://publish@192.168.1.2:22/./bin/
func Rsync(src, deploy_addr string) error {

	u, err := url.Parse(deploy_addr)

	if err != nil {
		return err
	}

	user := u.User.Username()

	ip, _, err := net.SplitHostPort(u.Host)

	if err != nil {
		return err
	}

	dest := u.Path

	if dest != "" && (!strings.HasPrefix(dest, ".") || !strings.HasPrefix(dest, "~")) {
		dest = dest[1:]
	}

	if "" == dest {
		dest = "."
	}

	// rsync -auve 'ssh -i /home/red/.ssh/cron_jobs_key' red@othermachine:/source/dir1 /dest/dir2
	cmd := exec.Command("rsync", "-avzque", "ssh", src, Sprintf("%s@%s:%s", user, ip, dest))

	INFO("Deploy: %v", cmd.Args)

	return cmd.Run()
}

// 异步解压缩
func AsynExtractTo(tarFile, folder string, callback func(error)) {

	go func() {
		err := ExtractTo(tarFile, folder)
		callback(err)
	}()
}

// 发送信号
func Signal(pid int, sig os.Signal) bool {

	process, err := os.FindProcess(pid)

	if nil != err {
		LOG_ERROR("find process error: %d", err.Error())
		return false
	}

	if err := process.Signal(sig); err != nil {
		LOG_ERROR("send signal error %s", err.Error())
		return false
	}

	return true
}

// 杀死指定进程
func Kill(pid int) bool {

	process, err := os.FindProcess(pid)

	if nil != err {
		LOG_ERROR("find process error: %d", err.Error())
		return false
	}

	if err := process.Kill(); err != nil {
		LOG_ERROR("kill process error %s", err.Error())
		return false
	}

	return true
}

// 发送升级信号
func SendUpgrade(pid int) bool {
	return Signal(pid, syscall.SIGHUP)
}

// 发送关闭信号
func SendKill(pid int) bool {
	return Signal(pid, syscall.SIGINT)
}

// 启动进程
func StartProcess(name string, argv []string) error {

	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	p, err := os.StartProcess(name, argv, attr)

	if err != nil {
		return err
	}

	ret, err := p.Wait()

	if err != nil {
		return err
	}

	LOG_DEBUG("StartProcess: %s => %d", name, ret.Exited())

	return nil
}

// 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤
func ListDirName(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	//	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

// 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

//获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)                                                     //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil { //忽略错误
			//	return err
			return nil
		}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})

	return files, err
}
