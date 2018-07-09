////////////////////////////////////////////////////////////////////////////////
// 
// (C) 2011 Kevin Druelle <kevin@druelle.info>
//
// this software is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This software is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this software.  If not, see <http://www.gnu.org/licenses/>.
// 
///////////////////////////////////////////////////////////////////////////////

package server

import(
    "net"
    "bufio"
    "sync"
    "time"
    "crypto/tls"
    "errors"
)

var(
    EOF = errors.New("End of file")
)

type Connection struct {
    conn        net.Conn
    server      *Server
    
    reader      *bufio.Reader
    writer      *bufio.Writer

    closeOnce   sync.Once
    closeChan   chan bool
}

func NewConnection(s * Server, c net.Conn) (*Connection) {
    conn := &Connection {
        conn:       c,
        server:     s,
        closeChan:  make(chan bool),
    }
    conn.reader = bufio.NewReader(conn.conn)
    conn.writer = bufio.NewWriter(conn.conn)
    return conn
}

func (c * Connection) Read(b []byte) (n int, err error) {
    return c.reader.Read(b)
}

func (c * Connection) Write(b []byte) (n int, err error) {
    n, err = c.writer.Write(b)
    if err != nil {
        return
    }
    err = c.writer.Flush()
    return
}

func (c * Connection) Close() {
    c.closeOnce.Do(func(){
        close(c.closeChan)
        c.server.handler.OnClose(c)
        c.conn.Close()
    })
}

func (c * Connection) IsClosed() (bool) {
    return false
}


func (c * Connection) SetReadDeadline(t time.Time) (err error) {
    err = c.conn.SetReadDeadline(t)
    return
}

func (c * Connection) SetWriteDeadline(t time.Time) (err error) {
    err = c.conn.SetWriteDeadline(t)
    return
}

func (c * Connection) SetDeadline(t time.Time) (err error) {
    err = c.conn.SetDeadline(t)
    return
}

func (c * Connection) IsSecure() (bool) {
    if _, ok := c.conn.(*tls.Conn); ok {
        return true
    }
    return false
}


func (c * Connection) RemoteAddrString() (s string) {
    if c.IsSecure() {
        s += "ssl:"
    } else {
        s += "tcp:"
    }
    s += c.conn.RemoteAddr().String()
    return
}


func (c * Connection) Do() {
    c.server.waitGroup.Add(1)
    defer c.server.waitGroup.Done()

    if !c.server.handler.OnAccept(c) {
        c.Close()
        return
    }
    c.handleRead()
}

func (c * Connection) handleRead() {
    defer c.Close()

    for {
        select {
        case <-c.closeChan:
            return
        default:
        }

        _, err := c.reader.Peek(1)
        if err != nil {
            if err, ok := err.(net.Error); ok && err.Timeout() {
                if c.server.handler.OnTimeout(c) {
                    continue
                }
            }
            return
        }
        p, err := c.server.protocol.ReadPacket(c)
        if err != nil {
            return
        }
        if !c.server.handler.OnMessage(c, p) {
            return
        }
    }
}


