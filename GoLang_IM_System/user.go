package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server //记录用户属于哪个server
}

//创建用户的API
func NewUser(con net.Conn, serv *Server) *User {
	userAddr := con.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: con,

		server: serv,
	}

	go user.ListenMessage()
	return user
}

//用户上线业务
func (this *User) Online() {
	//将用户加到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")
}

//用户下线业务
func (this *User) Offline() {
	//将用户从OnlineMap删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线\n")
}

//给当前user对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[0:7] == "rename|" { //修改名字
		//消息格式："rename|zhang3"
		newName := strings.Split(msg, "|")[1] //1表示|之后的部分
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用.\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 "to|张三|消息内容"
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("格式不对，请使用to|张三|消息内容\n")
		}
		this.server.mapLock.Lock()
		user, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("用户名不存在\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("内容为空\n")
			return
		}
		user.SendMsg(this.Name + "对您说:" + content + "\n")

		this.server.mapLock.Unlock()
	} else {
		this.server.BroadCast(this, msg)
	}
}

//监听user channel的go，一旦有消息，就发给user
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		//发确认消息
		this.conn.Write([]byte(msg + "\n"))
	}
}
