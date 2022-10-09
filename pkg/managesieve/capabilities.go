package managesieve

import (
	"sort"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Capabilities struct {
	Implementation string   `managesieve:"IMPLEMENTATION"` // MUST be present
	SASL           []string `managesieve:"SASL"`           //
	Sieve          []string `managesieve:"SIEVE"`          // MUST be present
	StartTLS       bool     `managesieve:"STARTTLS"`       //
	MaxRedirects   int      `managesieve:"MAXREDIRECTS"`   // may be missing; 0 = no limit
	Notify         []string `managesieve:"NOTIFY"`         // may be empty if the server does not support enotify
	Language       string   `managesieve:"LANGUAGE"`       // RFC4656; empty = i-default (RFC2277)
	Owner          string   `managesieve:"OWNER"`          // may be present only when authenticated
	Version        string   `managesieve:"VERSION"`        // MUST be present
}

func ParseCapabilities(content []string) (capabilities *Capabilities, err error) {
	values := make(map[string]string, len(content))

	for _, line := range content {
		parts := strings.SplitN(line, " ", 2)

		value := ""
		if len(parts) > 1 {
			value, err = strconv.Unquote(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, err
			}
		}

		key, err := strconv.Unquote(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}

		values[key] = value
	}

	capabilities = &Capabilities{}
	err = mapstructure.WeakDecode(values, &capabilities)
	if err != nil {
		return nil, err
	}

	if len(capabilities.Sieve) > 0 {
		capabilities.Sieve = strings.Split(capabilities.Sieve[0], " ")
	}

	if len(capabilities.SASL) > 0 {
		capabilities.SASL = strings.Split(capabilities.SASL[0], " ")
	}

	sort.Strings(content)
	searchStr := strconv.Quote(StartTLS)
	searchIndex := sort.SearchStrings(content, searchStr)
	capabilities.StartTLS = searchIndex != len(content) && content[searchIndex] == searchStr

	return capabilities, err
}
