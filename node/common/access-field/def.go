package access_field

import "reflect"

type FieldInfo struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Select bool   `json:"select"`
}

func GenSelectFieldList(selectFields []string) []*FieldInfo {
	isAll := len(selectFields) == 0
	set := make(map[string]bool)
	for _, v := range selectFields {
		set[v] = true
	}

	v := reflect.ValueOf(new(Fields)).Elem()
	t := v.Type()
	n := t.NumField()
	r := make([]*FieldInfo, 0, n)
	tmp := make([]*FieldInfo, 0, n-len(selectFields))

	for i := 0; i < n; i++ {
		structField := t.Field(i)
		fn := structField.Tag.Get("field")
		desc := structField.Tag.Get("desc")
		if fn == "" {
			continue
		}
		info := &FieldInfo{
			Name:   fn,
			Desc:   desc,
			Select: isAll || set[fn],
		}
		if info.Select {
			r = append(r, info)
		} else {
			tmp = append(tmp, info)
		}
	}

	r = append(r, tmp...)
	return r
}
