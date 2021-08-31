package field

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/eosc/log"
)

const (
	//DefaultTimeStampFormatter 时间戳默认格式化字符串
	DefaultTimeStampFormatter = "[2006-01-02 15:04:05]"
	//TimeIso8601Formatter iso8601格式化
	TimeIso8601Formatter = "[" + time.RFC3339 + "]"

	//RequestTime = "request_time"
	//TimeIso8601 = "time_iso8601"
	//Msec = "msec"
	//TimeLocal = "time_local"

)

type FieldKey = string
type FieldKeys []FieldKey

//FieldsFormatter access日志格式器
type FieldsFormatter struct {
	fields          FieldKeys
	locker          sync.RWMutex
	timestampFormat string
	ignoreTime      bool
}

func NewFieldsLogFormatter(fields FieldKeys, timestampFormat string) *FieldsFormatter {

	return &FieldsFormatter{
		fields:          fields,
		locker:          sync.RWMutex{},
		timestampFormat: timestampFormat,
		ignoreTime:      false,
	}
}

func (f *FieldsFormatter) IgnoreTime() bool {
	return f.ignoreTime
}

func (f *FieldsFormatter) SetIgnoreTime(ignoreTime bool) {
	f.ignoreTime = ignoreTime
}

//SetFields 设置域
func (f *FieldsFormatter) SetFields(fields FieldKeys) {
	f.locker.Lock()
	f.fields = fields
	f.locker.Unlock()
}

//SetFields 设置域
func (f *FieldsFormatter) Fields() []FieldKey {
	f.locker.Lock()
	fields := f.fields
	f.locker.Unlock()
	return fields
}

//Format 格式化
func (f *FieldsFormatter) Format(entry *log.Entry) ([]byte, error) {

	b := &bytes.Buffer{}

	timestampFormat := f.timestampFormat
	if timestampFormat == "" {
		timestampFormat = DefaultTimeStampFormatter
	}

	data := entry.Data
	//if !f.ignoreTime{
	//	data[TimeLocal] = entry.Time.Format(timestampFormat)
	//	data[TimeIso8601] = entry.Time.Format(TimeIso8601Formatter)
	//
	//	msec := entry.Time.UnixNano() / int64(time.Millisecond)
	//	data[Msec] = fmt.Sprintf("%d.%d", msec/1000, msec%1000)
	//
	//	requestTIme := data[RequestTime].(time.Duration)
	//	data[RequestTime] = fmt.Sprintf("%dms", requestTIme/time.Millisecond)
	//}

	for _, key := range f.Fields() {
		b.WriteByte('\t')
		if v, has := data[key]; has {
			f.appendValue(b, v)
		} else {
			f.appendValue(b, "-")
		}
	}
	b.WriteByte('\n')
	p := b.Bytes()
	return p[1:], nil
}

func (f *FieldsFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(stringVal)

}
