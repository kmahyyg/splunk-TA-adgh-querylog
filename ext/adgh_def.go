package ext

import (
	"net"
	"time"
)

// ADGHFilteringResult is simplified from filtering.Results due to efficiency consideration, this is not cybersecurity related stuff.
// check https://github.com/AdguardTeam/AdGuardHome/blob/bd99e3e09d00b7e6a66eee568c0133b6fce0fdc6/internal/filtering/filtering.go#L547
type ADGHFilteringResult struct {
	// REMOVED because unrelated to Security
	// DNSRewriteResult is the $dnsrewrite filter rule result.
	// DNSRewriteResult *DNSRewriteResult `json:",omitempty"`

	// REMOVED because unrelated to Security
	// CanonName is the CNAME value from the lookup rewrite result.  It is empty
	// unless ADGHFilteringReason is set to Rewritten or RewrittenRule.
	// CanonName string `json:",omitempty"`

	// REMOVED because unrelated to Security
	// ServiceName is the name of the blocked service.  It is empty unless
	// ADGHFilteringReason is set to FilteredBlockedService.
	// ServiceName string `json:",omitempty"`

	// REMOVED because unrelated to Security
	// IPList is the lookup rewrite result.  It is empty unless ADGHFilteringReason is set to
	// Rewritten.
	// IPList []netip.Addr `json:",omitempty"`

	// REMOVED because unrelated to Security
	// Rules are applied rules.  If Rules are not empty, each rule is not nil.
	// Rules []*ResultRule `json:",omitempty"`

	// Reason is the reason for blocking or unblocking the request.
	Reason ADGHFilteringReason `json:",omitempty"`

	// IsFiltered is true if the request is filtered.
	IsFiltered bool `json:",omitempty"`
}

// ADGHLogEntry is extracted from ADGH code
// check https://github.com/AdguardTeam/AdGuardHome/blob/bd99e3e09d00b7e6a66eee568c0133b6fce0fdc6/internal/querylog/entry.go#L20
// for more details.
//
// Client information about whois/DHCP related definition is removed due to efficiency consideration
// check https://github.com/AdguardTeam/AdGuardHome/blob/master/internal/querylog/client.go#L7 for more.
type ADGHLogEntry struct {
	Time time.Time `json:"T"`

	QHost  string `json:"QH"`
	QType  string `json:"QT"`
	QClass string `json:"QC"`

	ReqECS string `json:"ECS,omitempty"`

	ClientID    string `json:"CID,omitempty"`
	ClientProto string `json:"CP"`

	Upstream string `json:",omitempty"`

	Answer     []byte `json:",omitempty"`
	OrigAnswer []byte `json:",omitempty"`

	// Customized Field
	ParsedAnswer []string `json:",omitempty"`

	IP net.IP `json:"IP"`

	Result ADGHFilteringResult
	// Customized Field
	AdghResultStr string `json:",omitempty"`
	// Customized Field
	NearestTTL int `json:",omitempty"`
	// Customized Field
	ResponseCode int `json:",omitempty"`

	Elapsed time.Duration

	Cached            bool `json:",omitempty"`
	AuthenticatedData bool `json:"AD,omitempty"`
}

// ADGHFilteringReason is imported from https://github.com/AdguardTeam/AdGuardHome/blob/bd99e3e09d00b7e6a66eee568c0133b6fce0fdc6/internal/filtering/filtering.go#L332C21-L332C27
type ADGHFilteringReason int

const (
	// reasons for not filtering

	// NotFilteredNotFound - host was not find in any checks, default value for result
	NotFilteredNotFound ADGHFilteringReason = iota
	// NotFilteredAllowList - the host is explicitly allowed
	NotFilteredAllowList
	// NotFilteredError is returned when there was an error during
	// checking.  Reserved, currently unused.
	NotFilteredError

	// reasons for filtering

	// FilteredBlockList - the host was matched to be advertising host
	FilteredBlockList
	// FilteredSafeBrowsing - the host was matched to be malicious/phishing
	FilteredSafeBrowsing
	// FilteredParental - the host was matched to be outside of parental control settings
	FilteredParental
	// FilteredInvalid - the request was invalid and was not processed
	FilteredInvalid
	// FilteredSafeSearch - the host was replaced with safesearch variant
	FilteredSafeSearch
	// FilteredBlockedService - the host is blocked by "blocked services" settings
	FilteredBlockedService

	// Rewritten is returned when there was a rewrite by a legacy DNS rewrite
	// rule.
	Rewritten

	// RewrittenAutoHosts is returned when there was a rewrite by autohosts
	// rules (/etc/hosts and so on).
	RewrittenAutoHosts

	// RewrittenRule is returned when a $dnsrewrite filter rule was applied.
	// See https://github.com/AdguardTeam/AdGuardHome/issues/2499.
	RewrittenRule
)

var reasonNames = []string{
	NotFilteredNotFound:  "NotFilteredNotFound",
	NotFilteredAllowList: "NotFilteredWhiteList",
	NotFilteredError:     "NotFilteredError",

	FilteredBlockList:      "FilteredBlackList",
	FilteredSafeBrowsing:   "FilteredSafeBrowsing",
	FilteredParental:       "FilteredParental",
	FilteredInvalid:        "FilteredInvalid",
	FilteredSafeSearch:     "FilteredSafeSearch",
	FilteredBlockedService: "FilteredBlockedService",

	Rewritten:          "Rewrite",
	RewrittenAutoHosts: "RewriteEtcHosts",
	RewrittenRule:      "RewriteRule",
}

func (r ADGHFilteringReason) String() string {
	if r < 0 || int(r) >= len(reasonNames) {
		return ""
	}

	return reasonNames[r]
}
