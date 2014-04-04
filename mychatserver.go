package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)
//用户信息
type User struct {
	 userName string
	 userAddr *net.UDPAddr
	 userListenConn *net.UDPConn
	 chatToConn *net.UDPConn
}

//服务器监听端口
const LISTENPORT = 1616
//缓冲区
const BUFFSIZE = 1024
var buff = make([]byte, BUFFSIZE)
//在线用户
var onlineUser = make([]User, 0)
//在线状态判断缓冲区
var onlineCheckAddr = make([]*net.UDPAddr, 0)

//错误处理
func HandleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
//消息处理
func HandleMessage(udpListener *net.UDPConn) {
	n, addr, err := udpListener.ReadFromUDP(buff)
	HandleError(err)

	if n > 0 {
		msg := AnalyzeMessage(buff, n)
		
		switch msg[0] {
			//连接信息
			case "connect  ":
				//获取昵称+端口
				userName := msg[1]
				userListenPort := msg[2]
				//获取用户ip
				ip := AnalyzeMessage([]byte(addr.String()), len(addr.String()))
				//显示登录信息
				fmt.Println(" 昵称:", userName, " 地址:", ip[0], " 用户监听端口:", userListenPort, " 登录成功！")
				//创建对用户的连接，用于消息转发
				userAddr, err := net.ResolveUDPAddr("udp4", ip[0] + ":" + userListenPort)
				HandleError(err)
				
				userConn, err := net.DialUDP("udp4", nil, userAddr)
				HandleError(err)
				
				//因为连接要持续使用，不能在这里关闭连接
				//defer userConn.Close()
				//添加到在线用户
				onlineUser = append(onlineUser, User{userName, addr, userConn, nil})
				
			case "online   ":
				//收到心跳包
				onlineCheckAddr = append(onlineCheckAddr, addr)
				
			case "outline  ":
				//退出消息，未实现
			case "chat     ":
				//会话请求
				//寻找请求对象
				index := -1
				for i := 0; i < len(onlineUser); i++ {
					if onlineUser[i].userName == msg[1] {
						index = i
					}
				}
				//将所请求对象的连接添加到请求者中
				if index != -1 {
					nowUser, _ := FindUser(addr)
					onlineUser[nowUser].chatToConn = onlineUser[index].userListenConn
				}
			case "get      ":
				//向请求者返回在线用户信息
				index, _ := FindUser(addr)
				onlineUser[index].userListenConn.Write([]byte("当前共有" + strconv.Itoa(len(onlineUser)) + "位用户在线"))
				for i, v := range onlineUser {
					onlineUser[index].userListenConn.Write([]byte("" + strconv.Itoa(i + 1) + ":" + v.userName))
				}
			default:
				//消息转发
				//获取当前用户
				index, _ := FindUser(addr)
				//获取时间
				nowTime := time.Now()
				nowHour := strconv.Itoa(nowTime.Hour())
				nowMinute := strconv.Itoa(nowTime.Minute())
				nowSecond := strconv.Itoa(nowTime.Second())
				//请求会话对象是否存在
				if onlineUser[index].chatToConn == nil {
					onlineUser[index].userListenConn.Write([]byte("对方不在线"))
				} else {
					onlineUser[index].chatToConn.Write([]byte(onlineUser[index].userName + " " + nowHour + ":" + nowMinute + ":" + nowSecond + "\n" + msg[0]))
				}
				
		}
	}
}
//消息解析，[]byte -> []string
func AnalyzeMessage(buff []byte, len int) ([]string) {
	analMsg := make([]string, 0)
	strNow := ""
	for i := 0; i < len; i++ {
		if string(buff[i:i + 1]) == ":" {
			analMsg = append(analMsg, strNow)
			strNow = ""
		} else {
			strNow += string(buff[i:i + 1])
		}
	}
	analMsg = append(analMsg, strNow)
	return analMsg
}
//寻找用户，返回（位置，是否存在）
func FindUser(addr *net.UDPAddr) (int, bool) {
	alreadyhave := false
	index := -1
	for i := 0; i < len(onlineUser); i++ {
		
		if onlineUser[i].userAddr.String() == addr.String() {
			alreadyhave = true
			index = i
			break
		}
	}
	return index, alreadyhave
}
//处理用户在线信息（暂时仅作删除用户使用）
func HandleOnlineMessage(addr *net.UDPAddr, state bool) {
	index, alreadyhave := FindUser(addr)
	if state == false {
		if alreadyhave {
			onlineUser = append(onlineUser[:index], onlineUser[index + 1:len(onlineUser)]...) 
		}
	}
}
//在线判断，心跳包处理，每5s查看一次所有已在线用户状态
func OnlineCheck() {
	for {
		onlineCheckAddr = make([]*net.UDPAddr, 0)
		sleepTimer := time.NewTimer(time.Second * 5)
		<- sleepTimer.C
		for i := 0; i < len(onlineUser); i++ {
			haved := false
			FORIN:for j := 0; j < len(onlineCheckAddr); j++ {
				if onlineUser[i].userAddr.String() == onlineCheckAddr[j].String() {
					haved = true
					break FORIN
				}
			}
			if !haved {
				fmt.Println(onlineUser[i].userAddr.String() + "退出！")
				HandleOnlineMessage(onlineUser[i].userAddr, false)
				i--
			}

		}
	}
}

func main() {
	//监听地址
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + strconv.Itoa(LISTENPORT))
	HandleError(err)
	//监听连接
	udpListener, err := net.ListenUDP("udp4", udpAddr)
	HandleError(err)

	defer udpListener.Close()

	fmt.Println("开始监听：")

	//在线状态判断
	go OnlineCheck()

	for {
		//消息处理
		HandleMessage(udpListener)
	}

}
