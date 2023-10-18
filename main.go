package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var (
	listenAddr = flag.String("l", "127.0.0.1:8388", "local listen address")
	proxyAddr  = flag.String("p", "", "proxy address (eg. socks5://192.168.1.100:1080)")
)

var wg sync.WaitGroup

func main() {
	flag.Parse()

	fmt.Printf("Listen %s\n", *listenAddr)
	fmt.Printf("Proxy server:%s\n", *proxyAddr)

	ln, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		fmt.Println("listen failed!")
		return
	}

	/* set transparent proxy options */
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

	/* buf pool */
	bufPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	/* setup socks proxy dialer */
	proxyUrl, err := url.Parse(*proxyAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	proxyDialer, err := proxy.FromURL(proxyUrl, &net.Dialer{
		KeepAlive: 3 * time.Minute,
		DualStack: true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	/* proxy main loop */
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Listener.Accept error", err)
			continue
		}

		fmt.Println("Start proxy:", conn.LocalAddr().String())

		//destConn, err := net.Dial("tcp", conn.LocalAddr().String())
		destConn, err := proxyDialer.Dial("tcp", conn.LocalAddr().String())
		if err != nil {
			fmt.Println(err)
			continue
		}

		/* proxy dest write */
		go func(ctx context.Context) error {
			var written int64

			wg.Add(1)

			buf := bufPool.Get().([]byte)
			written, err := io.CopyBuffer(destConn, conn, buf)
			bufPool.Put(buf)

			conn.Close()
			wg.Done()
			fmt.Printf("dest write channel closed, written: %d Bytes\n", written)
			return err
		}(context.Background())

		/* proxy dest read */
		go func(ctx context.Context) error {
			var written int64

			wg.Add(1)

			buf := bufPool.Get().([]byte)
			written, err := io.CopyBuffer(conn, destConn, buf)
			bufPool.Put(buf)
			conn.Close()

			wg.Done()
			fmt.Printf("dest read channel closed, written: %d Bytes\n", written)
			return err
		}(context.Background())
	}

	/* TOOD: 优雅地退出 */
	wg.Wait()
}
