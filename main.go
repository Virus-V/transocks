package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	listenAddr = flag.String("l", "127.0.0.1:8388", "local listen address")
	proxyAddr  = flag.String("p", "", "proxy address")
)

var wg sync.WaitGroup

func main() {
	fmt.Println("hello world")

	fmt.Printf("listen %s\n", *listenAddr)
	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		fmt.Println("listen failed!")
		return
	}

	lntcp := ln.(*net.TCPListener)
	sysConn, err := lntcp.SyscallConn()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = SetTransparentListener(sysConn)
	if err != nil {
		fmt.Println(err)
		return
	}

	bufPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Listener.Accept error", err)
			continue
		}

		fmt.Println("new connect addr ", conn.LocalAddr().String())
		destConn, err := net.Dial("tcp", conn.LocalAddr().String())
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func(ctx context.Context) error {
			var written int64
			wg.Add(1)
			buf := bufPool.Get().([]byte)
			written, err := io.CopyBuffer(destConn, conn, buf)
			bufPool.Put(buf)
			conn.Close()
			wg.Done()
			fmt.Printf("local to remote: %d\n", written)
			return err
		}(context.Background())

		go func(ctx context.Context) error {
			var written int64
			wg.Add(1)
			buf := bufPool.Get().([]byte)
			written, err := io.CopyBuffer(conn, destConn, buf)
			bufPool.Put(buf)
			conn.Close()
			wg.Done()
			fmt.Printf("remote to local: %d\n", written)
			return err
		}(context.Background())
	}
	wg.Wait()
}
