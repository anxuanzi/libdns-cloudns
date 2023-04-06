package cloudns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libdns/libdns"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ClouDNS API docs: https://www.cloudns.net/wiki/article/41/

var baseUrl = "https://api.cloudns.net/dns/"

// Provider facilitates DNS record manipulation with <TODO: PROVIDER NAME>.
type Provider struct {
	AuthId       string `json:"auth_id"`
	SubAuthId    string `json:"sub_auth_id"`
	AuthPassword string `json:"auth_password"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	bUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	endpoint := bUrl.JoinPath("records.json")
	resp, err := p.sendGetRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
	})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var apiResult map[string]ApiDnsRecord
	err = json.Unmarshal(respContent, &apiResult)
	if err != nil {
		return nil, errors.New(string(respContent))
	}
	var records []libdns.Record
	for _, v := range apiResult {
		records = append(records, libdns.Record{
			ID:    v.Id,
			Type:  v.Type,
			Name:  v.Host,
			TTL:   parseDuration(v.Ttl + "s"),
			Value: v.Record,
		})
	}
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	bUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	endpoint := bUrl.JoinPath("add-record.json")
	resultRec := make([]libdns.Record, 0)
	for _, record := range records {
		resp, err := p.sendPostRequest(ctx, endpoint, map[string]string{
			"domain-name": zone,
			"record-type": record.Type,
			"host":        record.Name,
			"record":      record.Value,
			"ttl":         strconv.Itoa(ttlRounder(record.TTL)),
		})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var resultModel *ApiResponse
		err = json.Unmarshal(respContent, &resultModel)
		if err != nil {
			return nil, errors.New(string(respContent))
		}
		if resultModel.Status != "Success" {
			return nil, errors.New(resultModel.StatusDescription)
		}
		record.ID = strconv.Itoa(resultModel.Data.Id)
		resp.Body.Close()
		resultRec = append(resultRec, record)
	}
	return resultRec, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	recordsOnServ, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	resultRec := make([]libdns.Record, 0)
	for _, record := range records {
		found := false
		for _, recordOnServ := range recordsOnServ {
			if recordOnServ.ID == record.ID {
				found = true
				break
			}
		}
		if found {
			// exist, update
			bUrl, err := url.Parse(baseUrl)
			if err != nil {
				return nil, err
			}
			endpoint := bUrl.JoinPath("mod-record.json")
			resp, err := p.sendPostRequest(ctx, endpoint, map[string]string{
				"domain-name": zone,
				"record-id":   record.ID,
				"host":        record.Name,
				"record":      record.Value,
				"ttl":         strconv.Itoa(ttlRounder(record.TTL)),
			})
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			respContent, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			var resultModel *ApiResponse
			err = json.Unmarshal(respContent, &resultModel)
			if err != nil {
				return nil, errors.New(string(respContent))
			}
			if resultModel.Status != "Success" {
				return nil, errors.New(resultModel.StatusDescription)
			}
			record.ID = strconv.Itoa(resultModel.Data.Id)
			resp.Body.Close()
		} else {
			// not exist, create
			addedRecord, err := p.AppendRecords(ctx, zone, []libdns.Record{record})
			if err != nil {
				return nil, err
			}
			record.ID = addedRecord[0].ID
		}
		resultRec = append(resultRec, record)
	}
	return resultRec, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	bUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	endpoint := bUrl.JoinPath("delete-record.json")
	for _, record := range records {
		resp, err := p.sendPostRequest(ctx, endpoint, map[string]string{
			"domain-name": zone,
			"record-id":   record.ID,
		})
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		respContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var resultModel *ApiResponse
		err = json.Unmarshal(respContent, &resultModel)
		if err != nil {
			return nil, errors.New(string(respContent))
		}
		if resultModel.Status != "Success" {
			return nil, errors.New(resultModel.StatusDescription)
		}
		record.ID = strconv.Itoa(resultModel.Data.Id)
		resp.Body.Close()
	}
	return records, nil
}

func (p *Provider) sendPostRequest(ctx context.Context, reqUrl *url.URL, payload map[string]string) (*http.Response, error) {
	queries := reqUrl.Query()
	//fill in auth params
	if p.SubAuthId != "" {
		queries.Set("sub-auth-id", p.SubAuthId)
	} else {
		queries.Set("auth-id", p.AuthId)
	}
	queries.Set("auth-password", p.AuthPassword)

	//fill in payload
	for k, v := range payload {
		queries.Set(k, v)
	}

	reqUrl.RawQuery = queries.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func (p *Provider) sendGetRequest(ctx context.Context, reqUrl *url.URL, payload map[string]string) (*http.Response, error) {
	queries := reqUrl.Query()
	//fill in auth params
	if p.SubAuthId != "" {
		queries.Set("sub-auth-id", p.SubAuthId)
	} else {
		queries.Set("auth-id", p.AuthId)
	}
	queries.Set("auth-password", p.AuthPassword)

	//fill in payload
	for k, v := range payload {
		queries.Set(k, v)
	}

	reqUrl.RawQuery = queries.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
