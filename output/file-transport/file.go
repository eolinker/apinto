package file_transport

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileController struct {
	expire time.Duration
	dir    string
	file   string
	period LogPeriod
}

func (w *FileController) timeTag(t time.Time) string {

	tag := t.Format(w.period.FormatLayout())

	return filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.file, tag))
}
func (w *FileController) fileName() string {
	return filepath.Join(w.dir, fmt.Sprintf("%s.log", w.file))
}
func (w *FileController) history(history string) {

	path := w.fileName()
	os.Rename(path, history)

}

func (w *FileController) dropHistory() {

	expireTime := time.Now().Add(-w.expire)
	pathPatten := filepath.Join(w.dir, fmt.Sprintf("%s-*", w.file))
	files, err := filepath.Glob(pathPatten)
	if err == nil {
		for _, f := range files {
			if info, e := os.Stat(f); e == nil {

				if expireTime.After(info.ModTime()) {
					_ = os.Remove(f)
				}
			}

		}
	}
}
