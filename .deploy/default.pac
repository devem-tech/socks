const domains = [
  "myexternalip.com",
]

function FindProxyForURL(url, host) {
  if (domains.some(domain => dnsDomainIs(host, domain))) {
    return "SOCKS5 127.0.0.1:7010"
  }

  return "DIRECT"
}
