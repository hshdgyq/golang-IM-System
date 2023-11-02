package main

import (
	"fmt"
	"net"
	"sync"
	"io"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message的goroutine，一旦有消息立即发送给所有用户
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		//将message发送给全部在线的user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
		cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//当前链接的业务
	//fmt.Println("业务启动成功")

	user := NewUser(conn, this)
	user.Online()
	//监听用户是否活跃的channel
	isLive := make(chan bool)
	//接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err:=conn.Read(buf)
			if n==0{
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:",err)
				return
			}

			//提取用户的消息（去除'\n'）
			msg:=string(buf[:n-1])

			//将得到的消息进行广播
			user.DoMessage(msg)
			//消息发送，可以确定用户活跃
			isLive <-true
		}
	}()

	//当前Handler的阻塞
	for {
		select {
		case <-isLive:
			//当前用户活跃，重置定时器
			//不做任何事情，为了触发select，更新定时器
		case <-time.After(time.Second*180)://time.After实际上是一个管道类型
			//已经超时，强制关闭当前user

			user.SendMsg("由于不活跃，你已被踢出")

			//销毁用户资源
			close(user.C)

			//关闭连接
			conn.Close()

			//退出Handler
			return //runtime.Goexit()
		}
	}

}

// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}

	//close socket listen
	defer listener.Close()
	//启动监听Message的goroutine
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}

	//
}
