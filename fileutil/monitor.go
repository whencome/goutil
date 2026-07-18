package fileutil

import (
	"path/filepath"
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
type FileMonitor struct {
	watcher   *fsnotify.Watcher
	files     map[string]*fileInf
	watchDirs []string // 监控目录
	Error     error    // 记录错误信息
}

func NewFileMonitor() *FileMonitor {
	return &FileMonitor{
		files:     map[string]*fileInf{},
		watchDirs: make([]string, 0),
	}
}

// Watch 添加文件监控
// file 要监控的文件
// callback 文件发生变更后执行的回调操作
// checkNow 是否立即执行一次检查
// 注意：如果重复添加一个文件，则会覆盖之前的设置
func (m *FileMonitor) Watch(file string, callback func(f string) error, checkNow bool) error {
	m.files[file] = &fileInf{
		LastCheckTime:  time.Now().Unix(),
		ChangeCallback: callback,
	}
	dirPath := filepath.Dir(file)
	if !array.Contains(m.watchDirs, dirPath) {
		m.watchDirs = append(m.watchDirs, dirPath)
	}
	if checkNow && !goutil.IsNil(callback) {
		e := callback(file)
		if e != nil {
			return e
		}
		m.files[file].LastCheckTime = time.Now().Unix()
	}
	return nil
}

// Run 运行文件监控服务
func (m *FileMonitor) Run() {
	var err error
	m.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		m.Error = err
		log.Errorf("create fsnotify.Watcher fail: %v", err)
		return
	}
	for _, dirPath := range m.watchDirs {
		m.watcher.Add(dirPath)
	}
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					log.Infof("watcher Events channel closed")
					return
				}
				if event.Has(fsnotify.Write) {
					// event.Name 就是变更的文件名（含路径）
					if fi, ok := m.files[event.Name]; ok {
						fi.LastCheckTime = timeutil.Now().Unix()
						if goutil.IsNil(fi.ChangeCallback) {
							continue
						}
						err = fi.ChangeCallback(event.Name)
						if err != nil {
							log.Errorf("file monitor: exec file [%s] change callback failed: %s", event.Name, err)
						} else {
							log.Infof("file monitor: exec file [%s] change callback success", event.Name)
						}
					}
				}
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				log.Errorf("watcher Errors: %v", err)
			}
		}
	}()
}

// Close 关闭并释放资源
func (m *FileMonitor) Close() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}
