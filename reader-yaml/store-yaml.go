package reader_yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
	"github.com/ghodss/yaml"
)

//Reader yaml文件读取器
type Reader struct {
	data eosc.IUntyped
}

type byteData []byte

func load(path string) ([]byteData, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	data := make([]byteData, 0, 10)
	if f.IsDir() {
		// 如果是目录，则遍历文件
		fs, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		tmpPath := strings.TrimSuffix(path, "/")
		for _, f := range fs {
			filePath := fmt.Sprintf("%s/%s", tmpPath, f.Name())
			log.Info(fmt.Sprintf("read file, path is %s", filePath))
			d, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Error(err)
				continue
			}
			data = append(data, d)
		}
	} else {
		d, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		data = append(data, d)
	}
	return data, nil
}

//NewYaml 根据文件路径创建一个yaml读取器
func NewYaml(file string) (*Reader, error) {
	datas, err := load(file)
	if err != nil {
		return nil, err
	}
	s := &Reader{
		data: eosc.NewUntyped(),
	}
	for _, data := range datas {
		c := new(Config)

		if err = yaml.Unmarshal(data, c); err != nil {
			return nil, err
		}

		now := eosc.Now()
		err = s.setData(c.Router, "router", now)
		if err != nil {
			return nil, err
		}

		err = s.setData(c.Service, "service", now)
		if err != nil {
			return nil, err
		}

		err = s.setData(c.Upstream, "upstream", now)
		if err != nil {
			return nil, err
		}

		err = s.setData(c.Discovery, "discovery", now)
		if err != nil {
			return nil, err
		}
		err = s.setData(c.Auth, "auth", now)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Reader) getByProfession(profession string) eosc.IUntyped {
	if pd, has := s.data.Get(profession); has {
		p, ok := pd.(eosc.IUntyped)
		if ok {
			return p
		}
	}
	pd := eosc.NewUntyped()
	s.data.Set(profession, pd)
	return pd
}

func (s *Reader) setData(items []Item, profession string, now string) error {
	for _, r := range items {
		pd := s.getByProfession(profession)

		v, err := r.newStoreValue(profession, now)
		if err != nil {
			return err
		}
		log.Infof("%s: %s", profession, v.Id)
		pd.Set(v.Id, v)

	}
	return nil
}

//AllByProfession 根据profession返回StoreValue实例列表
func (s *Reader) AllByProfession(profession string) []eosc.StoreValue {
	pd, has := s.data.Get(profession)
	if !has {
		return nil
	}
	p, ok := pd.(eosc.IUntyped)
	if !ok {
		return nil
	}
	list := p.List()
	res := make([]eosc.StoreValue, len(list))
	for i, v := range list {
		res[i] = *(v.(*eosc.StoreValue))
	}
	return res
}
