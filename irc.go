package irc

import (
	"os"
	"net"
	"log"
	"bufio"
	"strings"
)

type Conn struct {
	conn  *net.TCPConn
	read  chan string
	write chan string
}

func Dial(server string, nick string) (*Conn, os.Error) {
	ipAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, ipAddr)
	if err != nil {
		return nil, err
	}

	r := make(chan string, 100)
	w := make(chan string, 100)
	c := &Conn{conn: conn, read: r, write: w}

	// Reading task
	go func() {
		r := bufio.NewReader(conn)
		for {
			data, err := r.ReadString('\n')
			if err != nil {
				log.Println("Read error: ", err)
				return
			}
			if strings.HasPrefix(data, "PING") {
				c.write <- "PONG" + data[4:len(data)-2]
			} else {
				c.read <- data[0 : len(data)-2]
			}
		}
	}()

	// Writing task
	go func() {
		w := bufio.NewWriter(conn)
		for {
			data, ok := <-c.write
			if !ok {
				return
			}
			_, err := w.WriteString(data + "\r\n")
			if err != nil {
				log.Println("Write error: ", err)
			}
			w.Flush()
		}
	}()

	c.Write("NICK " + nick)
	c.Write("USER bot * * :...")

	return c, nil
}

func (c *Conn) Close() {
}

func (c *Conn) Write(data string) os.Error {
	c.write <- data
	return nil
}

func (c *Conn) Read() (string, os.Error) {
	// blocks until message is available
	data, ok := <-c.read
	if !ok {
		return "", os.NewError("Read stream closed")
	}
	return data, nil
}
