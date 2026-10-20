package fileutil

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/whencome/goutil"
	"github.com/whencome/goutil/array"
	"github.com/whencome/goutil/log"
	"github.com/whencome/goutil/timeutil"
)

// 定义一个文件信息结构提
type fileInf struct {
	LastCheckTime  int64
	ChangeCallback func(string) error
}

// FileMonitor 文件变更监控器，主动监控文件是否有变更
// 注意：这里的逻辑是主动定时检查，而不是文件发生变化时自动触发，应用场景时类似配置变化主动reload服务等
// 此对象为并发安全对象，Watch与Run可以在不同的goroutine中调用
type FileMonitor struct {
	watcher   *fsnotify.Watcher
	files     map[string]*fileInf
	watchDirs []string // 监控目录
	Error     error    // 记录错误信息

	mu         sync.RWMutex // 保护files/watchDirs/watcher/running
	running    bool         // 是否已经启动
	closed     bool         // 是否已经关闭
	stopCh     chan struct{}
	stopOnce   sync.Once
	closeOnce  sync.Once
	debounce   time.Duration   // 去抖时间，避免同一文件在短时间内被多次触发
	lastEvents map[string]int64 // 记录每个文件上次触发回调的时间（纳秒），用于去抖
}

// NewFileMonitor 创建一个文件监控器
func NewFileMonitor() *FileMonitor {
	return &FileMonitor{
		files:      map[string]*fileInf{},
		watchDirs:  make([]string, 0),
		stopCh:     make(chan struct{}),
		lastEvents: make(map[string]int64),
	}
}

// SetDebounce 设置去抖时间
// 说明：同一个文件在去抖时间内的多次变更只会触发一次回调，用于避免"保存文件"等操作
// 在极短时间内产生多次写事件导致回调被重复执行；默认值为0，表示不去抖（保持原有行为）
func (m *FileMonitor) SetDebounce(d time.Duration) {
	m.mu.Lock()
	m.debounce = d
	m.mu.Unlock()
}

// Watch 添加文件监控
// file 要监控的文件
// callback 文件发生变更后执行的回调操作
// checkNow 是否立即执行一次检查
// 注意：如果重复添加一个文件，则会覆盖之前的设置
func (m *FileMonitor) Watch(file string, callback func(f string) error, checkNow bool) error {
	if file == "" {
		return errors.New("file to watch is empty")
	}
	// 统一路径格式，避免因路径写法不一致（如含./前缀）导致事件匹配失败
	file = filepath.Clean(file)
	fi := &fileInf{
		LastCheckTime:  time.Now().Unix(),
		ChangeCallback: callback,
	}
	dirPath := filepath.Dir(file)
	m.mu.Lock()
	if m.files == nil {
		m.files = map[string]*fileInf{}
	}
	m.files[file] = fi
	isNewDir := !array.Contains(m.watchDirs, dirPath)
	if isNewDir {
		m.watchDirs = append(m.watchDirs, dirPath)
	}
	watcher := m.watcher
	running := m.running
	m.mu.Unlock()

	// 若监控器已经启动，需要把新增的目录加入监控，否则新增文件的变更不会被感知
	if isNewDir && running && watcher != nil {
		if err := watcher.Add(dirPath); err != nil {
			m.mu.Lock()
			m.Error = err
			m.mu.Unlock()
			return err
		}
	}

	if checkNow && !goutil.IsNil(callback) {
		e := callback(file)
		if e != nil {
			return e
		}
		m.mu.Lock()
		if cur, ok := m.files[file]; ok {
			cur.LastCheckTime = time.Now().Unix()
		}
		m.mu.Unlock()
	}
	return nil
}

// Run 运行文件监控服务
// 说明：重复调用此方法不会重复启动监控协程
func (m *FileMonitor) Run() {
	m.mu.Lock()
	if m.running || m.closed {
		m.mu.Unlock()
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.Error = err
		m.mu.Unlock()
		log.Errorf("create fsnotify.Watcher fail: %v", err)
		return
	}
	if m.files == nil {
		m.files = map[string]*fileInf{}
	}
	m.watcher = watcher
	dirs := make([]string, len(m.watchDirs))
	copy(dirs, m.watchDirs)
	if m.stopCh == nil {
		m.stopCh = make(chan struct{})
	}
	stopCh := m.stopCh
	m.running = true
	m.mu.Unlock()

	for _, dirPath := range dirs {
		if err := watcher.Add(dirPath); err != nil {
			// 单个目录添加失败不影响其它目录的监控
			m.mu.Lock()
			m.Error = err
			m.mu.Unlock()
			log.Errorf("file monitor: watch dir [%s] failed: %s", dirPath, err)
		}
	}

	go m.loop(watcher, stopCh)
}

// loop 监控事件循环
func (m *FileMonitor) loop(watcher *fsnotify.Watcher, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				log.Infof("watcher Events channel closed")
				return
			}
			if !event.Has(fsnotify.Write) {
				continue
			}
			m.handleEvent(event.Name)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("watcher Errors: %v", err)
		}
	}
}

// handleEvent 处理文件的变更事件
func (m *FileMonitor) handleEvent(name string) {
	eventName := filepath.Clean(name)
	m.mu.Lock()
	fi, ok := m.files[eventName]
	if !ok {
		m.mu.Unlock()
		return
	}
	// 去抖处理：同一文件在去抖时间内的重复事件直接忽略
	if m.debounce > 0 {
		now := time.Now().UnixNano()
		if last, exists := m.lastEvents[eventName]; exists && now-last < int64(m.debounce) {
			m.mu.Unlock()
			return
		}
		m.lastEvents[eventName] = now
	}
	fi.LastCheckTime = timeutil.Now().Unix()
	callback := fi.ChangeCallback
	m.mu.Unlock()

	if goutil.IsNil(callback) {
		return
	}
	if err := callback(eventName); err != nil {
		log.Errorf("file monitor: exec file [%s] change callback failed: %s", eventName, err)
	} else {
		log.Infof("file monitor: exec file [%s] change callback success", eventName)
	}
}

// Close 关闭并释放资源
// 说明：此方法可以重复调用
func (m *FileMonitor) Close() error {
	m.mu.Lock()
	watcher := m.watcher
	stopCh := m.stopCh
	m.running = false
	m.closed = true
	m.mu.Unlock()

	if stopCh != nil {
		m.stopOnce.Do(func() {
			close(stopCh)
		})
	}
	if watcher == nil {
		return nil
	}
	var err error
	m.closeOnce.Do(func() {
		err = watcher.Close()
	})
	return err
}
