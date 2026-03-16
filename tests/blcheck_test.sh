#!/bin/bash
#
# Unit tests for blcheck functions.
# Requires bashunit: https://bashunit.typeddevs.com/

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/blcheck"

function set_up() {
    # Globals required by the sourced functions
    VERBOSE=0
    PLAIN=
    SPINNER="-\|/"
    RED="" GREEN="" YELLOW="" CLEAR=""
    CONF_DNS_TRIES=2
    CONF_DNS_DURATION=3
    COUNT=10
    COUNT_FILE="$(mktemp)"
    # Source regex definitions directly from the script so tests stay in sync
    local _regexfile
    _regexfile="$(mktemp)"
    grep -E '^\s+REGEX_[A-Z0-9_]+=' "${SCRIPT}" > "${_regexfile}"
    # shellcheck source=/dev/null
    source "${_regexfile}"
    rm -f "${_regexfile}"
    CMD_DIG="$(command -v dig || true)"
    CMD_HOST="$(command -v host || true)"
    CMD="${CMD_DIG:-${CMD_HOST}}"
    DNSSERVER=""

    # Source only the Macros section (function definitions)
    local _tmpfile
    _tmpfile="$(mktemp)"
    awk '/^# Macros \{/{p=1} p{print} /^#~ ~ ~.*\}$/{if(p){exit}}' "${SCRIPT}" > "${_tmpfile}"
    # shellcheck source=/dev/null
    source "${_tmpfile}"
    rm -f "${_tmpfile}"
}

function tear_down() {
    rm -f "${COUNT_FILE}"
}

# ---------------------------------------------------------------------------
# info()
# ---------------------------------------------------------------------------

function test_info_prints_when_verbose_equals_required_level() {
    VERBOSE=1
    assert_equals "hello" "$(info 1 "hello")"
}

function test_info_prints_when_verbose_exceeds_required_level() {
    VERBOSE=3
    assert_equals "hello" "$(info 1 "hello")"
}

function test_info_suppresses_message_when_verbose_is_below_required_level() {
    VERBOSE=0
    assert_empty "$(info 1 "hello")"
}

# ---------------------------------------------------------------------------
# error()
# ---------------------------------------------------------------------------

function test_error_exits_with_status_2() {
    ( error "something went wrong" 2>/dev/null )
    assert_exit_code 2
}

function test_error_outputs_error_message() {
    local output
    output=$(error "something went wrong" 2>&1) || true
    assert_contains "ERROR: something went wrong" "$output"
}

# ---------------------------------------------------------------------------
# resolve() — IP passthrough (no DNS required)
# ---------------------------------------------------------------------------

function test_resolve_returns_plain_ip_unchanged() {
    assert_equals "1.2.3.4" "$(resolve "1.2.3.4")"
}

function test_resolve_returns_loopback_ip_unchanged() {
    assert_equals "127.0.0.1" "$(resolve "127.0.0.1")"
}

# ---------------------------------------------------------------------------
# IP reversal (sed expression used inline in the main script)
# ---------------------------------------------------------------------------

function test_ip_octets_are_reversed_for_dnsbl_lookup() {
    local result
    result=$(echo "1.2.3.4" | sed -ne "s~^${REGEX_IP}$~\4.\3.\2.\1~p")
    assert_equals "4.3.2.1" "$result"
}

function test_ip_reversal_handles_multi_digit_octets() {
    local result
    result=$(echo "192.168.10.1" | sed -ne "s~^${REGEX_IP}$~\4.\3.\2.\1~p")
    assert_equals "1.10.168.192" "$result"
}

# ---------------------------------------------------------------------------
# loadList()
# ---------------------------------------------------------------------------

function test_load_list_loads_single_entry() {
    local list_file
    list_file="$(mktemp)"
    echo "black.list.com" > "${list_file}"
    loadList "${list_file}"
    rm -f "${list_file}"
    assert_equals "black.list.com" "${CONF_LISTS}"
}

function test_load_list_ignores_comment_lines() {
    local list_file
    list_file="$(mktemp)"
    printf "# comment\nblack.list.com\n" > "${list_file}"
    loadList "${list_file}"
    rm -f "${list_file}"
    assert_equals "black.list.com" "${CONF_LISTS}"
}

function test_load_list_ignores_empty_lines() {
    local list_file
    list_file="$(mktemp)"
    printf "\nblack.list.com\n\n" > "${list_file}"
    loadList "${list_file}"
    rm -f "${list_file}"
    assert_equals "black.list.com" "${CONF_LISTS}"
}

function test_load_list_loads_multiple_entries() {
    local list_file
    list_file="$(mktemp)"
    printf "list.one.com\nlist.two.com\n" > "${list_file}"
    loadList "${list_file}"
    rm -f "${list_file}"
    assert_contains "list.one.com" "${CONF_LISTS}"
    assert_contains "list.two.com" "${CONF_LISTS}"
}

function test_load_list_fails_with_empty_argument() {
    ( loadList "" 2>/dev/null )
    assert_exit_code 2
}

function test_load_list_fails_with_nonexistent_file() {
    ( loadList "/nonexistent/file/path" 2>/dev/null )
    assert_exit_code 2
}

# ---------------------------------------------------------------------------
# REGEX_LIST — provider list entry format
# ---------------------------------------------------------------------------

function test_regex_list_matches_plain_domain() {
    [[ "black.list.com" =~ ${REGEX_LIST} ]]
    assert_equals "black.list.com" "${BASH_REMATCH[1]}"
}

function test_regex_list_captures_filter_when_present() {
    [[ 'black.list.com=127\.0\.0\.2' =~ ${REGEX_LIST} ]]
    assert_equals "black.list.com" "${BASH_REMATCH[1]}"
    assert_equals '127\.0\.0\.2' "${BASH_REMATCH[3]}"
}

function test_regex_list_matches_dnsbl_type() {
    [[ "black.list.com#DNSBL" =~ ${REGEX_LIST} ]]
    assert_equals "DNSBL" "${BASH_REMATCH[6]}"
}

function test_regex_list_matches_dnswl_type() {
    [[ "white.list.com#DNSWL" =~ ${REGEX_LIST} ]]
    assert_equals "DNSWL" "${BASH_REMATCH[6]}"
}

function test_regex_list_matches_uribl_type() {
    [[ "uri.list.com#URIBL" =~ ${REGEX_LIST} ]]
    assert_equals "URIBL" "${BASH_REMATCH[6]}"
}

function test_regex_list_matches_filter_and_type_combined() {
    [[ 'black.list.com=127\.0\.0\.2#DNSBL' =~ ${REGEX_LIST} ]]
    assert_equals "black.list.com" "${BASH_REMATCH[1]}"
    assert_equals '127\.0\.0\.2' "${BASH_REMATCH[3]}"
    assert_equals "DNSBL" "${BASH_REMATCH[6]}"
}

function test_regex_list_does_not_match_unknown_type() {
    [[ ! "black.list.com#INVALID" =~ ${REGEX_LIST} ]]
    assert_successful_code
}

# ---------------------------------------------------------------------------
# REGEX_IP6 — IPv6 address detection
# ---------------------------------------------------------------------------

function test_regex_ip6_matches_full_address() {
    [[ "2001:0db8:0000:0000:0000:0000:0000:0001" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

function test_regex_ip6_matches_compressed_address() {
    [[ "2001:db8::1" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

function test_regex_ip6_matches_loopback() {
    [[ "::1" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

function test_regex_ip6_matches_all_zeros() {
    [[ "::" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

function test_regex_ip6_does_not_match_ipv4() {
    [[ ! "1.2.3.4" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

function test_regex_ip6_does_not_match_plain_string() {
    [[ ! "not-an-ip" =~ ${REGEX_IP6} ]]
    assert_successful_code
}

# ---------------------------------------------------------------------------
# expand_ipv6()
# ---------------------------------------------------------------------------

function test_expand_ipv6_full_address() {
    assert_equals "20010db8000000000000000000000001" \
        "$(expand_ipv6 "2001:0db8:0000:0000:0000:0000:0000:0001")"
}

function test_expand_ipv6_compressed_middle() {
    assert_equals "20010db8000000000000000000000001" \
        "$(expand_ipv6 "2001:db8::1")"
}

function test_expand_ipv6_loopback() {
    assert_equals "00000000000000000000000000000001" \
        "$(expand_ipv6 "::1")"
}

function test_expand_ipv6_all_zeros() {
    assert_equals "00000000000000000000000000000000" \
        "$(expand_ipv6 "::")"
}

function test_expand_ipv6_trailing_compression() {
    assert_equals "20010db8000000000000000000000000" \
        "$(expand_ipv6 "2001:db8::")"
}

function test_expand_ipv6_uppercases_are_lowercased() {
    assert_equals "20010db8000000000000000000000001" \
        "$(expand_ipv6 "2001:DB8::1")"
}

# ---------------------------------------------------------------------------
# resolve() — IPv6 passthrough (no DNS required)
# ---------------------------------------------------------------------------

function test_resolve_returns_compressed_ipv6_unchanged() {
    assert_equals "2001:db8::1" "$(resolve "2001:db8::1")"
}

function test_resolve_returns_loopback_ipv6_unchanged() {
    assert_equals "::1" "$(resolve "::1")"
}

function test_resolve_returns_full_ipv6_unchanged() {
    assert_equals "2001:0db8:0000:0000:0000:0000:0000:0001" \
        "$(resolve "2001:0db8:0000:0000:0000:0000:0000:0001")"
}

# ---------------------------------------------------------------------------
# IPv6 reversal — nibble-reversed dot-separated format for DNSBL lookups
# ---------------------------------------------------------------------------

function test_ipv6_reversal_compressed_address() {
    local result
    result=$(expand_ipv6 "2001:db8::1" | rev | sed 's/./&./g; s/\.$//')
    assert_equals "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2" "$result"
}

function test_ipv6_reversal_loopback() {
    local result
    result=$(expand_ipv6 "::1" | rev | sed 's/./&./g; s/\.$//')
    assert_equals "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0" "$result"
}

function test_ipv6_reversal_produces_31_dots() {
    local result
    result=$(expand_ipv6 "2001:db8::1" | rev | sed 's/./&./g; s/\.$//')
    local dot_count
    dot_count=$(tr -cd '.' <<< "$result" | wc -c | tr -d ' ')
    assert_equals "31" "$dot_count"
}
