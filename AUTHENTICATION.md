# Authentication Features in gURL

gURL now supports multiple HTTP authentication methods similar to curl.

## Basic Authentication

Basic authentication is the most common and simplest form of HTTP authentication.

### Usage

```bash
# Explicit basic auth
gURL GET https://example.com/protected --basic --user username:password

# Basic auth is the default when --user is specified without an auth type
gURL GET https://example.com/protected --user username:password

# Using the short form
gURL GET https://example.com/protected -u username:password
```

### How it works

Basic authentication encodes the username and password in Base64 and sends them in the `Authorization` header:
```
Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

## Digest Authentication

Digest authentication is more secure than basic auth as it doesn't send passwords in plain text.

### Usage

```bash
gURL GET https://example.com/protected --digest --user username:password
```

### Current Implementation Status

**Note**: The current digest implementation is basic and handles simple cases. Full digest authentication requires a challenge-response mechanism that involves:

1. Making an initial request
2. Receiving a 401 response with `Www-Authenticate` header containing challenge parameters
3. Computing a digest response using MD5 hashing
4. Making a second request with the computed digest

The current implementation sets the auth type but may require enhancements for full compatibility with all digest authentication scenarios.

## NTLM Authentication

NTLM (NT LAN Manager) authentication is primarily used in Windows environments.

### Usage

```bash
gURL GET https://example.com/protected --ntlm --user domain\\username:password
```

### Current Implementation Status

**Note**: NTLM authentication requires a complex multi-step handshake. The current implementation is a placeholder that sets the appropriate auth type. Full NTLM support would require implementing the NTLM protocol specification.

## Negotiate Authentication (SPNEGO)

Negotiate authentication uses SPNEGO (Simple and Protected GSSAPI Negotiation Mechanism) and is commonly used with Kerberos.

### Usage

```bash
gURL GET https://example.com/protected --negotiate --user username:password
```

### Current Implementation Status

**Note**: Negotiate authentication requires GSSAPI/Kerberos integration. The current implementation is a placeholder that sets the appropriate auth type. Full Negotiate support would require GSSAPI libraries and proper Kerberos configuration.

## Examples

### Testing with httpbin.org

```bash
# Test basic auth
gURL GET https://httpbin.org/basic-auth/user/pass --user user:pass

# Test with verbose output to see headers
gURL GET https://httpbin.org/get --basic --user myuser:mypass -v

# Test digest auth (will show challenge)
gURL GET https://httpbin.org/digest-auth/auth/user/pass --digest --user user:pass
```

### Integration with other flags

Authentication works with all other gURL features:

```bash
# With custom headers
gURL GET https://example.com/api --user token:secret -H "Accept: application/json"

# With POST data
gURL POST https://example.com/api --user admin:pass -d "data=value"

# With file upload
gURL POST https://example.com/upload --user user:pass -T file.txt
```

## Security Considerations

- **Basic Auth**: Credentials are Base64 encoded (not encrypted). Always use HTTPS.
- **Digest Auth**: More secure than basic but still vulnerable to certain attacks. Use HTTPS when possible.
- **NTLM/Negotiate**: Designed for enterprise environments with proper domain/Kerberos setup.

Always use HTTPS when transmitting credentials to prevent interception.
