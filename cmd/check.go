package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/teqneers/blcheck/dnsutil"
	"github.com/teqneers/blcheck/iputil"
	l "github.com/teqneers/blcheck/logutil"
	"github.com/teqneers/blcheck/progressbar"
	"github.com/teqneers/blcheck/provider"
)

type checkElement struct {
	IP   string
	host string
}

type listedBlacklist struct {
	name   string
	txt    string
	status string
}

var (
	// arguments
	toCheckHostOrIP string
	// flags
	dnsTimeout  int
	dnsRetries  int
	dnsThrottle int
	filePath    string

	// properties
	disabledCount    uint64
	skippedCount     uint64
	notListedCount   uint64
	listedCount      uint64
	timeOutCount     uint64
	listedBlacklists []listedBlacklist
)

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&filePath, "checklist", "./bl_list", "custom checklist file path")
	checkCmd.Flags().IntVar(&dnsTimeout, "timeout", 20, "defines the timeout for the dns request")
	checkCmd.Flags().IntVar(&dnsRetries, "retries", 2, "defines the amount of retries if request was unsuccessful")
	checkCmd.Flags().IntVar(&dnsThrottle, "throttle", 25, "defines the amount of dns requests per second")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run blchecker against a list of blacklist providers",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New(color.Error.Render("requires an ip as argument"))
		}
		if len(args) > 1 {
			return errors.New(color.Error.Render("Too many arguments provided"))
		}

		toCheckHostOrIP = args[0]
		if iputil.ValidateIP(toCheckHostOrIP) == false && iputil.IsIP(toCheckHostOrIP) == false && iputil.IsHostname(toCheckHostOrIP) == false {
			return errors.New(color.Error.Sprintf("This is not a valid ipv4 or hostname, provided: %s", toCheckHostOrIP))
		}

		return nil
	},
	Long: `Run blchecker against a list of blacklist providers via the dns lookup tool`,
	Run:  runCommand,
}

func runCommand(cmd *cobra.Command, args []string) {
	start := time.Now()

	var checkElement checkElement
	var checkIP string

	isQuiet, _ := cmd.Flags().GetBool("quiet")
	verboseLevel, _ := cmd.Flags().GetCount("verbose")
	l.CurrentLogLevel = verboseLevel
	l.IsQuiet = isQuiet

	if iputil.IsIP(toCheckHostOrIP) {
		checkElement.IP = checkIP
		checkElement.host = ""

		l.LogInfo(0, fmt.Sprintf("Checking: %s", toCheckHostOrIP))
	} else {
		checkElement.IP = dnsutil.LookupIPFromHostname(toCheckHostOrIP, dnsTimeout)
		checkElement.host = toCheckHostOrIP

		ptrRecord := dnsutil.LookupPtrRecordForIP(checkElement.IP, dnsTimeout)

		l.LogInfo(1, fmt.Sprintf("Found PTR Record for IP %s: %s", checkElement.IP, ptrRecord))

		if ptrRecord != fmt.Sprintf("%s.", checkElement.host) {
			l.LogError(0, fmt.Sprintf("PTR Record does not match host: %s != %s", ptrRecord, checkElement.host))
		}
		l.LogInfo(0, fmt.Sprintf("PTR Records do match (%s)", ptrRecord))

		l.LogInfo(0, fmt.Sprintf("Checking IP %s (from Host: %s)", checkElement.IP, checkElement.host))
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(144)
	}

	checkProviderList := provider.BuildProviderList(file)
	file.Close()

	l.LogInfo(0, fmt.Sprint("Checking ", color.Bold.Render(len(checkProviderList)), " providers"))

	progressbar.CreateProgressbar(len(checkProviderList), (isQuiet || verboseLevel > 0))

	var waitGroup sync.WaitGroup

	rate := time.Second / time.Duration(dnsThrottle)
	throttle := time.Tick(rate)

	for _, blChecker := range checkProviderList {
		<-throttle
		waitGroup.Add(1)
		go dnsLookup(checkElement, blChecker, &waitGroup)
	}
	waitGroup.Wait()

	if isQuiet == false {
		printResults(time.Now().Sub(start), checkProviderList)
	}

	os.Exit(int(listedCount))
}

func dnsLookup(checkElement checkElement, blChecker provider.CheckProvider, waitGroup *sync.WaitGroup) {
	var lookupDomain string

	resultChan := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()
	defer waitGroup.Done()
	defer progressbar.AddToBar(1)

	if !blChecker.Active {
		atomic.AddUint64(&disabledCount, 1)
		l.LogWarning(1, fmt.Sprintf("Provider %s is deactivated", blChecker.URL))
	} else {
		if blChecker.BlType == provider.ProviderTypeURIBlacklist {
			if checkElement.host != "" {
				lookupDomain = fmt.Sprintf("%s.%s", checkElement.host, blChecker.URL)
			} else {
				atomic.AddUint64(&skippedCount, 1)
				l.LogInfo(1, fmt.Sprintf("Notice: URIBL entry '%s' will be ignore, because %s is an IP address.", blChecker.URL, checkElement.IP))
				return
			}
		} else {
			lookupDomain = fmt.Sprintf("%s.%s", iputil.Reverse(checkElement.IP), blChecker.URL)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, lookupDomain)
		resultChan <- true

		select {
		case <-resultChan:
			if err != nil {
				atomic.AddUint64(&notListedCount, 1)
				l.LogSuccess(1, fmt.Sprintf("Checked %s ✓", blChecker.URL))
			} else {
				resultIP := ips[0].String()

				if blChecker.Filter.Match([]byte(resultIP)) {
					atomic.AddUint64(&notListedCount, 1)
					l.LogSuccess(1, fmt.Sprintf("Checked %s (matched provided filter %s) ✓", blChecker.URL, blChecker.Filter.String()))
				} else {
					atomic.AddUint64(&listedCount, 1)
					l.LogError(1, fmt.Sprintf("Checked %s ✓", blChecker.URL))

					listedBlacklist := listedBlacklist{name: blChecker.URL, status: resultIP}

					txt, txtErr := net.DefaultResolver.LookupTXT(ctx, lookupDomain)
					if txtErr == nil {
						listedBlacklist.txt = strings.Join(txt, " ")
					}
					listedBlacklists = append(listedBlacklists, listedBlacklist)
					l.LogError(1, fmt.Sprintf("Checked %s - TXT: %s", blChecker.URL, listedBlacklist.txt))
				}
			}
		case <-ctx.Done():
			atomic.AddUint64(&timeOutCount, 1)
			l.LogInfo(1, fmt.Sprintf("Checked %s, but timed out !", blChecker.URL))
		}
	}
}

func printResults(elapsed time.Duration, checkProviderList []provider.CheckProvider) {
	fmt.Println("")
	fmt.Println("")

	fmt.Println("took", fmt.Sprintf("%6.2f", elapsed.Seconds()), "seconds", "for", len(checkProviderList), "providers")
	fmt.Println(color.Error.Render(listedCount), "times listed.")
	fmt.Println(color.Success.Render(notListedCount), "times not listed.")

	if timeOutCount > 0 {
		fmt.Println(color.Warn.Render(timeOutCount), "timeouts (with", dnsTimeout, "seconds value).")
	}
	if skippedCount > 0 {
		fmt.Println(color.Warn.Render(skippedCount), "skipped.")
	}
	if disabledCount > 0 {
		fmt.Println(color.Warn.Render(disabledCount), "disabled.")
	}

	if len(listedBlacklists) > 0 {
		fmt.Println("")
		fmt.Println("")
		fmt.Println(color.Warn.Render("########## LISTED ##########"))

		t := tabby.New()
		t.AddHeader("NAME", "STATUS", "TXT")

		for _, lb := range listedBlacklists {
			t.AddLine(lb.name, lb.status, lb.txt)
		}
		t.Print()
	}
}
