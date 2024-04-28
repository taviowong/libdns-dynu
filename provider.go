// Package dynu implements a DNS record management client compatible
// with the libdns interfaces for dynu.
package dynu

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with dynu.
type Provider struct {
	// config fields (with snake_case json struct tags on exported fields)
	APIToken string `json:"api_token,omitempty"`

	Once   sync.Once
	Client *Client
}

func (p *Provider) Init() {
	p.Client = NewClient(p.APIToken)
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.Once.Do(func() { p.Init() })

	var libRecords []libdns.Record

	fqdn := zoneToFqdn(zone)

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, fqdn)
	if err != nil {
		return nil, err
	}

	// GET /dns/{id}/record
	dnsRecords, err := p.Client.GetRecords(ctx, dnsHostName.ID)
	if err != nil {
		return nil, err
	}

	for _, dnsRecord := range dnsRecords {
		libRecords = append(libRecords, dnsRecordToLibdnsRecord(dnsRecord))
	}

	return libRecords, nil
}

func dnsRecordToLibdnsRecord(dnsRecord DNSRecord) libdns.Record {
	var relativeName string
	if dnsRecord.NodeName == "" {
		relativeName = "@"
	} else {
		relativeName = dnsRecord.NodeName
	}

	libRecord := libdns.Record{
		ID:   fmt.Sprint(dnsRecord.ID),
		Type: dnsRecord.Type,
		Name: relativeName,
		TTL:  time.Duration(dnsRecord.TTL) * time.Second,
	}

	switch dnsRecord.Type {
	case "A":
		libRecord.Value = dnsRecord.Ipv4Address
	case "AAAA":
		libRecord.Value = dnsRecord.Ipv6Address
	case "CNAME":
		libRecord.Value = dnsRecord.Host
	case "MX":
		libRecord.Value = dnsRecord.Host
		libRecord.Priority = uint(dnsRecord.Priority)
	case "NS":
		libRecord.Value = dnsRecord.Host
	case "PTR":
		libRecord.Name = dnsRecord.Host
		libRecord.Value = dnsRecord.Hostname
	case "SPF":
		libRecord.Value = dnsRecord.TextData
	case "TXT":
		libRecord.Value = dnsRecord.TextData
	default:
		libRecord.Value = dnsRecord.Content
	}

	return libRecord
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.appendOrSetRecords(ctx, zone, records, true)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.appendOrSetRecords(ctx, zone, records, false)
}

// if ignoreRecordId is true, the records will be added even if record id is provided
func (p *Provider) appendOrSetRecords(ctx context.Context, zone string, records []libdns.Record, ignoreRecordId bool) ([]libdns.Record, error) {
	p.Once.Do(func() { p.Init() })

	var updatedRecords []libdns.Record
	var updateErrors []error

	fqdn := zoneToFqdn(zone)

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, fqdn)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		dnsRecord, err := libdnsRecordToDnsRecord(fqdn, rec)
		if err != nil {
			updateErrors = append(updateErrors, err)
			continue
		}

		// POST /dns/{id}/record[/{dnsRecordId}]
		updateResponse, err := p.Client.AddOrUpdateRecord(ctx, dnsHostName.ID, dnsRecord, ignoreRecordId)

		if err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("dnsRecord %+v: %w", rec, err))
		} else {
			updatedRecords = append(updatedRecords, dnsRecordToLibdnsRecord(*updateResponse))
		}
	}

	return updatedRecords, errors.Join(updateErrors...)
}

func libdnsRecordToDnsRecord(fqdn string, record libdns.Record) (DNSRecord, error) {
	var id int64
	fmt.Sscan(record.ID, &id)

	dnsRecord := DNSRecord{
		ID:       id,
		Type:     record.Type,
		NodeName: record.Name,
		TTL:      int(record.TTL.Seconds()),
		State:    true, // must be set to true to take effect
	}

	var err error

	switch record.Type {
	case "A":
		dnsRecord.Ipv4Address = record.Value
	case "AAAA":
		dnsRecord.Ipv6Address = record.Value
	case "CNAME":
		dnsRecord.Host = record.Value
	case "MX":
		dnsRecord.Host = record.Value
		dnsRecord.Priority = int(record.Priority)
	case "NS":
		dnsRecord.Host = record.Value
	case "PTR":
		dnsRecord.Host = record.Name
		dnsRecord.NodeName = libdns.RelativeName(record.Value, fqdn) // seems Dynu can only point to subdomain; get relative name from input
	case "SPF":
		dnsRecord.TextData = record.Value
	case "TXT":
		dnsRecord.TextData = record.Value
	default:
		err = fmt.Errorf("dnsRecord %+v: record type not implemented", record)
	}

	if dnsRecord.NodeName == "@" {
		dnsRecord.NodeName = ""
	}

	return dnsRecord, err
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.Once.Do(func() { p.Init() })

	var deletedRecords []libdns.Record
	var deleteErrors []error

	fqdn := zoneToFqdn(zone)

	// GET /dns/getroot/{hostname}
	dnsHostName, err := p.Client.GetRootDomain(ctx, fqdn)
	if err != nil {
		return nil, err
	}

	// DELETE /dns/{id}/record/{dnsRecordId}
	for _, rec := range records {
		err := p.Client.DeleteRecord(ctx, dnsHostName.ID, rec.ID)

		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("dnsRecordId %s: %w", rec.ID, err))
		} else {
			deletedRecords = append(deletedRecords, rec)
		}
	}

	return deletedRecords, errors.Join(deleteErrors...)
}

func zoneToFqdn(zone string) string {
	// we trim the dot at the end of the zone name to get the fqdn
	return strings.TrimRight(zone, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
