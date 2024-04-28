Dynu for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/dynu)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for Dynu, allowing you to manage DNS records.

## Authenticating

This package uses **API Token authentication**. Refer to the [Dynu documentation](https://www.dynu.com/Support/API) for more information.

Start by retrieving your API token (API-Key) from the [table on the API Credentials page](https://www.dynu.com/ControlPanel/APICredentials) to be able to make authenticated requests to the API.

## Tests

Several tests for the basic functionality of the real Dynu API are available. These tests are not run by default. Set the environment variables TEST_ZONE and TEST_API_TOKEN to enable the tests like so:

```
TEST_ZONE=example.com. TEST_API_TOKEN=dynu_api_token go test -v
```

If the tests fail, you can manually check and fix the DNS records on the [DDNS Services page](https://www.dynu.com/en-US/ControlPanel/DDNS).
