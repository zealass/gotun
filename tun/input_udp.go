package tun

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type inputUDP struct {
	addr string
	cfg  UDPConfig

	conn net.PacketConn

	streamHandler func(stream Stream)
}

func NewInputUDP(addr string, extra string) (*inputUDP, error) {
	var cfg UDPConfig

	if extra != "" {
		err := json.Unmarshal([]byte(extra), &cfg)
		if err != nil {
			return nil, err
		}
	}

	return &inputUDP{
		addr: addr,
		cfg:  cfg,
	}, nil
}

func (p *inputUDP) Run() error {
	conn, err := net.ListenPacket("udp", p.addr)
	if err != nil {
		return err
	}

	p.conn = conn
	go p.serve()
	return nil
}

func (p *inputUDP) Close() error {
	return p.conn.Close()
}

func (p *inputUDP) SetStreamHandler(f func(stream Stream)) {
	p.streamHandler = f
}

// send: client->relayer->sender->remote
// receive: client<-relayer<-sender<-remote
func (p *inputUDP) serve() {
	relayer := p.conn
	timeout := time.Duration(p.cfg.Timeout) * time.Second

	var workers WorkerMap
	for {
		buf := make([]byte, socketBufSize)
		n, srcAddr, err := relayer.ReadFrom(buf)
		if err != nil {
			logrus.WithError(err).Error("relayer.ReadFrom")
			continue
		}

		id := srcAddr.String()

		data := buf[0:n]

		worker := &UDPWorker{
			timeout:  timeout,
			srcAddr:  srcAddr,
			relayer:  relayer,
			bReaders: make(chan io.Reader, 100),
			onClear: func() {
				workers.Del(id)
			},
		}

		w, load := workers.LoadOrStore(id, worker)
		if !load {
			go func() {
				p.streamHandler(w)
			}()
		}

		w.insert(data)

		w.Logger().WithField("len", len(data)).Debug("client udp")
	}
}
