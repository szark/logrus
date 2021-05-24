package logrus2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Message represents the contents of the GELF message.  It is gzipped
// before sending.
type Message struct {
	Version   string                 `json:"version"`
	Host      string                 `json:"host"`
	Short     string                 `json:"short_message"`
	Full      string                 `json:"full_message,omitempty"`
	TimeUnix  float64                `json:"timestamp"`
	Level     float64                `json:"level,omitempty"`
	LevelName string                 `json:"level_name,omitempty"`
	Facility  string                 `json:"facility,omitempty"`
	Env       string                 `json:"environment,omitempty"`
	Service   string                 `json:"service,omitempty"`
	Extra     map[string]interface{} `json:"-"`
	RawExtra  json.RawMessage        `json:"-"`
}

func (m *Message) MarshalJSONBuf(buf *bytes.Buffer) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// write up until the final }
	if _, err = buf.Write(b[:len(b)-1]); err != nil {
		return err
	}
	if len(m.Extra) > 0 {
		eb, err := json.Marshal(m.Extra)
		if err != nil {
			return err
		}
		// merge serialized message + serialized extra map
		if err = buf.WriteByte(','); err != nil {
			return err
		}
		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(eb[1 : len(eb)-1]); err != nil {
			return err
		}
	}

	if len(m.RawExtra) > 0 {
		if err := buf.WriteByte(','); err != nil {
			return err
		}

		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(m.RawExtra[1 : len(m.RawExtra)-1]); err != nil {
			return err
		}
	}

	// write final closing quotes
	return buf.WriteByte('}')
}

func (m *Message) UnmarshalJSON(data []byte) error {
	i := make(map[string]interface{}, 16)
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	for k, v := range i {
		if k[0] == '_' {
			if m.Extra == nil {
				m.Extra = make(map[string]interface{}, 1)
			}
			m.Extra[k] = v
			continue
		}

		ok := true
		switch k {
		case "version":
			m.Version, ok = v.(string)
		case "host":
			m.Host, ok = v.(string)
		case "short_message":
			m.Short, ok = v.(string)
		case "full_message":
			m.Full, ok = v.(string)
		case "timestamp":
			m.TimeUnix, ok = v.(float64)
		case "level":
			var level float64
			level, ok = v.(float64)
			m.Level = float64(level)
		case "facility":
			m.Facility, ok = v.(string)
		}

		if !ok {
			return fmt.Errorf("invalid type for field %s", k)
		}
	}
	return nil
}

func (m *Message) toBytes(buf *bytes.Buffer) (messageBytes []byte, err error) {
	if err = m.MarshalJSONBuf(buf); err != nil {
		return nil, err
	}
	messageBytes = buf.Bytes()
	return messageBytes, nil
}

func constructMessage(p []byte, hostname string, facility string, file string, line int) (m *Message, err error) {
	//func constructMessage(p []byte) (m *Message) {
	// remove trailing and leading whitespace
	p = bytes.TrimSpace(p)

	// If there are newlines in the message, use the first line
	// for the short message and set the full message to the
	// original input.  If the input has no newlines, stick the
	// whole thing in Short.
	short := p
	full := []byte("")
	if i := bytes.IndexRune(p, '\n'); i > 0 {
		short = p[:i]
		full = p
	}

	gelf := make(map[string]interface{})
	if err := json.Unmarshal(short, &gelf); err != nil {
		fmt.Println(err)
	}

	if os.Getenv("HOST_HOSTNAME") == "" {
		os.Setenv("HOST_HOSTNAME", os.Getenv("HOSTNAME"))
	}

	// fmt.Println("Gelf: ", gelf)

	// for k, v := range gelf {
	// 	fmt.Println(k, " value is ", v)
	// 	fmt.Printf("And the value type was %T\n", v)
	// 	fmt.Sprintln(gelf[k])
	// 	// ok := true
	// 	// switch k {
	// 	// case "version":
	// 	// 	gelf[k] = v
	// 	// case "host":
	// 	// 	m.Host, ok = v
	// 	// case "short_message":
	// 	// 	m.Short, ok = v.(string)
	// 	// case "level_name":
	// 	// 	m.LevelName, ok = v.(string)
	// 	// case "timestamp":
	// 	// 	m.TimeUnix, ok = v.(float64)
	// 	// case "level":
	// 	// 	var level float64
	// 	// 	level, ok = v.(float64)
	// 	// 	m.Level = int32(level)
	// 	// case "facility":
	// 	// 	m.Facility, ok = v.(string)
	// 	// }

	// 	// if !ok {
	// 	// 	return nil, fmt.Errorf("invalid type for field %s", k)
	// 	// }
	// }
	// //fmt.Println("m.levelname", m.LevelName)
	// fmt.Println("Short m: ", string(short))
	// fmt.Println("Long m: ", string(full))
	m = &Message{
		Version:   "1.1",
		Host:      os.Getenv("HOST_HOSTNAME"),
		Short:     fmt.Sprintln(gelf["short_message"]),
		Full:      string(full),
		TimeUnix:  float64(time.Now().UnixNano()) / float64(time.Second),
		Level:     gelf["level"].(float64), // info
		LevelName: fmt.Sprintln(gelf["level_name"]),
		Env:       os.Getenv("APP_INSTANCE"),
		Service:   os.Getenv("SERVICE"),
		Facility:  facility,
		// Extra: map[string]interface{}{
		// 	"_file": file,
		// 	"_line": line,
		// },
	}
	//mt.Println(m)
	return m, err
}
