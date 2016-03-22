package xnet

import (
	"db"
	. "logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ITicker interface {

	// tick delay.
	TickDelay() time.Duration

	// on tick.
	OnTick()
}

type IProcessor interface {
	Process()
	Recycle()
}

/////////////////////////////////////////////////////////////////
var task_speed = 0
var last_task_speed = 0

func USecond() int64 {
	return time.Now().UnixNano() / 1000
}

var msg_pool = &sync.Pool{New: func() interface{} {
	return &MessageProcessor{}
}}

type MessageProcessor struct {
	session ISession
	packet  *Packet
}

func (this *MessageProcessor) Process() {
	this.session.Process(this.packet)
}

func (this *MessageProcessor) Recycle() {
	this.session = nil
	this.packet = nil
	msg_pool.Put(this)
}

//////////////////////////////////////////////////////////////
var task_pool = &sync.Pool{New: func() interface{} {
	return &TaskProcessor{}
}}

type TaskProcessor struct {
	fn func()
}

func (this *TaskProcessor) Process() {

	defer func() {
		if err := recover(); err != nil {
			PrintStack("TaskProcessor.Process: %v", err)
		}
	}()

	this.fn()
}

func (this *TaskProcessor) Recycle() {
	this.fn = nil
	task_pool.Put(this)
}

type closeChan chan int
type sigChan chan os.Signal
type procChan chan IProcessor

const PROC_QUEUE_CAPCITY = 100000

type LogicRunner struct {
	stop    closeChan
	wc      closeChan
	sig     sigChan
	queue   procChan
	tick    <-chan time.Time
	service ILogicService
	restart bool
}

func NewLogicRunner(capcity int) *LogicRunner {

	return &LogicRunner{
		stop:    make(closeChan),
		wc:      make(closeChan),
		sig:     make(sigChan, 1),
		queue:   make(procChan, capcity),
		restart: false,
	}
}

// 停止运行
func (this *LogicRunner) Stop() {
	signal.Stop(this.sig)
	close(this.stop)
}

func (this *LogicRunner) stopLogic() {
	close(this.wc)
}

func (this *LogicRunner) runLogic() {

	defer func() {
		if err := recover(); err != nil {
			PrintStack("LogicRunner.runLogic: %v", err)
			close(this.stop)
		}
	}()

	// 消息性能统计
	dumpChan := time.Tick(time.Hour * 1)

	for {
		select {
		// 处理一个消息
		case process := <-this.queue:
			process.Process()
			process.Recycle()
			task_speed++
		// 逻辑帧
		case <-this.tick:
			this.onTick()
			last_task_speed = task_speed
			task_speed = 0
			//		LOG_INFO( "LogicRunner.runLogic: queue: %d, speed: %d", len(this.queue), last_task_speed )
		// 输出消息累计耗时
		case <-dumpChan:
			DefaultRequestRecorder.Dump()
		// 系统信号处理
		case sig := <-this.sig:
			this.OnSignal(sig, this.service)
		// 等待停止指令
		case <-this.wc:
			LOG_DEBUG("逻辑线程已停止!")
			return
		}
	}
}

func (this *LogicRunner) onTick() {

	if nil == this.service {
		return
	}

	PushTask(func() {
		this.service.OnTick()
	})
}

func (this *LogicRunner) safeStopLogic() {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	LOG_DEBUG("正在等待逻辑队列执行完毕(%d)...", TaskLen())

	// 逻辑队列执行完毕推出之
	PushTask(func() {
		// 资源释放完毕之后继续向下执行
		defer wg.Done()
		LOG_DEBUG("执行完毕, 正在清理资源...")
		// 游戏加载的资源清理释放
		this.service.OnDestroy()
	})

	// 等待逻辑队列执行完毕
	wg.Wait()
	LOG_DEBUG("清理完毕, 正在解除节点注册...")

	// 解除节点注册
	this.service.OnRelease()
	LOG_DEBUG("解除完毕, 正在停止逻辑线程...")

	// 停止逻辑线程
	this.stopLogic()
}

// 主逻辑调用
func (this *LogicRunner) Run(service ILogicService) {

	Assert(nil == this.service, "use error: service running at LogicRunner")
	this.service = service

	// 最后等待日志落地完毕
	defer WaitLogger(time.Second * 30)

	// 安全退出
	defer func() {

		if err := recover(); err != nil {
			PrintStack("LogicRunner.Run: %v", err)
		}

		// 1. 安全的停止逻辑相关的任务和资源
		this.safeStopLogic()

		// 2. 最后关闭数据连接
		db.Destroy()

		// 3. 如果需要重启服务器的尝试拉起新的服务器
		if this.restart {
			// 尝试启动新的进程
			if service.Restart() {
				LOG_DEBUG("重启服务器成功!")
			} else {
				LOG_ERROR("重启服务器失败!")
			}
		}

		// 4. 关闭服务
		service.Stop(ServiceOnDestory)
	}()

	// 初始化服务
	PushTask(func() {
		DEBUG("初始化.....")
		if !service.OnInitialize() {
			FATAL("服务初始化失败!")
		}

		// 帧延迟
		this.tick = time.Tick(service.TickDelay())
	})

	// receive system signal.
	signal.Notify(this.sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP)

	// 等待退出
	<-this.stop
}

// 系统信号处理
func (this *LogicRunner) OnSignal(sig os.Signal, service ILogicService) {

	LOG_DEBUG("OnSignal: %v", sig)

	if nil == this.service {
		return
	}

	var result SignalResult

	switch sig {
	// 停止信号
	case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP:
		result = service.OnSigQuit()
	// 升级信号
	case syscall.SIGHUP:
		if result = service.OnSigHup(); SignalResult_Quit == result {
			// 需要重启
			this.restart = true
		}
	// 自定义信号处理
	default:
		result = service.OnSignal(sig)
	}

	// 信号处理结果需要退出服务器
	if SignalResult_Quit == result {
		this.Stop()
	}
}

// 添加一个待运行任务
func (this *LogicRunner) PushTask(handler func()) {
	//this.queue <- &TaskProcessor{ fn : handler }
	processor := task_pool.Get().(*TaskProcessor)
	processor.fn = handler
	this.queue <- processor
}

// 添加一个待制定的消息
func (this *LogicRunner) PutProcess(session ISession, packet *Packet) {
	//this.queue <- &MessageProcessor { session : session, packet : packet }
	processor := msg_pool.Get().(*MessageProcessor)
	processor.session = session
	processor.packet = packet
	this.queue <- processor
}

//----------------------------------------------------------------------------------------------
func init() {
	// 启动逻辑线程
	go RunnerInstance().runLogic()
}

var default_runner *LogicRunner

func RunnerInstance() *LogicRunner {

	if nil == default_runner {
		default_runner = NewLogicRunner(PROC_QUEUE_CAPCITY)
	}

	return default_runner
}

func PushTask(fn func()) {
	RunnerInstance().PushTask(fn)
}

func PutProcess(session ISession, packet *Packet) {
	RunnerInstance().PutProcess(session, packet)
}

func Run(service ILogicService) {
	RunnerInstance().Run(service)
}

func Stop() {
	RunnerInstance().Stop()
}

func TaskSpeed() int32 {
	return int32(last_task_speed)
}

func TaskCap() int32 {
	return int32(cap(RunnerInstance().queue))
}

func TaskLen() int32 {
	return int32(len(RunnerInstance().queue))
}
