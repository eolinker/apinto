package yaml

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
)

func NewStore() *Store {
	return &Store{
		employees: make(map[string][]string),
		drivers:   map[string][]eosc.IStore{},
		canWrite:  false,
	}
}

type Store struct {
	employees   map[string][]string
	professions []string
	drivers     map[string][]eosc.IStore
	canWrite    bool
}

func (s *Store) SetEmployee(employees Employees) {
	for key, value := range employees {
		if _, ok := s.employees[key]; !ok {
			s.employees[key] = make([]string, 0, len(value))
		}
		for _, v := range value {
			val := reflect.ValueOf(v)
			result := make(map[string]interface{})
			if val.Kind() == reflect.Map {
				m := val.MapRange()
				for m.Next() {
					result[m.Key().Interface().(string)] = m.Value().Interface()
				}
			}
			data, err := json.Marshal(result)
			if err != nil {
				log.Error(err)
				continue
			}
			s.employees[key] = append(s.employees[key], string(data))
		}
		s.professions = append(s.professions, key)
	}
}

func (s *Store) GetEmployee(profession string) ([]string, error) {
	if v, ok := s.employees[profession]; ok {
		return v, nil
	}
	return nil, errors.New("the employee does not exist")
}

func (s *Store) Professions() []string {
	return s.professions
}

func (s *Store) Info(profession string, nameId string) (string, error) {
	panic("implement me")
}

func (s *Store) All(profession string) (string, error) {
	panic("implement me")
}

func (s *Store) Mode() string {
	return mode
}

func (s *Store) InfoByID(id string) string {
	if v, ok := s.employees[id]; ok {
		data, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(data)
	}
	return ""
}

func (s *Store) CanWrite() bool {
	return s.canWrite
}
