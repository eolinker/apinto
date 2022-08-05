package file_transport

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/eolinker/eosc/log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//MaxBuffer buffer最大值
const MaxBuffer = 1024 * 500

var (
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

//FileWriterByPeriod 文件周期写入
type FileWriterByPeriod struct {
	wC chan *bytes.Buffer

	enable     bool
	cancelFunc context.CancelFunc
	locker     sync.RWMutex
	wg         sync.WaitGroup
	resetChan  chan FileController
}

//NewFileWriteByPeriod 获取新的FileWriterByPeriod
func NewFileWriteByPeriod(cfg *Config) *FileWriterByPeriod {
	w := &FileWriterByPeriod{
		locker:    sync.RWMutex{},
		wg:        sync.WaitGroup{},
		enable:    false,
		resetChan: make(chan FileController),
	}

	w.Open(&FileController{
		dir:    cfg.Dir,
		file:   cfg.File,
		period: cfg.Period,
		expire: time.Duration(cfg.Expire) * 24 * time.Hour,
	})
	return w
}

func (w *FileWriterByPeriod) Reset(cfg *Config) {
	w.resetChan <- FileController{
		dir:    cfg.Dir,
		file:   cfg.File,
		period: cfg.Period,
		expire: time.Duration(cfg.Expire) * 24 * time.Hour,
	}
}

//Open 打开
func (w *FileWriterByPeriod) Open(config *FileController) {
	w.locker.Lock()
	defer w.locker.Unlock()

	if w.enable {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.cancelFunc = cancel
	w.wC = make(chan *bytes.Buffer, 100)

	w.enable = true
	go func() {
		w.wg.Add(1)
		w.do(ctx, config)
		w.wg.Done()
	}()
}

//Close 关闭
func (w *FileWriterByPeriod) Close() {

	isClose := false
	w.locker.Lock()
	if !w.enable {
		w.locker.Unlock()
		return
	}

	if w.cancelFunc != nil {
		isClose = true
		w.cancelFunc()
		w.cancelFunc = nil
	}
	w.enable = false
	w.locker.Unlock()
	if isClose {
		w.wg.Wait()
	}
}
func (w *FileWriterByPeriod) isEnable() bool {
	w.locker.Lock()
	defer w.locker.Unlock()
	return w.enable
}
func (w *FileWriterByPeriod) Write(p []byte) (n int, err error) {

	l := len(p)

	if l == 0 {
		return
	}
	if !w.isEnable() {
		return l, nil
	}
	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	buffer.Write(p)
	if p[l-1] != '\n' {
		buffer.WriteByte('\n')
	}
	w.wC <- buffer
	return l, nil
}

func (w *FileWriterByPeriod) do(ctx context.Context, config *FileController) {
	fileController := *config
	fileController.initFile()
	f, lastTag, e := fileController.openFile()
	if e != nil {
		log.Errorf("open log file:%s\n", e.Error())
		return
	}

	buf := bufio.NewWriter(f)
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	tFlush := time.NewTimer(time.Second)

	for {
		select {
		case <-ctx.Done():
			{
				for len(w.wC) > 0 {
					p := <-w.wC
					buf.Write(p.Bytes())
					bufferPool.Put(p)
				}
				buf.Flush()
				f.Close()
				t.Stop()
				//w.wg.Done()
				return
			}

		case <-t.C:
			{
				if buf.Buffered() > 0 {
					buf.Flush()
					tFlush.Reset(time.Second)
				}
				if lastTag != fileController.timeTag(time.Now()) {

					f.Close()
					fileController.history(lastTag)
					fnew, tag, err := fileController.openFile()
					if err != nil {
						return
					}
					lastTag = tag
					f = fnew
					buf.Reset(f)

					go fileController.dropHistory()
				}

			}
		case <-tFlush.C:
			{
				if buf.Buffered() > 0 {
					buf.Flush()
				}
				tFlush.Reset(time.Second)
			}
		case p := <-w.wC:
			{
				buf.Write(p.Bytes())
				bufferPool.Put(p)
				if buf.Buffered() > MaxBuffer {
					buf.Flush()
				}
				tFlush.Reset(time.Second)
			}
		case controller, ok := <-w.resetChan:
			{
				if ok {
					fileController = controller
				}
			}

		}
	}
}

func (w *FileController) initFile() {
	err := os.MkdirAll(w.dir, 0755)
	if err != nil {
		log.Error(err)
	}
	path := filepath.Join(w.dir, fmt.Sprintf("%s.log", w.file))
	nowTag := w.timeTag(time.Now())
	if info, e := os.Stat(path); e == nil {

		timeTag := w.timeTag(info.ModTime())
		if timeTag != nowTag {
			w.history(timeTag)
		}
	}

	w.dropHistory()

}

func (w *FileController) openFile() (*os.File, string, error) {
	path := filepath.Join(w.dir, fmt.Sprintf("%s.log", w.file))
	nowTag := w.timeTag(time.Now())
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, "", err
	}
	return f, nowTag, err

}
