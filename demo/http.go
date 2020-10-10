package main

import (
	"cosgo/express"
	"cosgo/logger"
	"fmt"
	"sync"
)

func NewHttpMod(name string) *httpMod {
	return &httpMod{name: name}
}

type httpMod struct {
	name    string
	express *express.Engine
}

func (this *httpMod) ID() string {
	return this.name
}

func (this *httpMod) Load() error {
	this.express = express.New("")
	//this.express.Use(middleware1, middleware2)
	return nil
}

func (this *httpMod) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	this.express.GET("/*", hello)
	this.express.POST("/*", hello)
	this.express.GET("/:api", hello2)
	//this.express.Static("/*", ".")
	return this.express.Start()
}

func (this *httpMod) Close(wgp *sync.WaitGroup) error {
	this.express.Close()
	wgp.Done()
	return nil
}

func middleware1(c *express.Context) {
	logger.Debug("middleware1")
	c.Next()
}
func middleware2(c *express.Context) {
	logger.Debug("middleware2")
	c.Next()
}

// Handler
func hello(c *express.Context) error {
	//logger.Debug("hello1")
	//return c.End()
	c.String(fmt.Sprintf("Hello, World 1!  %v\n", c.Param("api")))
	c.Next()
	return nil
	//return c.String(fmt.Sprintf("Hello, World 1!  %v\n", c.Param("api")))
}

func hello2(c *express.Context) error {
	//logger.Debug("hello2")
	return c.String(fmt.Sprintf("Hello, World 2!  %v\n", c.Param("api")))
}
