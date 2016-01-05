// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package luna

import (
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"

	"github.com/polaris1119/goutils"
	"github.com/ugorji/go/codec"
)

const (
	HttpPortAdd = 100
)

const (
	EncodingTypeJson = iota
	EncodingTypeMsgpack
	EncodingTypeGob
)

const (
	JsonRpcPath, JsonDebugPath       = "/_rpc_json", "/debug/rpc_json"
	MsgpackRpcPath, MsgpackDebugPath = "/_rpc_msgpack", "/debug/rpc_msgpack"
	GobRpcPath, GobDebugPath         = "/_rpc_gob", "/debug/rpc_gob"
)

var (
	RpcPathMap = map[uint8]string{
		EncodingTypeJson:    JsonRpcPath,
		EncodingTypeMsgpack: MsgpackRpcPath,
		EncodingTypeGob:     GobRpcPath,
	}

	DebugPathMap = map[uint8]string{
		EncodingTypeJson:    JsonDebugPath,
		EncodingTypeMsgpack: MsgpackDebugPath,
		EncodingTypeGob:     GobDebugPath,
	}
)

func Register(services ...interface{}) error {
	return DefaultRpcServer.Register(services...)
}

func RegisterName(name string, rcvr interface{}) error {
	return DefaultRpcServer.RegisterName(name, rcvr)
}

func ListenAndServe(tcpAddr string) error {
	return DefaultRpcServer.ListenAndServe(tcpAddr)
}

var DefaultRpcServer = NewRpcServer()

type RpcServer struct {
	*rpc.Server

	// 基于 TCP 的 RPC，如 127.0.0.1:1234
	tcpAddr string
	// TODO:未实现
	// 基于 HTTP 的 RPC，如 127.0.0.1:1334
	// 如果没指定，默认端口是 tcpAddr端口 + 100
	httpAddr string

	// 数据编码类型
	encodingType uint8
}

func NewRpcServer() *RpcServer {
	return NewRpcServerWithEncoding(EncodingTypeJson)
}

func NewRpcServerWithEncoding(encodingType uint8) *RpcServer {
	return &RpcServer{
		Server:       rpc.NewServer(),
		encodingType: encodingType,
	}
}

func (r *RpcServer) ListenAndServe(tcpAddr string) error {
	tcpPort, host := tcpAddr, ""
	tcpAddrs := strings.Split(tcpAddr, ":")
	if len(tcpAddrs) == 2 {
		host = tcpAddrs[0]
		tcpPort = tcpAddrs[1]
	}

	httpPort := strconv.Itoa(goutils.MustInt(tcpPort) + HttpPortAdd)

	r.tcpAddr = tcpAddr
	r.httpAddr = host + ":" + httpPort

	go r.ListenTcpAndServe(tcpAddr)
	// r.ListenHttpAndServe(r.httpAddr)

	select {}
}

func (r *RpcServer) ListenTcpAndServe(tcpAddr string) error {
	if tcpAddr == "" {
		tcpAddr = r.tcpAddr
	}

	ln, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		return err
	}

	var conn net.Conn

	for {
		conn, err = ln.Accept()
		if err != nil {
			// TODO:
			continue
		}

		go r.serveCodec(conn)
	}

}

func (r *RpcServer) ListenHttpAndServe(httpAddr string) error {
	if httpAddr == "" {
		httpAddr = r.httpAddr
	}

	r.HandleHTTP(RpcPathMap[r.encodingType], DebugPathMap[r.encodingType])

	ln, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return err
	}

	go http.Serve(ln, nil)

	return nil
}

func (r *RpcServer) Register(services ...interface{}) error {
	var err error
	for _, service := range services {
		err = r.Server.Register(service)
	}

	return err
}

func (r *RpcServer) serveCodec(conn net.Conn) {
	switch r.encodingType {
	case EncodingTypeJson:
		r.serveJson(conn)
	case EncodingTypeMsgpack:
		r.serveMsgpack(conn)
	case EncodingTypeGob:
		r.ServeConn(conn)
	default:
		r.serveJson(conn)
	}
}

func (r *RpcServer) serveJson(conn net.Conn) {
	rpcCodec := jsonrpc.NewServerCodec(conn)
	r.ServeCodec(rpcCodec)
}

func (r *RpcServer) serveMsgpack(conn net.Conn) {
	rpcCodec := codec.MsgpackSpecRpc.ServerCodec(conn, new(codec.MsgpackHandle))
	r.ServeCodec(rpcCodec)
}
