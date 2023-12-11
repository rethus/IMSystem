package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前user channel消息的 goroutine
	go user.ListenMessage()

	return user
}

// 监听当前User channel，一有消息就发送给对应客户端
func (this *User) ListenMessage() {
	for msg := range this.C {
		n, err := this.conn.Write([]byte(msg + "\n"))
		if n == 0 {
			fmt.Println("conn close")
			return
		}
		if err != nil {
			fmt.Println("conn Write error:", err)
			return
		}
	}
}

// 用户上线业务
func (user *User) Online() {
	// 用户上线, 添加到OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 广播当前用户上线消息
	user.server.BroadCast(user, "online now")
}

// 用户下线业务
func (user *User) Offline() {
	// 用户下线, 从OnlineMap中删除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	// 广播当前用户下线消息
	user.server.BroadCast(user, "offline now")
}

// 发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

// 用户处理消息业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		user.server.mapLock.Lock()
		for _, _user := range user.server.OnlineMap {
			onlineMsg := "[" + _user.Addr + "] " + _user.Name + ": " + "is online ...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// rename|newName
		newName := strings.Split(msg, "|")[1]
		// 判断 newName 是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("new name: " + newName + " is been used ...\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("You has updated your name to \"" + user.Name + "\"\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// to|name|message content

		// 1. 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("wrong msg format ... please use \"to|name|message content\"\n")
			return
		}

		// 2. 根据用户名得到对方 user 对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("wrong username\n")
			return
		}

		// 3. 获取消息内容, 通过对方的 user 对象发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("mull message content ... please retry\n")
			return
		}

		remoteUser.SendMsg(user.Name + ": " + content + "\n")

	} else {
		user.server.BroadCast(user, msg)
	}
}
