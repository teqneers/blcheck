package provider

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// ProviderTypeDNSBlacklist is the const for dns bl
const ProviderTypeDNSBlacklist = "DNSBL"

// ProviderTypeDNSWhitelist is the const for dns wl
const ProviderTypeDNSWhitelist = "DNSWL"

// ProviderTypeURIBlacklist is the const for url bl
const ProviderTypeURIBlacklist = "URIBL"

// CheckProvider is the list element for the check list
type CheckProvider struct {
	// URL of the bl | wl provider
	URL    string
	BlType string
	Filter *regexp.Regexp
	Active bool
}

var (
	regexProviderData  = regexp.MustCompile(`^[ \t]*(?P<url>[^=#]+)[=]?(?P<filter>[^#]+)?[#]?(?P<type>DNSBL|DNSWL|URIBL)?[ \t]*$`)
	regexDefaultFilter = regexp.MustCompile(`127\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})`)
)

// IsNilProvider checks if given provider is valid
func IsNilProvider(checkProvider CheckProvider) bool {
	return checkProvider.URL == ""
}

// BuildProviderList reads the provided file line by line and generates the dns / uri bl|wl provider
func BuildProviderList(r io.Reader) []CheckProvider {
	var checkProviderList []CheckProvider

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		readLine := scanner.Text()

		if strings.HasPrefix(readLine, "#") {
			continue
		}
		checkProviderList = append(checkProviderList, buildCheckProvider(readLine))
	}

	return checkProviderList
}

func buildCheckProvider(readLine string) CheckProvider {
	var url string
	var blType string
	var filter *regexp.Regexp

	matches := regexProviderData.FindStringSubmatch(readLine)
	regexpNames := regexProviderData.SubexpNames()

	for i, match := range matches {
		if i == 0 || regexpNames[i] == "" {
			continue
		}

		switch regexpNames[i] {
		case "url":
			url = match
			break
		case "type":
			blType = match
			if match == "" {
				blType = ProviderTypeDNSBlacklist
			}
			break
		case "filter":
			filter = regexp.MustCompile(match)
			if match == "" {
				filter = regexDefaultFilter
			}
			break
		default:
			continue
		}
	}

	return CheckProvider{URL: url, BlType: blType, Filter: filter, Active: true}
}
