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
	"time"
)

type Client struct {
	authId       string `json:"auth_id"`
	subAuthId    string `json:"sub_auth_id"`
	authPassword string `json:"auth_password"`
	baseUrl      *url.URL
}

func UseClient(authId, subAuthId, authPassword string) *Client {
	burl, _ := url.Parse("https://api.cloudns.net/dns/")
	return &Client{
		authId:       authId,
		subAuthId:    subAuthId,
		authPassword: authPassword,
		baseUrl:      burl,
	}
}

// GetRecords
// @Description: Get all records of a zone
// @Param ctx
// @Param zone
func (c *Client) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	endpoint := c.baseUrl.JoinPath("records.json")
	resp, err := c.sendGetRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
	})
	defer resp.Body.Close()
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

// GetRecord
// @Description: Get a record of a zone
// @Param ctx
// @Param zone
// @Param recordId
func (c *Client) GetRecord(ctx context.Context, zone string, recordId string) (*libdns.Record, error) {
	rs, err := c.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	for _, v := range rs {
		if v.ID == recordId {
			return &v, nil
		}
	}
	return nil, errors.New("record not found")
}

// AddRecord
// @Description: Add a record to a zone
// @Param ctx
// @Param zone
// @Param recordType
// @Param host
// @Param record
// @Param ttl
func (c *Client) AddRecord(ctx context.Context, zone string, recordType string, host string, record string, ttl time.Duration) (*libdns.Record, error) {
	endpoint := c.baseUrl.JoinPath("add-record.json")
	resp, err := c.sendPostRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
		"record-type": recordType,
		"host":        host,
		"record":      record,
		"ttl":         strconv.Itoa(ttlRounder(ttl)),
	})
	defer resp.Body.Close()
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
	return &libdns.Record{
		ID:    strconv.Itoa(resultModel.Data.Id),
		Type:  recordType,
		Name:  host,
		TTL:   parseDuration(strconv.Itoa(ttlRounder(ttl)) + "s"),
		Value: record,
	}, nil
}

// UpdateRecord
// @Description: Update a record of a zone
// @Param ctx
// @Param zone
// @Param recordId
// @Param host
// @Param record
// @Param ttl
func (c *Client) UpdateRecord(ctx context.Context, zone string, recordId string, host string, record string, ttl time.Duration) (*libdns.Record, error) {
	endpoint := c.baseUrl.JoinPath("mod-record.json")
	resp, err := c.sendPostRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
		"record-id":   recordId,
		"host":        host,
		"record":      record,
		"ttl":         strconv.Itoa(ttlRounder(ttl)),
	})
	defer resp.Body.Close()
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
	getRecord, err := c.GetRecord(ctx, zone, recordId)
	if err != nil {
		return nil, err
	}
	return &libdns.Record{
		ID:    recordId,
		Type:  getRecord.Type,
		Name:  host,
		TTL:   parseDuration(strconv.Itoa(ttlRounder(ttl)) + "s"),
		Value: record,
	}, nil
}

// DeleteRecord
// @Description: Delete a record of a zone
// @Param ctx
// @Param zone
// @Param recordId
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) (*libdns.Record, error) {
	rInfo, err := c.GetRecord(ctx, zone, recordId)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		} else {
			return nil, err
		}
	}
	endpoint := c.baseUrl.JoinPath("delete-record.json")
	resp, err := c.sendPostRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
		"record-id":   recordId,
	})
	defer resp.Body.Close()
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
	return rInfo, nil
}

// sendPostRequest
// @Description: Send a post request to the API
// @Param ctx
// @Param reqUrl
// @Param payload
func (c *Client) sendPostRequest(ctx context.Context, reqUrl *url.URL, payload map[string]string) (*http.Response, error) {
	queries := reqUrl.Query()
	//fill in auth params
	if c.subAuthId != "" {
		queries.Set("sub-auth-id", c.subAuthId)
	} else {
		queries.Set("auth-id", c.authId)
	}
	queries.Set("auth-password", c.authPassword)

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

// sendGetRequest
// @Description: Send a get request to the API
// @Param ctx
// @Param reqUrl
// @Param payload
func (c *Client) sendGetRequest(ctx context.Context, reqUrl *url.URL, payload map[string]string) (*http.Response, error) {
	queries := reqUrl.Query()
	//fill in auth params
	if c.subAuthId != "" {
		queries.Set("sub-auth-id", c.subAuthId)
	} else {
		queries.Set("auth-id", c.authId)
	}
	queries.Set("auth-password", c.authPassword)

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
