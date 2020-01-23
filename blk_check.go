package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/swallo/blcheck/iputil"
)

type listedBlacklist struct {
	name string
	txt  string
}

var notListedCount uint64
var listedCount uint64
var listedBlacklists []listedBlacklist

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "--help" {
		fmt.Println("Usage: ./blk_check $CHECKIP $BLACKLIST_FILE")
		os.Exit(0)
	}

	checkIP := args[0]
	blkFilePath := args[1]

	fmt.Println("Checking IP:", checkIP)

	start := time.Now()

	file, err := os.Open(blkFilePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()

	var waitGroup sync.WaitGroup

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		waitGroup.Add(1)

		blChecker := scanner.Text()
		lookupDomain := buildLookupDomain(checkIP, blChecker)

		go dnsLookup(lookupDomain, blChecker, &waitGroup)
	}

	waitGroup.Wait()

	fmt.Println("")
	fmt.Println("")
	fmt.Println("########## RESULT ##########")

	t := time.Now()
	elapsed := t.Sub(start)

	fmt.Println("took", elapsed.Seconds(), "seconds")
	fmt.Println(notListedCount, "times not listed")
	fmt.Println(listedCount, "times not listed")

	fmt.Println("")
	fmt.Println("########## LISTED ##########")
	for _, lb := range listedBlacklists {
		fmt.Println("Checked: ", lb.name)
		fmt.Println("TXT: ", lb.txt)
	}
}

func dnsLookup(lookupDomain string, blChecker string, waitGroup *sync.WaitGroup) {
	ips, err := net.LookupIP(lookupDomain)

	if err != nil {
		atomic.AddUint64(&notListedCount, 1)
	} else {
		atomic.AddUint64(&listedCount, 1)
		fmt.Println(fmt.Sprintf("Listed on %s with status %s", blChecker, ips[0].String()))

		listedBlacklist := listedBlacklist{name: blChecker}

		txt, txtErr := net.LookupTXT(lookupDomain)
		if txtErr != nil {
			listedBlacklist.txt = strings.Join(txt, " ")
		}
		listedBlacklists = append(listedBlacklists, listedBlacklist)
	}

	defer waitGroup.Done()
}

func buildLookupDomain(ip string, host string) string {
	reverseIP := iputil.Reverse(ip)

	return fmt.Sprintf("%s.%s", reverseIP, host)
}
