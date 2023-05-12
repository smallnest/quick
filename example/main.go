package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/smallnest/quick"
)

func main() {
	// utils.SetLogLevel(utils.LogLevelDebug)

	startServer := flag.Bool("s", false, "server")
	startClient := flag.Bool("c", false, "client")
	flag.Parse()

	quicConfig := &quic.Config{
		MaxIdleTimeout: time.Minute,
	}
	if *startServer {
		// start the server
		go func() {
			tlsConf, err := generateTLSConfig()
			if err != nil {
				panic(err)
			}

			ln, err := quick.Listen("udp", ":8972", tlsConf, quicConfig)
			if err != nil {
				panic(err)
			}

			fmt.Println("Waiting for incoming connection")
			conn, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			fmt.Println("Established connection")

			for {
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					panic(err)
				}
				fmt.Print("Message from client: ", string(message))
				// echo back
				newmessage := strings.ToUpper(message)
				conn.Write([]byte(newmessage + "\n"))
			}
		}()
	}

	if *startClient {
		// run the client
		go func() {
			tlsConf := &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"quick"},
			}
			conn, err := quick.Dial("127.0.0.1:8972", tlsConf, quicConfig)
			if err != nil {
				panic(err)
			}

			for i := 0; ; i++ {
				message := "Ping from client #" + strconv.Itoa(i)
				fmt.Fprintf(conn, message+"\n")
				fmt.Printf("Sending message: %s\n", message)
				// listen for reply
				answer, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					panic(err)
				}
				fmt.Print("Message from server: " + answer)
				time.Sleep(time.Second)
			}

		}()
	}

	time.Sleep(time.Hour)
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM := pem.EncodeToMemory(&b)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quick"},
	}, nil
}
