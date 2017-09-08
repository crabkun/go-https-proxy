package main

import (
	"net/http"
	"net"
	"bufio"
	"bytes"
	"regexp"
	"log"
)
func HTTPProxyHandle(c net.Conn){
	proto:="http"
	defer func(){
		c.Close()
	}()
	var err error
	//解析http代理数据包
	r:=bufio.NewReader(c)
	rq,err:=http.ReadRequest(r)
	if err!=nil{
		return
	}
	reconnect:
	if rq.Method=="CONNECT"{
		proto="https"
	}
	host:=rq.Host
	if b,_:=regexp.MatchString("^.+:[0-9]+$",rq.Host);!b{
		if proto=="http"{
			host=rq.Host+":80"
		}else{
			host=rq.Host+":443"
		}
	}

	//呼叫代理服务器
	remote,err:=net.Dial("tcp",host)
	if err!=nil{
		return
	}
	log.Println("connect to",host)
	if proto=="https"{
		//特殊处理https代理客户端
		c.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	}else{
		rq.Write(remote)
	}

	//IO桥：代理服务器到客户端
	go func(remote net.Conn){
		defer func(){
			remote.Close()
		}()
		buf:=make([]byte,40960)
		for{
			n,err:=remote.Read(buf)
			if err!=nil{
				return
			}
			c.Write(buf[:n])
		}
	}(remote)

	//IO桥：客户端到代理服务器
	if rq=func(remote net.Conn) *http.Request{
		buf:=make([]byte,40960)
		for{
			n,err:=c.Read(buf)
			if err!=nil{
				return nil
			}
			if proto=="http"{
				if nrq:=IsHTTPpacket(buf[:n]);nrq!=nil{
					return nrq
				}
			}
			remote.Write(buf[:n])
		}
	}(remote);rq!=nil{
		goto reconnect
	}

}
func IsHTTPpacket(buf []byte) *http.Request{
	r:=bufio.NewReader(bytes.NewReader(buf))
	req,ReadRequestErr:=http.ReadRequest(r)
	if ReadRequestErr!=nil{
		return nil
	}
	return req
}
func main(){
	l,err:=net.Listen("tcp",":8080")
	if err!=nil{
		log.Println(err)
		return
	}
	log.Println("listen at",l.Addr())
	for {
		conn,err:=l.Accept()
		if err!=nil{
			log.Println(err.Error())
			return
		}
		go HTTPProxyHandle(conn)
	}
}

