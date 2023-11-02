package main

import (
	"net"
	"strings"
)
type User struct {
	Name string
	Addr string
	C    chan string //每个用户对应的管道
	conn net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}


//用户的上线业务
func (this *User) Online() {
	//用户上线，将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	
	//广播当前用户的上线消息
	this.server.BroadCast(this, "已上线")
}
//用户的下线业务
func (this *User) Offline() {
	//用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	
	//广播当前用户的上线消息
	this.server.BroadCast(this, "下线")
}
//给当前用户发送消息
func (this *User) SendMsg(msg string){
	this.conn.Write([]byte(msg))
}
//用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg=="who"{
		//查询当前在线用户

		this.server.mapLock.Lock()
		for _,user := range this.server.OnlineMap{
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	}else if len(msg)>7 && msg[:7]=="rename|"{
		//修改用户，格式：rename|name
		newName := strings.Split(msg,"|")[1]

		//判断name是否已经存在
		_,ok := this.server.OnlineMap[newName]
		if ok{
			this.SendMsg("当前用户已经存在\n")
		}else{
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			
			this.Name = newName
			this.SendMsg("您已经更新用户名" + this.Name + "\n")
		}

	}else if len(msg) > 4 && msg[:3] == "to|" {
		//格式：to|name|msg

		//获取用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式错误，请使用\"to|name|msg\"格式\n")
			return
		}
		//根据用户名，获取对方user对象
		remoteUser,ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户不存在\n")
			return
		}
		//获取消息内容，通过对方的user对象将消息发送
		content := strings.Split(msg,"|")[2]
		if content ==""{
			this.SendMsg("无消息内容，请重新输入\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说：" + content + "\n")

	}else{
		this.server.BroadCast(this, msg)
	}
}

// 监听当前User channel的方法，一旦有消息就发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
