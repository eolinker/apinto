package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/eolinker/apinto/example/dubbo2/model"
)

type Server struct {
}

func (s *Server) ComplexServer(ctx context.Context, servers *model.ComplexServer) (*model.ComplexServer, error) {
	return servers, nil
}

func (s *Server) UpdateList(ctx context.Context, servers []*model.Server) ([]*model.Server, error) {

	for _, server := range servers {
		if server.Id == 10 {
			return nil, errors.New("id为10的server不存在")
		}
		server.Name = "hello"
	}

	return servers, nil
}

func (s *Server) GetById(ctx context.Context, id int64) (*model.Server, error) {
	fmt.Println(ctx, id)

	if id == 10 {
		return nil, errors.New("id为10的server不存在")
	}
	return &model.Server{
		Id:    id,
		Name:  "apinto",
		Age:   20,
		Email: "apinto@qq.com",
	}, nil
}

func (s *Server) Update(ctx context.Context, server *model.Server) error {

	fmt.Println(*server)
	if server.Id == 10 {
		return errors.New("不能改为ID为10的server")
	}

	return nil
}

func (s *Server) List(ctx context.Context, server *model.Server) ([]*model.Server, error) {
	fmt.Println(*server)
	list := make([]*model.Server, 0)
	list = append(list, &model.Server{
		Id:    10,
		Name:  "apinto1",
		Age:   10,
		Email: "apinto1@qq.com",
	})

	list = append(list, &model.Server{
		Id:    20,
		Name:  "apinto2",
		Age:   20,
		Email: "apinto2@qq.com",
	})

	list = append(list, server)

	return list, nil
}
