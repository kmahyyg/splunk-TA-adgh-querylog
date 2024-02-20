package ext

import (
	"errors"
	"github.com/miekg/dns"
	"strconv"
)

var (
	ErrNoAnswerInResp     = errors.New("no answer in response")
	ErrNoPredefinedParser = errors.New("no predefined parser, this is intended, not an error")
	errNoNeedToCleanup    = errors.New("no need to cleanup original answer")
)

func ParseAnswerInLog(e *ADGHLogEntry) (int, error) {
	if e.Result.IsFiltered {
		e.ParsedAnswer = []string{"filtered"}
		return 1, nil
	}
	dnsResp := dns.Msg{}
	err := dnsResp.Unpack(e.Answer)
	if err != nil {
		return -1, err
	}
	if len(dnsResp.Answer) == 0 {
		return 0, nil
	}
	fDest := make([]string, len(dnsResp.Answer))
	for i := 0; i < len(dnsResp.Answer); i++ {
		switch t := dnsResp.Answer[i].(type) {
		case *dns.A:
			fDest[i] = t.A.String()
		case *dns.AAAA:
			fDest[i] = t.AAAA.String()
		case *dns.CNAME:
			fDest[i] = t.Target
		case *dns.DNAME:
			fDest[i] = t.Target
		case *dns.SRV:
			fDest[i] = t.Target + ":" + strconv.Itoa(int(t.Port))
		case *dns.MX:
			fDest[i] = t.Mx
		case *dns.NS:
			fDest[i] = t.Ns
		default:
			return 0, ErrNoPredefinedParser
		}
	}
	e.ParsedAnswer = fDest
	return len(dnsResp.Answer), nil
}

func RemoveAnswerInLog(e *ADGHLogEntry) error {
	switch e.QType {
	case "A":
		fallthrough
	case "CNAME":
		fallthrough
	case "AAAA":
		fallthrough
	case "TXT":
		fallthrough
	case "DNAME":
		fallthrough
	case "SRV":
		fallthrough
	case "MX":
		fallthrough
	case "NS":
		e.Answer = nil
	default:
		return errNoNeedToCleanup
	}

	return nil
}
