// Package mbox parses the mbox file format.
package mbox

// This code was adapted from https://github.com/bytbox/slark, but it was
// packaged as an app, not a library. Also, we switched to the stdlib mail
// parser.

// note: this is a modified version - Waitman Gobble <ns@waitman.net> - save original message content 12/24/15

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/mail"
	"os"
	"crypto/sha256"
	"encoding/hex"
)

const _MAX_LINE_LEN = 1024

var crlf = []byte{'\r', '\n'}

func nest(s string) string {
        hash := sha256.New()
        hash.Write([]byte(s))
        mdStr := hex.EncodeToString(hash.Sum(nil))
        nest := mdStr[:1] + "/" + mdStr[1:2] + "/" + mdStr[2:3] + "/" + mdStr[3:4] + "/" + mdStr[4:5] + "/" + mdStr[5:6] + "/" + mdStr[6:len(mdStr)]
        return nest
}

// If debug is true, errors parsing messages will be printed to stderr. If
// false, they will be ignored. Either way those messages will not appear in
// the msgs slice.
func Read(r io.Reader, path string, debug bool) (msgs []*mail.Message, err error) {
	var mbuf *bytes.Buffer
	lastblank := true
	br := bufio.NewReaderSize(r, _MAX_LINE_LEN)
	l, _, err := br.ReadLine()
	for err == nil {
		fs := bytes.SplitN(l, []byte{' '}, 3)
		if len(fs) == 3 && string(fs[0]) == "From" && lastblank {
			// flush the previous message, if necessary
			if mbuf != nil {
				msgs = parseAndAppend(mbuf, msgs, path, debug)
			}
			mbuf = new(bytes.Buffer)
		} else {
			_, err = mbuf.Write(l)
			if err != nil {
				return
			}
			_, err = mbuf.Write(crlf)
			if err != nil {
				return
			}
		}
		if len(l) > 0 {
			lastblank = false
		} else {
			lastblank = true
		}
		l, _, err = br.ReadLine()
	}
	if err == io.EOF {
		msgs = parseAndAppend(mbuf, msgs, path, debug)
		err = nil
	}
	return
}

// If debug is true, errors parsing messages will be printed to stderr. If
// false, they will be ignored. Either way those messages will not appear in
// the msgs slice.
func ReadFile(filename string, path string, debug bool) ([]*mail.Message, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	msgs, err := Read(f, path, debug)
	f.Close()
	return msgs, err
}

func parseAndAppend(mbuf *bytes.Buffer, msgs []*mail.Message, path string, debug bool) []*mail.Message {
	mbufx := bytes.NewBuffer(nil)
	mbufx.ReadFrom(mbuf)
	f,_ := os.Create("/tmp/orig")
        mbuf.WriteTo(f)
	f.Close()
	msg, err := mail.ReadMessage(mbufx)
	header := msg.Header
	filepath := nest(header.Get("Message-Id"))
	os.MkdirAll(path+"/"+filepath,0777);
	os.Link("/tmp/orig", path+"/"+filepath+"/orig")
	if err != nil {
		if debug {
			log.Print(err)
		}
		return msgs // don't append
	}
	return append(msgs, msg)
}
