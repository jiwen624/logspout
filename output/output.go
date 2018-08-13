package output

import (
	"encoding/json"

	"github.com/jiwen624/logspout/log"
)

type Output interface {
	Write(string) error
}

func BuildOutputMap(om map[string]map[string]interface{}) map[string]Output {
	op := map[string]Output{}
	for k, v := range om {
		ov := BuildOutpout(v)
		if ov != nil {
			op[k] = ov
		} else {
			log.Warnf("Invalid output: %s", k)
		}
	}
	return op
}

func BuildOutpout(m map[string]interface{}) Output {
	if m == nil {
		return nil
	}

	t, ok := m["type"]
	if !ok {
		log.Warn("Type is missing.")
		return nil
	}

	typ, ok := t.(string)
	if !ok {
		log.Warn("Type is not a string.")
		return nil
	}

	attrs, ok := m["attrs"]
	if !ok {
		log.Warn("Attrs is missing.")
		return nil
	}

	var op Output

	switch typ {
	case "console":
		cop := &Console{}
		if err := json.Unmarshal(attrs.([]byte), cop); err != nil {
			return nil
		}
	case "file":
		cop := &File{}
		if err := json.Unmarshal(attrs.([]byte), cop); err != nil {
			return nil
		}

	case "syslog":
	case "kafka":
	default:
		log.Warnf("Unknown output type: %s", typ)
		return nil
	}

	return op
}
