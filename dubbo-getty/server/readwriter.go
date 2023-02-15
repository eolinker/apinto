/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package getty

import (
	"github.com/eolinker/eosc/log"
	"reflect"
)

import (
	"github.com/eolinker/apinto/dubbo-getty"

	perrors "github.com/pkg/errors"
)

import (
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"dubbo.apache.org/dubbo-go/v3/remoting"
)

// RpcServerPackageHandler Read data from client and Write data to client
type RpcServerPackageHandler struct {
	server *Server
}

func NewRpcServerPackageHandler(server *Server) *RpcServerPackageHandler {
	return &RpcServerPackageHandler{server: server}
}

// Read data from client. if the package size from client is larger than 4096 byte, client will read 4096 byte
// and send to client each time. the Read can assemble it.
func (p *RpcServerPackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	req, length, err := (p.server.codec).Decode(data)
	if err != nil {
		err = perrors.WithStack(err)
	}
	if req == ((*remoting.DecodeResult)(nil)) {
		return nil, length, err
	}
	if req.Result == ((*remoting.Request)(nil)) || req.Result == ((*remoting.Response)(nil)) {
		return nil, length, err // as getty rule
	}
	return req, length, err
}

// Write send the data to client
func (p *RpcServerPackageHandler) Write(ss getty.Session, pkg interface{}) ([]byte, error) {
	res, ok := pkg.(*remoting.Response)
	maxBufLength := p.server.conf.GettySessionParam.MaxMsgLen + impl.HEADER_LENGTH
	if ok {
		buf, err := (p.server.codec).EncodeResponse(res)
		bufLength := buf.Len()
		if bufLength > maxBufLength {
			log.Errorf("Data length %d too large, max payload %d", bufLength-impl.HEADER_LENGTH, p.server.conf.GettySessionParam.MaxMsgLen)
			return nil, perrors.Errorf("Data length %d too large, max payload %d", bufLength-impl.HEADER_LENGTH, p.server.conf.GettySessionParam.MaxMsgLen)
		}
		if err != nil {
			log.Warnf("binary.Write(res{%#v}) = err{%#v}", res, perrors.WithStack(err))
			return nil, perrors.WithStack(err)
		}
		return buf.Bytes(), nil
	}

	req, ok := pkg.(*remoting.Request)
	if ok {
		buf, err := (p.server.codec).EncodeRequest(req)
		bufLength := buf.Len()
		if bufLength > maxBufLength {
			log.Errorf("Data length %d too large, max payload %d", bufLength-impl.HEADER_LENGTH, p.server.conf.GettySessionParam.MaxMsgLen)
			return nil, perrors.Errorf("Data length %d too large, max payload %d", bufLength-impl.HEADER_LENGTH, p.server.conf.GettySessionParam.MaxMsgLen)
		}
		if err != nil {
			log.Warnf("binary.Write(req{%#v}) = err{%#v}", res, perrors.WithStack(err))
			return nil, perrors.WithStack(err)
		}
		return buf.Bytes(), nil
	}

	log.Errorf("illegal pkg:%+v\n, it is %+v", pkg, reflect.TypeOf(pkg))
	return nil, perrors.New("invalid rpc response")
}
