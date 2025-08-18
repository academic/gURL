# gURL
gURL is a modern HTTP client tool with support for HTTP/1.0, HTTP/1.1, HTTP/2, and HTTP/3 protocols. It's designed as a curl-like command-line utility for making HTTP requests with various options and configurations.

## Available Options

| Option | Short | Description | Status |
|--------|-------|-------------|--------|
| **HTTP Methods** |
| `GET` | | Send GET request to specified URL | ✅ |
| `POST` | | Send POST request to specified URL | ✅ |
| `PUT` | | Send PUT request to specified URL | ✅ |
| `DELETE` | | Send DELETE request to specified URL | ✅ |
| `HEAD` | | Send HEAD request to specified URL | ✅ |
| `OPTIONS` | | Send OPTIONS request to specified URL | ✅ |
| `PATCH` | | Send PATCH request to specified URL | ✅ |
| **HTTP Protocol Versions** |
| `--http1.0` | `-0` | Force HTTP/1.0 | ✅ |
| `--http1.1` | | Force HTTP/1.1 (default) | ✅ |
| `--http2` | | Force HTTP/2 | ✅ |
| `--http3` | | Force HTTP/3 | ✅ |
| **Request Options** |
| `--header <header>` | `-H` | Pass custom header(s) to server | ✅ |
| `--data <data>` | `-d` | HTTP POST data | ✅ |
| `--json <data>` | | HTTP POST JSON data | ✅ |
| `--form <name=content>` | `-F` | Specify multipart MIME data | ✅ |
| `--upload-file <file>` | `-T` | Transfer local FILE to destination | ✅ |
| `--request <method>` | `-X` | Specify request command to use | ✅ |
| `--user-agent <name>` | `-A` | Send User-Agent to server | ✅ |
| `--referer <URL>` | `-e` | Referrer URL | ✅ |
| **Authentication & Security** |
| `--insecure` | `-k` | Allow insecure server connections when using SSL | ✅ |
| `--cert <certificate[:password]>` | `-E` | Client certificate file and password | ✅ |
| `--basic` | | Use HTTP Basic Authentication | ❌ |
| `--digest` | | Use HTTP Digest Authentication | ❌ |
| `--ntlm` | | Use HTTP NTLM authentication | ❌ |
| `--negotiate` | | Use HTTP Negotiate (SPNEGO) authentication | ❌ |
| **Proxy Options** |
| `--proxy <[protocol://]host[:port]>` | `-x` | Use this proxy | ✅ |
| `--proxy-user <user:password>` | `-U` | Proxy user and password | ✅ |
| `--proxy-basic` | | Use Basic authentication on the proxy | ✅ |
| `--proxy-digest` | | Use Digest authentication on the proxy | ✅ |
| `--proxy-ntlm` | | Use NTLM authentication on the proxy | ✅ |
| `--proxy-negotiate` | | Use HTTP Negotiate authentication on the proxy | ✅ |
| **Cookie Management** |
| `--cookie <data\|filename>` | `-b` | Send cookies from string/file | ✅ |
| `--cookie-jar <filename>` | `-c` | Write cookies to filename after operation | ❌ |
| **Output Control** |
| `--output <file>` | `-o` | Write to file instead of stdout | ✅ |
| `--verbose` | `-v` | Make the operation more talkative | ✅ |
| `--silent` | `-s` | Silent mode | ✅ |
| `--include` | `-i` | Include protocol response headers in the output | ✅ |
| `--head` | `-I` | Show document info only | ✅ |
| **Connection Options** |
| `--max-time <seconds>` | `-m` | Maximum time allowed for the transfer | ✅ |
| `--max-redirs <num>` | | Maximum number of redirects allowed | ✅ |
| `--location` | `-L` | Follow redirects | ✅ |
| `--connect-timeout <seconds>` | | Maximum time allowed for connection | ❌ |
| **SSL/TLS Options** |
| `--cacert <file>` | | CA certificate to verify peer against | ❌ |
| `--capath <dir>` | | CA directory to verify peer against | ❌ |
| `--cert-status` | | Verify the status of the server certificate | ❌ |
| `--cert-type <type>` | | Certificate file type (DER/PEM/ENG) | ❌ |
| `--ciphers <list>` | | SSL ciphers to use | ❌ |
| `--key <key>` | | Private key file name | ❌ |
| `--key-type <type>` | | Private key file type (DER/PEM/ENG) | ❌ |
| **Network Options** |
| `--interface <name>` | | Use network INTERFACE (or address) | ❌ |
| `--ipv4` | `-4` | Resolve names to IPv4 addresses | ❌ |
| `--ipv6` | `-6` | Resolve names to IPv6 addresses | ❌ |
| `--dns-servers <addresses>` | | DNS server addrs to use | ❌ |
| **Advanced Options** |
| `--compressed` | | Request compressed response | ❌ |
| `--limit-rate <speed>` | | Limit transfer speed to RATE | ❌ |
| `--retry <num>` | | Retry request if transient problems occur | ❌ |
| `--retry-delay <seconds>` | | Wait time between retries | ❌ |
| `--retry-max-time <seconds>` | | Retry only within this period | ❌ |
| `--fail` | `-f` | Fail silently on HTTP errors | ❌ |
| `--raw` | | HTTP POST raw data | ✅ |

### Legend
- ✅ **Implemented** - Feature is fully implemented and tested
- ❌ **Not Implemented** - Feature is planned but not yet implemented
