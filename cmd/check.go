package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/swallo/blcheck/iputil"
)

type blProvider struct {
	url    string
	active bool
}

type listedBlacklist struct {
	name   string
	txt    string
	status string
}

var (
	// arguments
	checkIP string
	// flags
	dnsTimeout int
	dnsRetries int
	blFilePath string

	// list of all providers
	blProviderList []blProvider

	// properties
	notListedCount   uint64
	listedCount      uint64
	timeOutCount     uint64
	listedBlacklists []listedBlacklist
)

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&blFilePath, "blacklist", "./bl_list", "custom blacklist file path")
	checkCmd.Flags().IntVar(&dnsTimeout, "timeout", 3, "defines the timeout for the dns request")
	checkCmd.Flags().IntVar(&dnsRetries, "retries", 2, "defines the amount of retries if request was unsuccessful")
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

		checkIP = args[0]
		if iputil.Validate(checkIP) == false {
			return errors.New(color.Error.Sprintf("This is not a valid ipv4, provided: %s", checkIP))
		}

		return nil
	},
	Long: `Run blchecker against a list of blacklist providers via the dns lookup tool`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking IP:", color.Bold.Render(checkIP))

		start := time.Now()

		file, err := os.Open(blFilePath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer file.Close()

		buildProviderList(file)
		fmt.Println("Checking", color.Bold.Render(len(blProviderList)), "providers")
		fmt.Println("")

		bar := buildProgressBar(len(blProviderList))

		var waitGroup sync.WaitGroup

		for _, blChecker := range blProviderList {
			if blChecker.active {
				waitGroup.Add(1)
				lookupDomain := buildLookupDomain(checkIP, blChecker.url)
				go dnsLookup(lookupDomain, blChecker.url, &waitGroup, bar)
			}
		}
		waitGroup.Wait()

		printResults(time.Now().Sub(start))
	},
}

func buildProgressBar(total int) *pb.ProgressBar {
	bar := pb.New(total)

	return bar
}

func dnsLookup(lookupDomain string, blChecker string, waitGroup *sync.WaitGroup, bar *pb.ProgressBar) {
	resultChan := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()
	defer waitGroup.Done()

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, lookupDomain)
	resultChan <- true

	select {
	case <-resultChan:
		if err != nil {
			atomic.AddUint64(&notListedCount, 1)
		} else {
			atomic.AddUint64(&listedCount, 1)
			listedBlacklist := listedBlacklist{name: blChecker, status: ips[0].String()}

			txt, txtErr := net.DefaultResolver.LookupTXT(ctx, lookupDomain)
			if txtErr == nil {
				listedBlacklist.txt = strings.Join(txt, " ")
			}
			listedBlacklists = append(listedBlacklists, listedBlacklist)
		}
	case <-ctx.Done():
		atomic.AddUint64(&timeOutCount, 1)
	}

	bar.Add(1)
}

func printResults(elapsed time.Duration) {
	fmt.Println("")
	fmt.Println("")

	fmt.Println("took", elapsed.Seconds(), "seconds", "for", len(blProviderList), "providers")
	fmt.Println(color.Success.Render(notListedCount), "times not listed")
	fmt.Println(color.Warn.Render(timeOutCount), "timeouts (with", dnsTimeout, "seconds value)")
	fmt.Println(color.Error.Render(listedCount), "times listed")

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

func buildLookupDomain(ip string, host string) string {
	reverseIP := iputil.Reverse(ip)

	return fmt.Sprintf("%s.%s", reverseIP, host)
}

func buildProviderList(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		blProviderList = append(blProviderList, blProvider{url: scanner.Text(), active: true})
	}
}
