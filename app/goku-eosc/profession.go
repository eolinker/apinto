/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package main

import (
	"io/ioutil"

	"github.com/eolinker/eosc"
	"github.com/ghodss/yaml"
)

type config struct {
	Name         string         `json:"name" yaml:"name"`
	Label        string         `json:"label" yaml:"label"`
	Desc         string         `json:"desc" yaml:"desc"`
	Dependencies []string       `json:"dependencies" yaml:"dependencies"`
	AppendLabel  []string       `json:"append_label" yaml:"append_label"`
	Drivers      []driverConfig `json:"drivers" yaml:"drivers"`
}

type driverConfig struct {
	ID     string            `json:"id" yaml:"id"`
	Name   string            `json:"name" yaml:"name"`
	Label  string            `json:"label" yaml:"label"`
	Desc   string            `json:"desc" yaml:"desc"`
	Params map[string]string `json:"params" yaml:"params"`
}

func readProfessionConfig(file string) ([]eosc.ProfessionConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cs := make([]config, 0)
	err = yaml.Unmarshal(data, &cs)
	if err != nil {
		return nil, err
	}
	pcs := make([]eosc.ProfessionConfig, 0, len(cs))
	for _, c := range cs {
		drivers := make([]eosc.DriverConfig, 0, len(c.Drivers))
		for _, driver := range c.Drivers {
			drivers = append(drivers, eosc.DriverConfig{
				ID:     driver.ID,
				Name:   driver.Name,
				Label:  driver.Label,
				Desc:   driver.Desc,
				Params: driver.Params,
			})
		}
		pcs = append(pcs, eosc.ProfessionConfig{
			Name:         c.Name,
			Label:        c.Label,
			Desc:         c.Desc,
			Dependencies: c.Dependencies,
			AppendLabel:  c.AppendLabel,
			Drivers:      drivers,
		})
	}
	return pcs, nil
}
