# blcheck

A powerful script for testing a domain or an IP against mailing block and allow lists.
Script will use dig if it is found. If dig is not found script will use host.


Features
--------------------

* More than __250 block lists__ already included!
* Automatic distinction between __domain or IP__
* Performs __PTR validation__ (only if domain is supplied, does not work for IP)
* 3 verbose (-v) levels and a quiet (-q) mode
* The result of the script is the number of services which blocklisted the domain, so it can be used for any kind of __automated scripts or cronjobs__
* Informative and pleasant output
* Concurrent execution


Requirements
--------------------

* Any Unix/Linux or macOS with BASH.
* Either dig or host command is available.
* GNU parallel command available.


Usage
--------------------

```
blcheck [options] <domain\_or\_IP>

Supplied domain must be a fully qualified domain name.
If the IP is supplied, the PTR check cannot be executed and will be skipped.

<pre>
-d dnshost  Use host as DNS server to make lookups
-l file     Load lists from file with one entry per line
-c          Warn if the top level domain of the list has expired
-v          Verbose mode, can be used multiple times (up to -vvv)
-q          Quiet mode with absolutely no output (useful for scripts)
-p          Plain text output (no coloring, no interactive status)
-j          Number of parallel jobs (e.g. 4) or percentage of cores (e.g. 50%), default: 75%
-h          The help you are just reading
</pre>

Result of the script is the number of blocklisted entries. So if the supplied
IP is not blocklisted on any of the servers the return code is 0.
```


Testing
--------------------

Unit tests live in `tests/blcheck_test.sh` and use [bashunit](https://bashunit.typeddevs.com/).

Install bashunit, then run the tests from the repository root:

```bash
bashunit tests/blcheck_test.sh
```

The tests cover the core helper functions (`info`, `error`, `resolve`, `loadList`) and the
provider list regex parser, without requiring a live DNS server.


TODO
--------------------

1. Handle domains with multiple DNS entries.
2. Handle IPv6.


License
--------------------

blcheck is distributed under the terms of the MIT license. See [license file](LICENSE.md) for details.


Credits
--------------------

Script has been written by the [Intellex](https://intellex.rs/en) team.
Additional contributors:
* [Darko Poljak](https://github.com/darko-poljak)
* [Oliver Mueller](https://github.com/ogmueller)
* [Andrej Walilko](https://github.com/ch604)
