package tun

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func Test_TcpTunHead(t *testing.T) {
	targetAddr := "127.0.0.1:7007"
	startEchoServer(targetAddr)

	inHead := []byte("GET / HTTP/1.1\r\nHost:")
	outHead := []byte("GET / HTTP/1.1\r\nHost:")
	in, out, err := prepareHeadInOutCfg(inHead, outHead)
	if err != nil {
		t.Fatal(err)
		return
	}

	relayClientAddr := "127.0.0.1:6000"
	relayServerAddr := "127.0.0.1:6001"
	c, err := NewServer(Config{
		Name:          "tcp",
		Input:         fmt.Sprintf("tcp@%s", relayClientAddr),
		Output:        fmt.Sprintf("tcp@%s", relayServerAddr),
		InDecryptKey:  "",
		InDecryptMode: "",
		OutProtoCfg:   string(out),
		OutCryptKey:   "111111",
		OutCryptMode:  "gcm",
	})
	if err != nil {
		t.Fatal(err)
	}

	c.Run()

	s, err := NewServer(Config{
		Name:          "tcp",
		Input:         fmt.Sprintf("tcp@%s", relayServerAddr),
		Output:        fmt.Sprintf("tcp@%s", targetAddr),
		InProtoCfg:    string(in),
		InDecryptKey:  "111111",
		InDecryptMode: "gcm",
		OutCryptKey:   "",
		OutCryptMode:  "",
	})
	if err != nil {
		t.Fatal(err)
	}

	s.Run()
	time.Sleep(time.Second * 2)
	echo(t, relayClientAddr)
}

func Test_TcpMuxTunHead(t *testing.T) {
	targetAddr := "127.0.0.1:7007"
	startEchoServer(targetAddr)

	inHead := []byte("ABCDE")
	outHead := []byte("ABCDE")
	in, out, err := prepareHeadInOutCfg(inHead, outHead)
	if err != nil {
		t.Fatal(err)
		return
	}

	relayClientAddr := "127.0.0.1:6000"
	relayServerAddr := "127.0.0.1:6001"

	s, err := NewServer(Config{
		Name:          "tcp",
		Input:         fmt.Sprintf("tcp_mux@%s", relayServerAddr),
		Output:        fmt.Sprintf("tcp@%s", targetAddr),
		InProtoCfg:    string(in),
		InDecryptKey:  "111111",
		InDecryptMode: "gcm",
		OutCryptKey:   "",
		OutCryptMode:  "",
	})

	if err != nil {
		t.Fatal(err)
	}

	s.Run()

	c, err := NewServer(Config{
		Name:          "tcp",
		Input:         fmt.Sprintf("tcp@%s", relayClientAddr),
		Output:        fmt.Sprintf("tcp_mux@%s", relayServerAddr),
		InDecryptKey:  "",
		InDecryptMode: "",
		OutProtoCfg:   string(out),
		OutCryptKey:   "111111",
		OutCryptMode:  "gcm",
		OutExtend:     Extend{MuxConn: 1},
	})
	if err != nil {
		t.Fatal(err)
	}

	c.Run()

	time.Sleep(time.Second * 2)
	echo(t, relayClientAddr)
}

func prepareHeadInOutCfg(inHead []byte, outHead []byte) (string, string, error) {
	cfg := InProtoTCP{HeadTrim: inHead}
	in, err := json.Marshal(cfg)
	if err != nil {
		return "", "", err
	}

	outCfg := OutProtoTCP{HeadAppend: outHead}
	out, err := json.Marshal(outCfg)
	if err != nil {
		return "", "", err
	}

	return string(in), string(out), nil
}
