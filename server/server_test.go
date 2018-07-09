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
    "testing"
    "time"
)

type TelnetPacket struct {
    b []byte
}
func (p * TelnetPacket) Serialize() ([]byte) {
    return p.b
}

type Handler struct {
    t *testing.T
}

func (h * Handler) OnAccept(c *Connection) bool {
    c.SetReadDeadline(time.Now().Add(time.Second))
    c.SetWriteDeadline(time.Now().Add(time.Second))
    return true
}

func (h * Handler) OnMessage(c *Connection, packet Packet) bool {
    b := packet.Serialize()
    b = append(b, '\r')
    b = append(b, '\n')
    c.Write(b)
    return true
}

func (h * Handler) OnTimeout(c * Connection) bool {
    return false
}

func (h * Handler) OnClose(c *Connection) {
}

type TelnetProtocol struct {
}

func (p * TelnetProtocol) ReadPacket(c *Connection) (Packet, error) {
    scanner := bufio.NewScanner(c)
    if !scanner.Scan() {
        return nil, EOF
    }
    packet := &TelnetPacket{scanner.Bytes()}
    return packet, nil
}

func TestConn(t *testing.T) {
    message := "Hello World !\r\n"

    c := make(chan string)

    h := &Handler{t}
    p := &TelnetProtocol{}
    s := NewServer("0.0.0.0:3498", h, p)

    go func() {
        s.Start()
    }()
    go func() {
        time.Sleep(time.Millisecond * 2)
        conn, err := net.Dial("tcp", "127.0.0.1:3498")
        if err != nil {
            t.Fatal(err)
        }

        conn.Write([]byte(message))
        b := make([]byte, 20)
        _, err = conn.Read(b)
        if err != nil {
            t.Fatal(err)
        }
        conn.Close()
        s.Stop()
        c <- string(b[:13])
    }()
    m := <-c
    if m != message[:13] {
        t.Fatal(m, " != ", message[:13])
    }
}

func TestTimeout(t * testing.T) {
    message := "Hello World !\r\n"

    c := make(chan string)

    h := &Handler{t}
    p := &TelnetProtocol{}
    s := NewServer("0.0.0.0:3498", h, p)

    go func() {
        s.Start()
    }()
    go func() {
        time.Sleep(time.Millisecond * 2)
        conn, err := net.Dial("tcp", "127.0.0.1:3498")
        if err != nil {
            t.Fatal(err)
        }

        conn.Write([]byte(message))
        b := make([]byte, 20)
        _, err = conn.Read(b)
        if err != nil {
            t.Fatal(err)
        }
        s.Stop()
        c <- string(b[:13])
    }()
    m := <-c
    if m != message[:13] {
        t.Fatal(m, " != ", message[:13])
    }
}
