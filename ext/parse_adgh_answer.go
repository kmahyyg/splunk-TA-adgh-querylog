package ext

import (
	"errors"
	"math"
	"strconv"

	"github.com/miekg/dns"
)

var (
	ErrNoAnswerInResp     = errors.New("no answer in response")
	ErrNoPredefinedParser = errors.New("no predefined parser, this is intended, not an error")
	errNoNeedToCleanup    = errors.New("no need to cleanup original answer")
)

func ParseAnswerInLog(e *ADGHLogEntry) (int, error) {
	if e.Result.IsFiltered {
		e.ParsedAnswer = []string{"filtered"}
		e.AdghResultStr = e.Result.Reason.String()
		return 1, nil
	}
	dnsResp := dns.Msg{}
	err := dnsResp.Unpack(e.Answer)
	if err != nil {
		return -1, err
	}
	if len(dnsResp.Answer) == 0 {
		return 0, ErrNoAnswerInResp
	}
	e.ResponseCode = dnsResp.MsgHdr.Rcode
	e.RequestOpCode = dnsResp.MsgHdr.Opcode
	fDest := make([]string, len(dnsResp.Answer))
	updatedTTL := math.MaxUint32
	for i := 0; i < len(dnsResp.Answer); i++ {
		if ttl := parseTTLFromResponse(dnsResp.Answer[i]); ttl < updatedTTL {
			updatedTTL = ttl
		}
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
	e.NearestTTL = updatedTTL
	e.ParsedAnswer = fDest
	return len(dnsResp.Answer), nil
}

func parseTTLFromResponse(rr dns.RR) int {
	return int(rr.Header().Ttl)
}

func RemoveAnswerInLog(e *ADGHLogEntry) error {
	if e.AdghResultStr != "" {
		e.Result.Reason = 0
	}
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
