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
    "crypto/tls"
    "sync"
    "github.com/kdruelle/golib/workerpool"
)


type Server struct {
    localAddr       string
    handler         ConnectionHandler
    protocol        Protocol
    
    ssl             bool
    sslCert         string
    sslKey          string

    listener      net.Listener
    stop          bool
    pool        * workerpool.WorkerPool
    
    laddr       * net.TCPAddr

    waitGroup   *sync.WaitGroup
}

func NewServer(localAddr string, handler ConnectionHandler, protocol Protocol) (*Server){
    server := &Server{
        localAddr  : localAddr,
        handler    : handler,
        protocol   : protocol,
        stop       : false,
        pool       : workerpool.NewWorkerPool(5, 100),
        ssl        : false,
        waitGroup  : &sync.WaitGroup{},
    }
    server.pool.Run()
    return server
}

func (s * Server) Start() (err error){
    s.laddr, err = net.ResolveTCPAddr("tcp", s.localAddr)
    if err != nil {
        return
    }
    if s.ssl {
        err = s.ListenAndServeTLS()
    } else {
        err = s.listenAndServe()
    }
    return
}

func (s * Server) Stop(){
    s.stop = true
    s.listener.Close()
    s.pool.Stop()
    s.waitGroup.Wait()
}

func (s * Server) listenAndServe() (err error){
    s.listener, err = net.ListenTCP("tcp", s.laddr)
    if err != nil {
        return
    }
    s.serve()
    return
}

func (s * Server) ListenAndServeTLS() (err error) {
    cer, err := tls.LoadX509KeyPair(s.sslCert, s.sslKey)
    if err != nil {
        return
    }
    config := &tls.Config{
        Certificates  : []tls.Certificate{cer},
        MinVersion    : tls.VersionSSL30,
    }

    s.listener, err = tls.Listen("tcp", s.laddr.String(), config)
    if err != nil {
        return
    }
    s.serve()
    return
}

func (s * Server) serve() {
    for {
        conn, err := s.listener.Accept()
        if err != nil {
            if s.stop {
                return
            }
            continue
        }
        c := NewConnection(s, conn)
        s.pool.Handle(c)
    }
}

