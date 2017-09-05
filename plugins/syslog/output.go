package syslog

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"text/template"
	"time"

	"github.com/coldog/logship/input"
)

const (
	FormatRFC5424 = "rfc5424"
	FormatRFC3164 = "rfc3164"
	DefaultFormat = FormatRFC3164
)

type Output struct {
	input.BaseInput

	Format   string
	Hostname string
	Network  string
	Address  string
	TLS      bool

	conn net.Conn
	tpl  *template.Template
}

func (o *Output) Open() error {
	if o.Hostname == "" {
		o.Hostname, _ = os.Hostname()
	}
	if o.Format == "" {
		o.Format = DefaultFormat
	}
	if o.Format != FormatRFC3164 && o.Format != FormatRFC5424 {
		return fmt.Errorf("syslog: format unrecognized: %s", o.Format)
	}
	err := o.connect()
	if err != nil {
		return err
	}
	return o.BaseInput.Open()
}

func (o *Output) Run(out chan *input.Message) {
	defer o.Finish()

	for {
		select {
		case m := <-out:
			var value string
			switch o.Format {
			case FormatRFC3164:
				value = o.format3164(m)
			case FormatRFC5424:
				value = o.format5424(m)
			}

			_, err := o.conn.Write([]byte(value))
			if err != nil {
				o.reconnect()
			}
		case <-o.Done():
			return
		}
	}
}

func (o *Output) connect() error {
	if o.conn != nil {
		o.conn.Close()
	}
	if o.TLS {
		conn, err := tls.Dial(o.Network, o.Address, nil)
		if err != nil {
			return err
		}
		o.conn = conn
	} else {
		conn, err := net.Dial(o.Network, o.Address)
		if err != nil {
			return err
		}
		o.conn = conn
	}
	return nil
}

func (o *Output) format5424(m *input.Message) string {
	return "<6>1 " + m.Time.Format(time.RFC3339) +
		" " + o.Hostname + " " + m.Tag + " " + m.Data["pid"] +
		" - " + m.Data["message"] + "\n"
}

func (o *Output) format3164(m *input.Message) string {
	return "<6>" + m.Time.Format(time.RFC3339) + " " + o.Hostname + " " +
		m.Tag + "[" + m.Data["pid"] + "]: " + m.Data["message"] + "\n"
}

func (o *Output) reconnect() {
	try := uint(0)
	for {
		err := o.connect()
		if err == nil {
			return
		}
		try++
		select {
		case <-o.Done():
			return
		case <-time.After((1 << try) * 10 * time.Millisecond):
		}
	}
}
