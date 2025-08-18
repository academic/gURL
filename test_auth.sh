#!/bin/bash

echo "Testing gURL Authentication Features"
echo "==================================="
echo

echo "1. Testing Basic Authentication (with --basic flag):"
./gURL GET https://httpbin.org/basic-auth/user/pass --basic --user user:pass
echo

echo "2. Testing Basic Authentication (default, no auth type specified):"
./gURL GET https://httpbin.org/basic-auth/user/pass --user user:pass
echo

echo "3. Testing invalid credentials (should fail with 401):"
./gURL GET https://httpbin.org/basic-auth/user/pass --basic --user wrong:password
echo

echo "4. Testing that auth headers are properly sent (using GET endpoint that echoes headers):"
./gURL GET https://httpbin.org/get --basic --user testuser:testpass | grep -A 10 '"headers"'
echo

echo "5. Testing digest authentication flag (will show 401 with digest challenge - full digest auth requires challenge-response):"
./gURL GET https://httpbin.org/digest-auth/auth/user/pass --digest --user user:pass
echo

echo "Authentication testing complete!"
