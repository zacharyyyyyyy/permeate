package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"permeate/util"
	"sync"
	"time"
)

const (
	gitlabPort     = ":8090"
	serverChanPort = ":8091"
	ConnectPort    = ":8092"
	serverHost     = "192.168.0.143"
	clientHost     = "192.168.0.143"
)

func readData(conn *net.TCPConn) error {
	reader := bufio.NewReader(conn)
	for {
		data, err := util.ClientDecode(reader) // 解码
		if errors.As(err, &io.EOF) {
			return err
		}
		if err != nil {
			return err
		}
		if data == "" {
			return errors.New("empty data")
		}
		fmt.Println("client get data:", data)
		if data == "KeepAlive" {
			fmt.Println("heartbeat!" + time.Now().Format("2006-01-02 15:04:05"))
		}
		if data == "New Connection" {
			//连接隧道转发
			fmt.Println("ready in iocopy!", time.Now().Format("2006-01-02 15:04:05"))
			connectConn := util.CreatConn(serverHost + ConnectPort)
			connectConn.SetKeepAlive(false)
			connectConn.SetNoDelay(false)
			connectConn.SetDeadline(time.Now().Add(1 * time.Second))
			gitlabConn := util.CreatConn(clientHost + gitlabPort)
			gitlabConn.SetKeepAlive(false)
			gitlabConn.SetNoDelay(false)
			gitlabConn.SetDeadline(time.Now().Add(1 * time.Second))
			fmt.Println("in iocopy!", time.Now().Format("2006-01-02 15:04:05"))
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				io.Copy(gitlabConn, connectConn)
			}()
			go func() {
				defer wg.Done()
				io.Copy(connectConn, gitlabConn)
			}()
			wg.Wait()
			gitlabConn.Close()
			connectConn.Close()
			fmt.Println("io copy close!", time.Now().Format("2006-01-02 15:04:05"))
		}
	}
	return nil
}

func main() {
	chanConn := util.CreatConn(serverHost + serverChanPort)
	defer chanConn.Close()
	fmt.Println("connect notice chan!")
	for {
		err := readData(chanConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("read err:" + err.Error())
			continue
		}
	}
	fmt.Println("client close!")
}
