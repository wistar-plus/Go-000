package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

/*
	问题：基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
*/

func main() {
	g := new(errgroup.Group)
	sc := make(chan os.Signal)
	stop := make(chan struct{})
	done := make(chan error, 2)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	g.Go(func() error {
		err := server(stop)
		done <- errors.WithMessage(err, "server退出")
		return err
	})

	g.Go(func() error {
		err := pprof(stop)
		done <- errors.WithMessage(err, "pprof退出")
		return err
	})

	go func() {
		select {
		case sig := <-sc:
			log.Printf("收到退出信号[%s]\n", sig.String())
		case err := <-done:
			log.Println(err)
		}
		close(stop)
	}()

	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
}

func server(stop <-chan struct{}) error {
	s := http.Server{Addr: ":12340"}
	go func() {
		<-stop
		fmt.Println("server收到退出指令")
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()
}

func pprof(stop <-chan struct{}) error {
	s := http.Server{Addr: ":12341"}
	go func() {
		<-stop
		fmt.Println("pprof收到退出指令")
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()
}
