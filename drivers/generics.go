package drivers

import (
	"fmt"
	"reflect"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/config"
	"github.com/eolinker/eosc/utils/schema"
)

type Driver[T any] struct {
	profession string
	driver     string
	label      string
	desc       string
	configType reflect.Type
	createFunc func(id, name string, v *T, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error)
}

func (d *Driver[T]) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	cfg, err := d.Assert(v)
	if err != nil {
		return nil, err
	}
	return d.createFunc(id, name, cfg, workers)
}

type DriverConfigChecker[T any] struct {
	Driver[T]
	configCheckFunc func(v *T, workers map[eosc.RequireId]eosc.IWorker) error
}

func (d *DriverConfigChecker[T]) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := d.Assert(v)
	if err != nil {
		return err
	}
	return d.configCheckFunc(cfg, workers)
}

func (d *Driver[T]) Assert(v interface{}) (*T, error) {
	return Assert[T](v)
}
func Assert[T any](v interface{}) (*T, error) {
	cfg, ok := v.(*T)
	if !ok {
		return nil, fmt.Errorf("%w:need %s,now %s", eosc.ErrorConfigType, config.TypeNameOf((*T)(nil)), config.TypeNameOf(v))
	}
	return cfg, nil
}

func (d *Driver[T]) ConfigType() reflect.Type {
	return d.configType
}

type Factory[T any] struct {
	configType      reflect.Type
	render          interface{}
	createFunc      func(id, name string, v *T, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error)
	configCheckFunc func(v *T, workers map[eosc.RequireId]eosc.IWorker) error
}

func NewFactory[T any](createFunc func(id, name string, v *T, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error), configCheckFunc ...func(v *T, workers map[eosc.RequireId]eosc.IWorker) error) *Factory[T] {
	configType := reflect.TypeOf((*T)(nil))
	render, _ := schema.Generate(configType, nil)

	f := &Factory[T]{
		createFunc: createFunc,
		configType: configType,
		render:     render,
	}
	if len(configCheckFunc) == 1 {
		f.configCheckFunc = configCheckFunc[0]
	}
	return f
}

func (f *Factory[T]) Render() interface{} {
	return f.render
}

func (f *Factory[T]) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	if f.configCheckFunc == nil {
		return &Driver[T]{
			profession: profession,
			driver:     name,
			label:      label,
			desc:       desc,
			configType: f.configType,
			createFunc: f.createFunc,
		}, nil
	} else {
		return &DriverConfigChecker[T]{
			Driver: Driver[T]{
				profession: profession,
				driver:     name,
				label:      label,
				desc:       desc,
				configType: f.configType,
				createFunc: f.createFunc,
			},
			configCheckFunc: f.configCheckFunc,
		}, nil
	}
}

type WorkerBase struct {
	id   string
	name string
}

func Worker(id string, name string) WorkerBase {
	return WorkerBase{id: id, name: name}
}

func (w *WorkerBase) Id() string {
	return w.id
}
func (w *WorkerBase) Name() string {
	return w.name
}
