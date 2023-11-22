package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"permeate/util"
	"sync"
	"time"
)

const serverPort = ":8093"
const chanPort = ":8091"
const clientConnectPort = ":8092"

const localHost = "192.168.0.143"
const localToken = "123"

var (
	chanConn    *net.TCPConn
	connectChan = make(chan *net.TCPConn, 10)
	localCache  *cache.Cache
)

func KeepAlive(conn *net.TCPConn) {
	for {
		data, _ := util.ServerEncode("KeepAlive")
		_, err := conn.Write(data)
		if err != nil {
			log.Printf("[KeepAlive] Error %s", err)
			return
		}
		time.Sleep(time.Second * 10)
	}
}

func serverListen(serverConn *net.TCPConn) {
	//listener, err := createListen(localHost + serverPort)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println("server listening")
	//for {
	//	serverConn, err := listener.AcceptTCP()
	fmt.Println("new connection!")
	//if err != nil {
	//	log.Printf("接收连接失败，错误信息为：%s\n", err.Error())
	//	return
	//}
	data, _ := util.ServerEncode("New Connection")
	_, err := chanConn.Write(data)
	if err != nil {
		log.Printf("发送消息失败，错误信息为：%s\n", err.Error())
	}
	fmt.Println("notify connection")
	fmt.Println("in connect!", serverConn.RemoteAddr().String(), time.Now().Format("2006-01-02 15:04:05"))
	clientConnectConn := <-connectChan
	serverConn.SetKeepAlive(false)
	serverConn.SetNoDelay(false)
	serverConn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	data, _ = util.ServerEncode("New Connection")
	clientConnectConn.SetKeepAlive(false)
	clientConnectConn.SetNoDelay(false)
	clientConnectConn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// 数据转发
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(serverConn, clientConnectConn)
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConnectConn, serverConn)
	}()
	wg.Wait()
	serverConn.Close()
	clientConnectConn.Close()
	fmt.Println("connect close!", time.Now().Format("2006-01-02 15:04:05"))
	//}
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
	if r.Method == "GET" {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		_, ok := localCache.Get(ip)
		fmt.Println("GET ip:", ip, "token:", ok)
		if ok {
			hijacker, ok := w.(http.Hijacker)
			fmt.Println(ok)
			if !ok {
				return
			}
			conn, _, err := hijacker.Hijack()
			fmt.Println(err)
			defer conn.Close()
			serverListen(conn.(*net.TCPConn))
		} else {
			util.ShowPasswordFormHtml(w)
		}
	} else if r.Method == "POST" {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		_, ok := localCache.Get(ip)
		fmt.Println("POST ip:", ip, "token:", ok)
		if ok {
			hijacker, ok := w.(http.Hijacker)
			fmt.Println(ok)
			if !ok {
				return
			}
			conn, _, err := hijacker.Hijack()
			fmt.Println(err)
			defer conn.Close()
			serverListen(conn.(*net.TCPConn))
		} else {
			token := r.FormValue("token")
			token = html.EscapeString(token)
			if token == localToken {
				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				localCache.Set(ip, 1, 86400*time.Second)
			} else {
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
	localCache = cache.New(86400*time.Second, 86400*time.Second)
	go func() {
		http.HandleFunc("/", serverRoute)
		log.Fatal(http.ListenAndServe(serverPort, nil))
		defer wg.Done()
	}()
	go func() {
		chanListen()
		defer wg.Done()
	}()
	go func() {
		connectListen()
		defer wg.Done()
	}()
	wg.Wait()
}
