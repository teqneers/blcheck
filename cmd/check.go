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
	pb "github.com/schollz/progressbar/v2"
	"github.com/spf13/cobra"
	"github.com/teqneers/blcheck/iputil"
	l "github.com/teqneers/blcheck/logutil"
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
	checkCmd.Flags().IntVar(&dnsTimeout, "timeout", 3, "defines the timeout for the dns request")
	checkCmd.Flags().IntVar(&dnsRetries, "retries", 2, "defines the amount of retries if request was unsuccessful")
	checkCmd.Flags().IntVar(&dnsThrottle, "throttle", 20, "defines the amount of dns requests per second")
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
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		var checkElement checkElement
		var checkIP string

		isQuiet, _ := cmd.Flags().GetBool("quiet")
		l.CurrentLogLevel, _ = cmd.Flags().GetCount("verbose")
		l.IsQuiet = isQuiet

		if iputil.IsIP(toCheckHostOrIP) {
			checkElement.IP = checkIP
			checkElement.host = ""

			l.LogInfo(0, fmt.Sprintf("Checking: %s", toCheckHostOrIP))
		} else {
			checkElement.IP = getIPFromHostname(toCheckHostOrIP)
			checkElement.host = toCheckHostOrIP

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

		bar := buildProgressBar(len(checkProviderList))

		var waitGroup sync.WaitGroup

		rate := time.Second / time.Duration(dnsThrottle)
		throttle := time.Tick(rate)

		for _, blChecker := range checkProviderList {
			<-throttle
			waitGroup.Add(1)
			go dnsLookup(checkElement, blChecker, &waitGroup, bar)
		}
		waitGroup.Wait()

		if isQuiet == false {
			printResults(time.Now().Sub(start), checkProviderList)
		}
		if listedCount > 0 {
			os.Exit(1)
		}
	},
}

func buildProgressBar(total int) *pb.ProgressBar {
	bar := pb.New(total)

	return bar
}

func dnsLookup(checkElement checkElement, blChecker provider.CheckProvider, waitGroup *sync.WaitGroup, bar *pb.ProgressBar) {
	var lookupDomain string

	resultChan := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()
	defer waitGroup.Done()
	defer bar.Add(1)

	if !blChecker.Active {
		atomic.AddUint64(&disabledCount, 1)
	} else {
		if blChecker.BlType == provider.ProviderTypeURIBlacklist {
			if checkElement.host != "" {
				lookupDomain = fmt.Sprintf("%s.%s", checkElement.host, blChecker.URL)
			} else {
				atomic.AddUint64(&skippedCount, 1)
				l.LogInfo(1, fmt.Sprintf("The host %s is a URI based blacklist, but no host provided.", blChecker.URL))
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
			} else {
				resultIP := ips[0].String()

				if blChecker.Filter.Match([]byte(resultIP)) {
					atomic.AddUint64(&notListedCount, 1)
				} else {
					atomic.AddUint64(&listedCount, 1)
					listedBlacklist := listedBlacklist{name: blChecker.URL, status: resultIP}

					txt, txtErr := net.DefaultResolver.LookupTXT(ctx, lookupDomain)
					if txtErr == nil {
						listedBlacklist.txt = strings.Join(txt, " ")
					}
					listedBlacklists = append(listedBlacklists, listedBlacklist)
				}
			}
		case <-ctx.Done():
			atomic.AddUint64(&timeOutCount, 1)
		}
	}
}

func getIPFromHostname(lookupDomain string) string {
	var resultIP string

	resultChan := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()

	ips, _ := net.DefaultResolver.LookupIPAddr(ctx, lookupDomain)
	resultChan <- true

	select {
	case <-resultChan:
		if len(ips) == 1 {
			resultIP = ips[0].IP.String()
		} else if len(ips) > 1 {
			l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s returned multiple IPs.", lookupDomain))
			os.Exit(155)
		} else {
			l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s failed.", lookupDomain))
			os.Exit(155)
		}

	case <-ctx.Done():
		l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s timed out after %d seconds.", lookupDomain, time.Duration(dnsTimeout+1)*time.Second))
		os.Exit(155)
	}

	return resultIP
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
