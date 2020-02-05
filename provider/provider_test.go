package provider

import (
	"regexp"
	"testing"
)

func TestMatchBasic(t *testing.T) {
	testDomain := "rbl.rbldns.ru"

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != testDomain {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, testDomain)
	}
	if checkProvider.Active != true {
		t.Errorf("Creating check provider failed for property active: got %v want %v", checkProvider.Active, true)
	}
	if checkProvider.BlType != ProviderTypeDNSBlacklist {
		t.Errorf("Creating check provider failed for property blType: got %v want %v", checkProvider.BlType, ProviderTypeDNSBlacklist)
	}
	if checkProvider.Filter != regexDefaultFilter {
		t.Errorf("Creating check provider failed for property filter: got %v want %v", checkProvider.Filter, regexDefaultFilter)
	}
}

func TestMatchWithFilter(t *testing.T) {
	testDomain := `rbl.rbldns.ru=127\.0\.0\.1`

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != "rbl.rbldns.ru" {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, "rbl.rbldns.ru")
	}
	if checkProvider.Active != true {
		t.Errorf("Creating check provider failed for property active: got %v want %v", checkProvider.Active, true)
	}
	if checkProvider.BlType != ProviderTypeDNSBlacklist {
		t.Errorf("Creating check provider failed for property blType: got %v want %v", checkProvider.BlType, ProviderTypeDNSBlacklist)
	}

	wantedFilter := regexp.MustCompile(`127\.0\.0\.1`)

	if checkProvider.Filter.String() != wantedFilter.String() {
		t.Errorf("Creating check provider failed for property filter: got %v want %v", checkProvider.Filter.String(), wantedFilter.String())
	}
}

func TestMatchWithFilter2(t *testing.T) {
	testDomain := `black.list.com=127.0.[0-9]+.0#URIBL`

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != "black.list.com" {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, "black.list.com")
	}
	if checkProvider.Active != true {
		t.Errorf("Creating check provider failed for property active: got %v want %v", checkProvider.Active, true)
	}
	if checkProvider.BlType != ProviderTypeURIBlacklist {
		t.Errorf("Creating check provider failed for property blType: got %v want %v", checkProvider.BlType, ProviderTypeURIBlacklist)
	}

	wantedFilter := regexp.MustCompile(`127.0.[0-9]+.0`)

	if checkProvider.Filter.String() != wantedFilter.String() {
		t.Errorf("Creating check provider failed for property filter: got %v want %v", checkProvider.Filter.String(), wantedFilter.String())
	}
}

func TestMatchWithType(t *testing.T) {
	testDomain := `rbl.rbldns.ru#DNSWL`

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != "rbl.rbldns.ru" {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, "rbl.rbldns.ru")
	}
	if checkProvider.Active != true {
		t.Errorf("Creating check provider failed for property active: got %v want %v", checkProvider.Active, true)
	}
	if checkProvider.BlType != ProviderTypeDNSWhitelist {
		t.Errorf("Creating check provider failed for property blType: got %v want %v", checkProvider.BlType, ProviderTypeDNSWhitelist)
	}

	if checkProvider.Filter != regexDefaultFilter {
		t.Errorf("Creating check provider failed for property filter: got %v want %v", checkProvider.Filter, regexDefaultFilter)
	}
}

func TestMatchWithType2(t *testing.T) {
	testDomain := `dbl.spamhaus.org#URIBL`

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != "dbl.spamhaus.org" {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, "dbl.spamhaus.org")
	}
	if checkProvider.Active != true {
		t.Errorf("Creating check provider failed for property active: got %v want %v", checkProvider.Active, true)
	}
	if checkProvider.BlType != ProviderTypeURIBlacklist {
		t.Errorf("Creating check provider failed for property blType: got %v want %v", checkProvider.BlType, ProviderTypeURIBlacklist)
	}

	if checkProvider.Filter != regexDefaultFilter {
		t.Errorf("Creating check provider failed for property filter: got %v want %v", checkProvider.Filter, regexDefaultFilter)
	}
}

func TestShouldFailBecauseOfType(t *testing.T) {
	testDomain := `t3direct.dnsbl.net.au#BLA`

	checkProvider := buildCheckProvider(testDomain)

	if checkProvider.URL != "" {
		t.Errorf("Creating check provider failed for property url: got %v want %v", checkProvider.URL, "<nil>")
	}
}
