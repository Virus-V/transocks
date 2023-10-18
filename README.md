# 透明代理到Socks5

它可以把透明代理的流量转发到socks5代理服务器上。

### 使用
```
Usage of ./transocks_ng:
  -l string
        local listen address (default "127.0.0.1:8388")
  -p string
        proxy address (eg. socks5://192.168.1.100:1080)
```

### 防火墙配置（FreeBSD）
```shell
ipfw add 100 fwd 127.0.0.1,8388 tcp from any to any recv tun1
```
其中tun1是OpenVPN的虚拟网卡。比如在iPad/iPhone设备上安装OpenVPN，将所有tcp流量重定向到OpenVPN Server，再用本工具+防火墙实现大内外网分流。
