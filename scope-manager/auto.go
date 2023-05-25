/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package scope_manager

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var (
	workers eosc.IWorkers
)

func init() {
	bean.Autowired(&workers)
}

func Auto[T any](requireId string, scope string) IProxyOutput[T] {
	if requireId == "" {
		return Get[T](scope)
	}
	w, has := workers.Get(requireId)
	if !has {
		return Get[T](requireId)
	}
	return NewProxy(w.(T))
}
