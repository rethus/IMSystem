package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播
	Message chan string
}

// 创建server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听 Message 广播消息 channel 的 goroutine
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 将 msg 发送给所有在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息方法
func (this *Server) BroadCast(user *User, msg string) {
	// "[user.Addr] user.Name: msg"
	sendMsg := "[" + user.Addr + "] " + user.Name + ": " + msg
	this.Message <- sendMsg
}

// 当前链接的业务
func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)

	user.Online()

	// 监听用户是否活跃的 channel
	isAlive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// n = 0 表示对端关闭
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])

			// 将消息进行广播
			user.DoMessage(msg)

			// 用户的任意消息, 表示当前用户活跃
			isAlive <- true
		}
	}()

	// 当前 handler 阻塞
	for {
		select {
		case <-isAlive:
			// 当前用户活跃, 激活 select, 更新计时器
		case <-time.After(time.Second * 60):
			// 超时, 强制关闭当前 User
			user.SendMsg("You are out ...")

			// 销毁资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出当前 Handler
			return // runtime.Goexit()
		}
	}
}

// 启动服务器接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听 Message 的 goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

	// close listen socket
}
