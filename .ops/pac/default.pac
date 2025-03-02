const domains = [
  "youtube.com",
  "youtu.be",
  "googlevideo.com",
  "ytimg.com",
  "googleapis.com",
  "gstatic.com",
  "doubleclick.net",
  "ggpht.com",
  "withgoogle.com",
  "chatgpt.com",
  "browser-intake-datadoghq.com",
  "proton.me",
  "facebook.com",
  "fbcdn.net",
  "instagram.com",
  "redis.io",
  "x.com",
  "twimg.com",
  "medium.com",
]

function FindProxyForURL(url, host) {
  if (domains.some(domain => dnsDomainIs(host, domain))) {
    return "SOCKS5 127.0.0.1:7010"
  }

  return "DIRECT"
}
