package main

import (
	"net/http"
	"net"
	"bufio"
	"strings"
	"bytes"
)
func ProxyRemoteHandle(c net.Conn,r net.Conn){
	buf:=make([]byte,65536)
	for{
		n,ReadErr:=r.Read(buf)
		if ReadErr!=nil{
			c.Close()
			return
		}
		c.Write(buf[:n])
	}
}
func ProxyHandle(c net.Conn){
	IsConnected:=false
	P:=""
	var Remote *net.Conn
	defer func(){
		c.Close()
		if Remote!=nil{
			(*Remote).Close()
		}
	}()
	buf:=make([]byte,65536)
	for{
		n,ReadErr:=c.Read(buf)
		if ReadErr!=nil{
			return
		}
		if IsConnected{
			if Remote==nil{
				return
			}
			if P=="http"{
				r:=bufio.NewReader(bytes.NewReader(buf[:n]))
				_,ReadRequestErr:=http.ReadRequest(r)
				if ReadRequestErr==nil{
					(*Remote).Close()
					goto reconnect
				}
			}
			(*Remote).Write(buf[:n])

			continue
		}
			reconnect:
			r:=bufio.NewReader(bytes.NewReader(buf[:n]))
			rq,ReadRequestErr:=http.ReadRequest(r)
			if ReadRequestErr!=nil{
				return
			}
			address:=rq.Host
			if strings.Index(address,":") ==-1{
					address+=":80"
			}
			tmpConn,DialErr:=net.Dial("tcp",address)
			if rq.Method=="CONNECT"{
				c.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
				P="https"
			}
			if DialErr!=nil{
				return
			}
			if rq.Method!="CONNECT"{
				newbuf:=bytes.Replace(buf[:n],[]byte(" http://"+rq.Host),[]byte(" "),1)
				tmpConn.Write(newbuf)
				P="http"
			}
			Remote=&tmpConn
			IsConnected=true
			go ProxyRemoteHandle(c,*Remote)
	}
}
func main(){
	l,_:=net.Listen("tcp",":10800")
	for {
		conn,_:=l.Accept()
		go ProxyHandle(conn)
	}
}
