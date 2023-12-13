package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"permeate/util"
	"strings"
	"sync"
	"time"
)

const (
	serverPort        = ":8093"
	chanPort          = ":8091"
	clientConnectPort = ":8092"          //后续client多进程需新增端口
	localHost         = ""               //server host 按实际部署host修改
	localToken        = ""               //校验权限时输的token
	defaultDeadLine   = 3 * time.Second  //默认超时时间
	gitDeadLine       = 5 * time.Second  //带.git的请求超时时间
	keepAliveTime     = 10 * time.Second //心跳包间隔
)

var (
	chanConn    *net.TCPConn
	connectChan = make(chan *net.TCPConn, 10)
	msgChan     = make(chan struct{}, 2)
)

func KeepAlive(conn *net.TCPConn) {
	for {
		data, _ := util.ServerEncode("KeepAlive")
		_, err := conn.Write(data)
		if err != nil {
			log.Printf("[KeepAlive] Error %s", err)
			return
		}
		time.Sleep(keepAliveTime)
	}
}

func serverListen(w http.ResponseWriter, r *http.Request, isGitMethod bool) {
	keepAliveTime := defaultDeadLine
	if isGitMethod {
		keepAliveTime = gitDeadLine
	}
	htmlData := util.GetHtmlDataPackage(r)
	hijacker, ok := w.(http.Hijacker)
	fmt.Println("hijack conn:", ok)
	if !ok {
		return
	}
	conn, _, err := hijacker.Hijack()
	fmt.Println("hijack err:", err)
	defer conn.Close()
	fmt.Println("new connection!")
	serverConn := conn.(*net.TCPConn)
	data, _ := util.ServerEncode("New Connection")
	_, err = chanConn.Write(data)
	if err != nil {
		log.Printf("发送消息失败，错误信息为：%s\n", err.Error())
	}
	fmt.Println("notify connection")
	fmt.Println("in connect!", serverConn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"))
	clientConnectConn := <-connectChan
	serverConn.SetKeepAlive(false)
	serverConn.SetNoDelay(false)
	serverConn.SetDeadline(time.Now().Add(keepAliveTime))
	clientConnectConn.SetKeepAlive(false)
	clientConnectConn.SetNoDelay(false)
	clientConnectConn.SetDeadline(time.Now().Add(keepAliveTime))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 数据转发
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		msgChan <- struct{}{}
		_, err := io.Copy(serverConn, clientConnectConn)
		if err != nil {
			fmt.Println("79:", err.Error())
		}
	}()
	go func() {
		defer wg.Done()
		msgChan <- struct{}{}
		_, err := io.Copy(clientConnectConn, serverConn)
		if err != nil {
			fmt.Println("86:", err.Error())
		}
	}()
	go func() {
		<-msgChan
		<-msgChan
		fmt.Println(clientConnectConn.Write([]byte(htmlData)))
		fmt.Println("write data")
	}()
	wg.Wait()
	serverConn.Close()
	clientConnectConn.Close()
	fmt.Println("connect close!", time.Now().Format("2006-01-02 15:04:05"))
}

func chanListen() {
	listener, err := util.CreateListen(localHost + chanPort)
	if err != nil {
		panic(err)
	}
	for {
		chanConn, err = listener.AcceptTCP()
		if err != nil {
			log.Printf("接收连接失败，错误信息为：%s\n", err.Error())
			return
		}
		//保持连接
		go KeepAlive(chanConn)
	}
}

func connectListen() {
	listener, err := util.CreateListen(localHost + clientConnectPort)
	if err != nil {
		panic(err)
	}
	for {
		clientConnectConn, err := listener.AcceptTCP()
		fmt.Println("get connect:", clientConnectConn.RemoteAddr().String())
		if err != nil {
			log.Printf("接收连接失败，错误信息为：%s\n", err.Error())
			return
		}
		connectChan <- clientConnectConn
	}
}

func serverRoute(w http.ResponseWriter, r *http.Request) {

	sessionCtl := util.NewSession(w, r)
	if !util.IpRejector.Pass(sessionCtl.GetSid()) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("IP has been restricted"))
		return
	}
	isGitMethod := strings.Contains(r.URL.Path, ".git")
	fmt.Println("get connect!", isGitMethod)
	if isGitMethod {
		serverListen(w, r, isGitMethod)
	} else if r.Method == http.MethodGet {
		ok := sessionCtl.HasAuth()
		if ok {
			serverListen(w, r, isGitMethod)
		} else {
			util.ShowPasswordFormHtml(w)
		}
	} else if r.Method == http.MethodPost {
		ok := sessionCtl.HasAuth()
		if ok {
			serverListen(w, r, isGitMethod)
		} else {
			token := r.FormValue("token")
			token = html.EscapeString(token)
			if token == localToken {
				sessionCtl.Login()
			} else {
				util.IpRejector.AddErrorTimes(sessionCtl.GetSid())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("wrong token"))
			}
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed) // 不允许其他HTTP方法
	}
}

func main() {
	var wg sync.WaitGroup
	fmt.Println("server start")
	wg.Add(3)
	go func() {
		defer wg.Done()
		http.HandleFunc("/", serverRoute)
		log.Fatal(http.ListenAndServe(serverPort, nil))
	}()
	go func() {
		defer wg.Done()
		chanListen()
	}()
	go func() {
		defer wg.Done()
		connectListen()
	}()
	wg.Wait()
}
