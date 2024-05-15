package dynu

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/stretchr/testify/assert"
)

var zone = os.Getenv("TEST_ZONE")
var apiToken = os.Getenv("TEST_API_TOKEN")
var testRealApi = zone != "" && apiToken != ""

var domain = "dynu.com"
var ownDomain = "my.dynu.com"

func checkSkipApiTest(t *testing.T) {
	if !testRealApi {
		t.Skip("Env variables not set. Skipping api test.")
	}
}

func TestGetRecords(t *testing.T) {
	checkSkipApiTest(t)

	ctx := context.TODO()

	provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}

	recs, err := provider.GetRecords(ctx, zone)

	if !assert.NoError(t, err) {
		return
	}

	for _, rec := range recs {
		t.Log(rec)
	}
}

func TestAddAndDeleteTxtRecord(t *testing.T) {
	checkSkipApiTest(t)

	ctx := context.TODO()

	provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}
	testRecord := libdns.Record{
		Type:  "TXT",
		Name:  "@",
		Value: "TEST TXT RECORD",
		TTL:   time.Duration(120) * time.Second,
	}

	addedRecords, err := provider.AppendRecords(ctx, zone, []libdns.Record{testRecord})
	if !assert.NoError(t, err) {
		return
	}

	for _, rec := range addedRecords {
		assert.NotEmpty(t, rec.ID)

		assert.Equal(t, testRecord.Type, rec.Type)
		assert.Equal(t, testRecord.Name, rec.Name)
		assert.Equal(t, testRecord.Value, rec.Value)
		assert.Equal(t, testRecord.TTL, rec.TTL)
	}

	deletedRecords, err := provider.DeleteRecords(ctx, zone, addedRecords)
	if !assert.NoError(t, err) {
		return
	}

	for _, rec := range deletedRecords {
		t.Log(rec)
	}
}

func TestAddUpdateAndDeleteTxtRecord(t *testing.T) {
	checkSkipApiTest(t)

	ctx := context.TODO()

	provider := Provider{APIToken: apiToken, OwnDomain: zoneToFqdn(zone)}
	testRecord := libdns.Record{
		Type:  "TXT",
		Name:  "test",
		Value: "TEST TXT RECORD",
		TTL:   time.Duration(120) * time.Second,
	}

	addedRecords, err := provider.AppendRecords(ctx, zone, []libdns.Record{testRecord})
	if !assert.NoError(t, err) {
		return
	}
	addedId := addedRecords[0].ID

	testRecord.ID = addedId
	testRecord.Value = "TEST UPDATED TXT RECORD"
	addedRecords, err = provider.SetRecords(ctx, zone, []libdns.Record{testRecord})
	if !assert.NoError(t, err) {
		return
	}
	updatedId := addedRecords[0].ID

	assert.Equal(t, addedId, updatedId, "Added record and updated record should have same ID")

	for _, rec := range addedRecords {
		assert.NotEmpty(t, rec.ID)

		assert.Equal(t, testRecord.Type, rec.Type)
		assert.Equal(t, testRecord.Name, rec.Name)
		assert.Equal(t, testRecord.Value, rec.Value)
		assert.Equal(t, testRecord.TTL, rec.TTL)
	}

	deletedRecords, err := provider.DeleteRecords(ctx, zone, addedRecords)
	if !assert.NoError(t, err) {
		return
	}

	for _, rec := range deletedRecords {
		t.Log(rec)
	}
}

func Test_dnsRecordToLibdnsRecord_Basic(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "123", libdnsRecord.ID)
	assert.Equal(t, "A", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, time.Duration(120)*time.Second, libdnsRecord.TTL)
}

func Test_dnsRecordToLibdnsRecord_EmptyNodeNameDomain(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.NodeName = ""
	dnsRecord.DomainName = "dynu.com"
	dnsRecord.Hostname = "dynu.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "@", libdnsRecord.Name)
}

func Test_dnsRecordToLibdnsRecord_EmptyNodeNameSubdomain(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.NodeName = ""
	dnsRecord.Hostname = "my.dynu.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "my", libdnsRecord.Name)
}

func Test_dnsRecordToLibdnsRecord_A(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "A"
	dnsRecord.Ipv4Address = "1.2.3.4"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "A", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "1.2.3.4", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_AAAA(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "AAAA"
	dnsRecord.Ipv6Address = "::1"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "AAAA", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "::1", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_CNAME(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "CNAME"
	dnsRecord.Host = "www.example.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "CNAME", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "www.example.com", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_MX(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "MX"
	dnsRecord.Host = "www.example.com"
	dnsRecord.Priority = 1
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "MX", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "www.example.com", libdnsRecord.Value)
	assert.Equal(t, uint(1), libdnsRecord.Priority)
}

func Test_dnsRecordToLibdnsRecord_NS(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "NS"
	dnsRecord.Host = "www.example.com"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "NS", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "www.example.com", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_PTR(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "PTR"
	dnsRecord.Host = "10.207.160.216.in-addr.arpa"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "PTR", libdnsRecord.Type)
	assert.Equal(t, "10.207.160.216.in-addr.arpa", libdnsRecord.Name)
	assert.Equal(t, "abc.my.dynu.com", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_SPF(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "SPF"
	dnsRecord.TextData = "ABCD"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "SPF", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "ABCD", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_TXT(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "TXT"
	dnsRecord.TextData = "ABCD"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "TXT", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "ABCD", libdnsRecord.Value)
}

func Test_dnsRecordToLibdnsRecord_UNKNOWN(t *testing.T) {
	dnsRecord := getBasicDnsRecord()
	dnsRecord.Type = "UNKNOWN"
	dnsRecord.Content = "CONTENT"
	libdnsRecord := dnsRecordToLibdnsRecord(dnsRecord, domain)

	assert.Equal(t, "UNKNOWN", libdnsRecord.Type)
	assert.Equal(t, "abc.my", libdnsRecord.Name)
	assert.Equal(t, "CONTENT", libdnsRecord.Value)
}

func getBasicDnsRecord() DNSRecord {
	return DNSRecord{
		ID:         123,
		Type:       "A",
		NodeName:   "abc",
		DomainName: "my.dynu.com",
		Hostname:   "abc.my.dynu.com",
		TTL:        120,
	}
}

func Test_libdnsRecordToDnsRecord_Basic(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(123), dnsRecord.ID)
	assert.Equal(t, "A", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, 120, dnsRecord.TTL)
	assert.Equal(t, true, dnsRecord.State)
}

func Test_libdnsRecordToDnsRecord_EmptyNodeNameDomain(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Name = "@"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, domain) // test the case where domain is same as owned domain

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "", dnsRecord.NodeName)
}

func Test_libdnsRecordToDnsRecord_EmptyNodeNameSubdomain(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Name = "my"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "", dnsRecord.NodeName)
}

func Test_libdnsRecordToDnsRecord_A(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "A"
	libdnsRecord.Value = "1.2.3.4"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "A", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "1.2.3.4", dnsRecord.Ipv4Address)
}

func Test_libdnsRecordToDnsRecord_AAAA(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "AAAA"
	libdnsRecord.Value = "::1"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "AAAA", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "::1", dnsRecord.Ipv6Address)
}

func Test_libdnsRecordToDnsRecord_CNAME(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "CNAME"
	libdnsRecord.Value = "www.example.com"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "CNAME", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_MX(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "MX"
	libdnsRecord.Value = "www.example.com"
	libdnsRecord.Priority = 1
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "MX", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
	assert.Equal(t, 1, dnsRecord.Priority)
}

func Test_libdnsRecordToDnsRecord_NS(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "NS"
	libdnsRecord.Value = "www.example.com"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "NS", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "www.example.com", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_PTR(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "PTR"
	libdnsRecord.Name = "10.207.160.216.in-addr.arpa"
	libdnsRecord.Value = "abc.my.dynu.com"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "PTR", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "10.207.160.216.in-addr.arpa", dnsRecord.Host)
}

func Test_libdnsRecordToDnsRecord_SPF(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "SPF"
	libdnsRecord.Value = "ABCD"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "SPF", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "ABCD", dnsRecord.TextData)
}

func Test_libdnsRecordToDnsRecord_TXT(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "TXT"
	libdnsRecord.Value = "ABCD"
	dnsRecord, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "TXT", dnsRecord.Type)
	assert.Equal(t, "abc", dnsRecord.NodeName)
	assert.Equal(t, "ABCD", dnsRecord.TextData)
}

func Test_libdnsRecordToDnsRecord_UNKNOWN(t *testing.T) {
	libdnsRecord := getBasicLibDnsRecord()
	libdnsRecord.Type = "UNKNOWN"
	libdnsRecord.Value = "CONTENT"
	_, err := libdnsRecordToDnsRecord(libdnsRecord, domain, ownDomain)

	assert.Error(t, err)
}

func getBasicLibDnsRecord() libdns.Record {
	return libdns.Record{
		ID:   "123",
		Type: "A",
		Name: "abc.my",
		TTL:  time.Duration(120) * time.Second,
	}
}
