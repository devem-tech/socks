const domains = [
  "youtube.com",
  "googlevideo.com",
  "googleapis.com",
  "gstatic.com",
  "chatgpt.com",
  "browser-intake-datadoghq.com",
  "proton.me",
  "facebook.com",
  "fbcdn.net",
  "instagram.com",
  "redis.io",
  "x.com",
  "twimg.com",
  // "myexternalip.com",
]

function FindProxyForURL(url, host) {
  if (domains.some(domain => dnsDomainIs(host, domain))) {
    return "SOCKS5 127.0.0.1:7010"
  }

  return "DIRECT"
}
