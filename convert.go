package json2xml

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strconv"
)

type ttype byte

const (
	typObject ttype = iota
	typArray
	typBool
	typNumber
	typString
	typNull
)

func (t ttype) String() string {
	switch t {
	case typObject:
		return "object"
	case typArray:
		return "array"
	case typBool:
		return "boolean"
	case typNumber:
		return "number"
	case typString:
		return "string"
	case typNull:
		return "null"
	default:
		return "unknown"
	}
}

type Converter struct {
	decoder *json.Decoder
	types   []ttype
	data    *string
}

func Tokens(j *json.Decoder) *Converter {
	return &Converter{
		decoder: j,
	}
}

func (c *Converter) Token() (xml.Token, error) {
	if len(c.types) > 0 {
		switch c.types[len(c.types)-1] {
		case typObject, typArray:
		default:
			if c.data != nil {
				token := xml.CharData(*c.data)
				c.data = nil
				return token, nil
			}
			return c.outputEnd(), nil
		}
	}
	var keyName *string
	for {
		token, err := c.decoder.Token()
		if err != nil {
			return nil, err
		}
		switch token := token.(type) {
		case json.Delim:
			switch token {
			case '{':
				return c.outputStart(typObject, keyName), nil
			case '[':
				return c.outputStart(typArray, keyName), nil
			case '}', ']':
				return c.outputEnd(), nil
			}
		case bool:
			if token {
				return c.outputType(typBool, &cTrue, keyName), nil
			}
			return c.outputType(typBool, &cFalse, keyName), nil
		case float64:
			number := strconv.FormatFloat(token, 'f', -1, 64)
			return c.outputType(typNumber, &number, keyName), nil
		case json.Number:
			return c.outputType(typNumber, (*string)(&token), keyName), nil
		case string:
			if len(c.types) > 0 && c.types[len(c.types)-1] == typObject && keyName == nil {
				keyName = &token
			} else {
				return c.outputType(typString, &token, keyName), nil
			}
		case nil:
			return c.outputType(typNull, nil, keyName), nil
		}
	}
}

func (c *Converter) outputType(typ ttype, data *string, keyName *string) xml.Token {
	c.data = data
	return c.outputStart(typ, keyName)
}

func (c *Converter) outputStart(typ ttype, keyName *string) xml.Token {
	c.types = append(c.types, typ)
	var attr []xml.Attr
	if keyName != nil {
		attr = []xml.Attr{
			xml.Attr{
				Name: xml.Name{
					Local: "name",
				},
				Value: *keyName,
			},
		}
	}
	return xml.StartElement{
		Name: xml.Name{
			Local: typ.String(),
		},
		Attr: attr,
	}
}

func (c *Converter) outputEnd() xml.Token {
	typ := c.types[len(c.types)-1]
	c.types = c.types[:len(c.types)-1]
	return xml.EndElement{
		Name: xml.Name{
			Local: typ.String(),
		},
	}
}

func Convert(j *json.Decoder, x *xml.Encoder) error {
	c := Converter{
		decoder: j,
	}
	for {
		tk, err := c.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err = x.EncodeToken(tk); err != nil {
			return err
		}
	}
}

var (
	cTrue  = "true"
	cFalse = "false"
)