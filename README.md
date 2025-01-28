# PROJECT MOVED TO https://github.com/libdns/cloudns

[//]: # (# ClouDNS for [`libdns`]&#40;https://github.com/libdns/libdns&#41;)

[//]: # ()

[//]: # ([![Go Reference]&#40;https://pkg.go.dev/badge/test.svg&#41;]&#40;https://pkg.go.dev/github.com/anxuanzi/libdns-cloudns&#41;)

[//]: # ()

[//]: # (This package implements the [libdns interfaces]&#40;https://github.com/libdns/libdns&#41; for ClouDNS, allowing you to manage)

[//]: # (DNS records.)

[//]: # ()

[//]: # (## Installation)

[//]: # ()

[//]: # (To install this package, use `go get`:)

[//]: # ()

[//]: # (```sh)

[//]: # (go get github.com/anxuanzi/libdns-cloudns)

[//]: # (```)

[//]: # ()

[//]: # (## Usage)

[//]: # ()

[//]: # (Here is an example of how to use this package to manage DNS records:)

[//]: # ()

[//]: # (```go)

[//]: # (package main)

[//]: # ()

[//]: # (import &#40;)

[//]: # (	"context")

[//]: # (	"fmt")

[//]: # (	"github.com/anxuanzi/libdns-cloudns")

[//]: # (	"github.com/libdns/libdns")

[//]: # (	"time")

[//]: # (&#41;)

[//]: # ()

[//]: # (func main&#40;&#41; {)

[//]: # (	provider := &cloudns.Provider{)

[//]: # (		AuthId:       "your_auth_id",)

[//]: # (		SubAuthId:    "your_sub_auth_id",)

[//]: # (		AuthPassword: "your_auth_password",)

[//]: # (	})

[//]: # ()

[//]: # (	ctx, cancel := context.WithTimeout&#40;context.Background&#40;&#41;, 30*time.Second&#41;)

[//]: # (	defer cancel&#40;&#41;)

[//]: # ()

[//]: # (	// Get records)

[//]: # (	records, err := provider.GetRecords&#40;ctx, "example.com"&#41;)

[//]: # (	if err != nil {)

[//]: # (		fmt.Printf&#40;"Failed to get records: %s\n", err&#41;)

[//]: # (		return)

[//]: # (	})

[//]: # (	fmt.Printf&#40;"Records: %+v\n", records&#41;)

[//]: # ()

[//]: # (	// Append a record)

[//]: # (	newRecord := libdns.Record{)

[//]: # (		Type:  "TXT",)

[//]: # (		Name:  "test",)

[//]: # (		Value: "test-value",)

[//]: # (		TTL:   300 * time.Second,)

[//]: # (	})

[//]: # (	addedRecords, err := provider.AppendRecords&#40;ctx, "example.com", []libdns.Record{newRecord}&#41;)

[//]: # (	if err != nil {)

[//]: # (		fmt.Printf&#40;"Failed to append record: %s\n", err&#41;)

[//]: # (		return)

[//]: # (	})

[//]: # (	fmt.Printf&#40;"Added Records: %+v\n", addedRecords&#41;)

[//]: # (})

[//]: # (```)

[//]: # ()

[//]: # (## Configuration)

[//]: # ()

[//]: # (The `Provider` struct has the following fields:)

[//]: # ()

[//]: # (- `AuthId` &#40;string&#41;: Your ClouDNS authentication ID.)

[//]: # (- `SubAuthId` &#40;string, optional&#41;: Your ClouDNS sub-authentication ID.)

[//]: # (- `AuthPassword` &#40;string&#41;: Your ClouDNS authentication password.)

[//]: # ()

[//]: # (## Testing)

[//]: # ()

[//]: # (To run the tests, you need to set up your ClouDNS credentials and zone in the test file `provider_test.go`. The tests)

[//]: # (require a live ClouDNS account.)

[//]: # ()

[//]: # (```go)

[//]: # (var &#40;)

[//]: # (TAuthId = "your_auth_id")

[//]: # (TSubAuthId = "your_sub_auth_id")

[//]: # (TAuthPassword = "your_auth_password")

[//]: # (TZone = "example.com")

[//]: # (&#41;)

[//]: # (```)

[//]: # ()

[//]: # (Run the tests using the following command:)

[//]: # ()

[//]: # (```sh)

[//]: # (go test ./...)

[//]: # (```)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.