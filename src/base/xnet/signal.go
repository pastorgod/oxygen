package xnet

import (
	"os"
)

type SignalResult int

const (
	SignalResult_Continue SignalResult = 0
	SignalResult_Quit     SignalResult = 1
)

type ISignalHandler interface {

	// 退出信号处理
	OnSigQuit() SignalResult

	// 升级信号处理
	OnSigHup() SignalResult

	// 其他信号处理
	OnSignal(os.Signal) SignalResult
}
