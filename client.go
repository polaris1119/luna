// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package luna

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/ugorji/go/codec"
)

type Client struct {
	*rpc.Client
	encodingType uint8
}

func NewDefaultClient() *Client {
	return &Client{}
}

func NewClientWithEncoding(encodingType uint8) *Client {
	return &Client{
		encodingType: encodingType,
	}
}

func Dial(network, address string, encodingType uint8) (*Client, error) {
	client := NewClientWithEncoding(encodingType)
	err := client.Dial(network, address)

	return client, err
}

func DialTCP(address string, encodingType uint8) (*Client, error) {
	return Dial("tcp", address, encodingType)
}

// func DialHTTP(address string, encodingType uint8) (*Client, error) {
// 	client := NewClientWithEncoding(encodingType)

// 	var err error
// 	conn, err := net.Dial("tcp", address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	io.WriteString(conn, "CONNECT "+RpcPathMap[encodingType]+" HTTP/1.0\n\n")

// 	// Require successful HTTP response
// 	// before switching to RPC protocol.
// 	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
// 	if err == nil && resp.Status == connected {
// 		client.fillRpcClient(conn)
// 		return client, nil
// 	}
// 	if err == nil {
// 		err = errors.New("unexpected HTTP response: " + resp.Status)
// 	}
// 	conn.Close()
// 	return nil, &net.OpError{
// 		Op:   "dial-http",
// 		Net:  "tcp " + address,
// 		Addr: nil,
// 		Err:  err,
// 	}
// }

func DialTimeout(address string, encodingType uint8, timeout time.Duration) (*Client, error) {
	client := NewClientWithEncoding(encodingType)
	err := client.DialTimeout("tcp", address, timeout)

	return client, err
}

func (c *Client) Dial(network, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}

	c.fillRpcClient(conn)

	return nil
}

func (c *Client) DialTimeout(network, address string, timeout time.Duration) error {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return err
	}

	c.fillRpcClient(conn)

	return nil
}

func (c *Client) fillRpcClient(conn net.Conn) {

	switch c.encodingType {
	case EncodingTypeJson:
		c.Client = jsonrpc.NewClient(conn)
	case EncodingTypeMsgpack:
		rpcCodec := codec.MsgpackSpecRpc.ClientCodec(conn, new(codec.MsgpackHandle))
		c.Client = rpc.NewClientWithCodec(rpcCodec)
	case EncodingTypeGob:
		c.Client = rpc.NewClient(conn)
	default:
		c.Client = jsonrpc.NewClient(conn)
	}
}
