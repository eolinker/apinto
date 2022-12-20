package monitor_entry

import "time"

var _ IPoint = (*Point)(nil)

type IPoint interface {
	Table() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
}

func NewPoint(table string, tags map[string]string, fields map[string]interface{}, pointTime time.Time) *Point {
	return &Point{table: table, tags: tags, fields: fields, pointTime: pointTime}
}

type Point struct {
	table     string
	tags      map[string]string
	fields    map[string]interface{}
	pointTime time.Time
}

func (p *Point) Table() string {
	return p.table
}

func (p *Point) Tags() map[string]string {
	return p.tags
}

func (p *Point) Fields() map[string]interface{} {
	return p.fields
}

func (p *Point) Time() time.Time {
	return p.pointTime
}
