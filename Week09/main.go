package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
	用 Go 实现一个 tcp server ，用两个 goroutine 读写 conn，两个 goroutine 通过 chan 可以传递 message，能够正确退出
*/

var mgr *sessionMgr

type sessionMgr struct {
	sessions map[int64]*session
	sync.Mutex
}

func NewSessionMgr() *sessionMgr {
	return &sessionMgr{sessions: make(map[int64]*session)}
}

func (sm *sessionMgr) Add(s *session) {
	sm.Lock()
	defer sm.Unlock()
	sm.sessions[s.id] = s
}

func (sm *sessionMgr) Close() {
	sm.Lock()
	defer sm.Unlock()
	for _, s := range sm.sessions {
		s.Stop()
		delete(sm.sessions, s.id)
	}
}

type session struct {
	id      int64
	conn    net.Conn
	bufChan chan []byte
	done    chan struct{}
	isClose bool
}

func NewSession(id int64, conn net.Conn) *session {
	return &session{
		id:      id,
		conn:    conn,
		bufChan: make(chan []byte, 1024),
		done:    make(chan struct{}, 1),
	}
}

func (s *session) write() {
	for {
		select {
		case msg := <-s.bufChan:
			if _, err := s.conn.Write(msg); err != nil {
				log.Printf("write error: %v\n", err)
				return
			}
		case <-s.done:
			return
		}
	}
}

func (s *session) read() {
	defer s.Stop()
	rd := bufio.NewReader(s.conn)
	for {
		select {
		case <-s.done:
			return
		default:
			line, _, err := rd.ReadLine()
			if err != nil {
				log.Printf("read error: %v\n", err)
				return
			}
			s.bufChan <- line
		}
	}
}

func (s *session) Start() {
	go s.read()
	go s.write()
}

func (s *session) Stop() {
	if s.isClose {
		return
	}
	s.isClose = true
	s.conn.Close()
	close(s.done)
}

func main() {
	listen, err := net.Listen("tcp", ":30001")
	if err != nil {
		panic(err)
	}

	mgr = NewSessionMgr()

	go func() {
		var id int64
		for {
			conn, err := listen.Accept()
			if err != nil {
				fmt.Println("accept err:", err)
				return
			}

			session := NewSession(id, conn)
			id++
			mgr.Add(session)
			session.Start()
		}
	}()

	sc := make(chan os.Signal)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sc:
		log.Printf("收到退出信号[%s]\n", sig.String())
		mgr.Close()
	}

}
