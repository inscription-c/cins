package util

import "strings"

type AcceptEncoding struct {
	accept string
}

func ParseAcceptEncoding(acceptEncoding string) *AcceptEncoding {
	return &AcceptEncoding{accept: acceptEncoding}
}

func (a *AcceptEncoding) IsAccept(encoding string) bool {
	if encoding == "" {
		return false
	}
	for _, v := range strings.Split(a.accept, ",") {
		for _, v := range strings.Split(v, ";") {
			if strings.TrimSpace(v) == encoding {
				return true
			}
		}
	}
	return false
}
