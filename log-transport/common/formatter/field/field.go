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

//FieldKey 域的key
type FieldKey = string

//FieldKeys 域列表
type FieldKeys []FieldKey

//FieldsFormatter access日志格式器
type FieldsFormatter struct {
	fields          FieldKeys
	locker          sync.RWMutex
	timestampFormat string
	ignoreTime      bool
}

//NewFieldsLogFormatter 创建携带域的日志输出格式处理器
func NewFieldsLogFormatter(fields FieldKeys, timestampFormat string) *FieldsFormatter {

	return &FieldsFormatter{
		fields:          fields,
		locker:          sync.RWMutex{},
		timestampFormat: timestampFormat,
		ignoreTime:      false,
	}
}

//IgnoreTime 返回是否忽略时间的布尔值
func (f *FieldsFormatter) IgnoreTime() bool {
	return f.ignoreTime
}

//SetIgnoreTime 设置是否忽略时间
func (f *FieldsFormatter) SetIgnoreTime(ignoreTime bool) {
	f.ignoreTime = ignoreTime
}

//SetFields 设置域
func (f *FieldsFormatter) SetFields(fields FieldKeys) {
	f.locker.Lock()
	f.fields = fields
	f.locker.Unlock()
}

//Fields 返回域列表
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
