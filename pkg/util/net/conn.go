// Copyright 2016 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"bytes"
	"context"
	"crypto/des"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"

	quic "github.com/quic-go/quic-go"

	"github.com/fatedier/frp/pkg/util/xlog"
)

var AesCipherKey = []byte("frp1234567890123")

type ContextGetter interface {
	Context() context.Context
}

type ContextSetter interface {
	WithContext(ctx context.Context)
}

func NewLogFromConn(conn net.Conn) *xlog.Logger {
	if c, ok := conn.(ContextGetter); ok {
		return xlog.FromContextSafe(c.Context())
	}
	return xlog.New()
}

func NewContextFromConn(conn net.Conn) context.Context {
	if c, ok := conn.(ContextGetter); ok {
		return c.Context()
	}
	return context.Background()
}

// ContextConn is the connection with context
type ContextConn struct {
	net.Conn

	ctx context.Context
}

func NewContextConn(ctx context.Context, c net.Conn) *ContextConn {
	return &ContextConn{
		Conn: c,
		ctx:  ctx,
	}
}

func (c *ContextConn) WithContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *ContextConn) Context() context.Context {
	return c.ctx
}

type WrapReadWriteCloserConn struct {
	io.ReadWriteCloser

	underConn net.Conn
}

func WrapReadWriteCloserToConn(rwc io.ReadWriteCloser, underConn net.Conn) net.Conn {
	return &WrapReadWriteCloserConn{
		ReadWriteCloser: rwc,
		underConn:       underConn,
	}
}

func (conn *WrapReadWriteCloserConn) LocalAddr() net.Addr {
	if conn.underConn != nil {
		return conn.underConn.LocalAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (conn *WrapReadWriteCloserConn) RemoteAddr() net.Addr {
	if conn.underConn != nil {
		return conn.underConn.RemoteAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (conn *WrapReadWriteCloserConn) SetDeadline(t time.Time) error {
	if conn.underConn != nil {
		return conn.underConn.SetDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (conn *WrapReadWriteCloserConn) SetReadDeadline(t time.Time) error {
	if conn.underConn != nil {
		return conn.underConn.SetReadDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (conn *WrapReadWriteCloserConn) SetWriteDeadline(t time.Time) error {
	if conn.underConn != nil {
		return conn.underConn.SetWriteDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

type CloseNotifyConn struct {
	net.Conn

	// 1 means closed
	closeFlag int32

	closeFn func()
}

// closeFn will be only called once
func WrapCloseNotifyConn(c net.Conn, closeFn func()) net.Conn {
	return &CloseNotifyConn{
		Conn:    c,
		closeFn: closeFn,
	}
}

func (cc *CloseNotifyConn) Close() (err error) {
	pflag := atomic.SwapInt32(&cc.closeFlag, 1)
	if pflag == 0 {
		err = cc.Close()
		if cc.closeFn != nil {
			cc.closeFn()
		}
	}
	return
}

type StatsConn struct {
	net.Conn

	closed     int64 // 1 means closed
	totalRead  int64
	totalWrite int64
	statsFunc  func(totalRead, totalWrite int64)
}

func WrapStatsConn(conn net.Conn, statsFunc func(total, totalWrite int64)) *StatsConn {
	return &StatsConn{
		Conn:      conn,
		statsFunc: statsFunc,
	}
}

// Pkcs5填充
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// des加密
func DesECBEncrypt(data, key []byte) ([]byte, error) {
	//NewCipher创建一个新的加密块
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	data = pkcs5Padding(data, bs)
	if len(data)%bs != 0 {
		return nil, errors.New("need a multiple of the blocksize")
	}

	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// des解密
func DesECBDecrypt(data, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = pkcs5UnPadding(out)
	return out, nil
}

func (statsConn *StatsConn) Read(p []byte) (n int, err error) {
	p, err = DesECBDecrypt(p, AesCipherKey)
	if err != nil {
		return 0, err
	}
	n, err = statsConn.Conn.Read(p)
	statsConn.totalRead += int64(n)
	return
}

func (statsConn *StatsConn) Write(p []byte) (n int, err error) {
	p, err = DesECBEncrypt(p, AesCipherKey)
	n, err = statsConn.Conn.Write(p)
	statsConn.totalWrite += int64(n)
	return
}

func (statsConn *StatsConn) Close() (err error) {
	old := atomic.SwapInt64(&statsConn.closed, 1)
	if old != 1 {
		err = statsConn.Conn.Close()
		if statsConn.statsFunc != nil {
			statsConn.statsFunc(statsConn.totalRead, statsConn.totalWrite)
		}
	}
	return
}

type wrapQuicStream struct {
	quic.Stream
	c quic.Connection
}

func QuicStreamToNetConn(s quic.Stream, c quic.Connection) net.Conn {
	return &wrapQuicStream{
		Stream: s,
		c:      c,
	}
}

func (conn *wrapQuicStream) LocalAddr() net.Addr {
	if conn.c != nil {
		return conn.c.LocalAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (conn *wrapQuicStream) RemoteAddr() net.Addr {
	if conn.c != nil {
		return conn.c.RemoteAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (conn *wrapQuicStream) Close() error {
	conn.Stream.CancelRead(0)
	return conn.Stream.Close()
}
