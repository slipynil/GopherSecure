package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"telegram-service/internal/dto"
)

type client struct {
	http *http.Client
	url  string
}

// client constructor
func New(endpoint string) *client {
	return &client{http: new(http.Client), url: endpoint}
}

// adds a new peer, use method post, and returns the response body with publicKey
func (c *client) AddPeer(hostID int, DNS bool, telegramID int64) (*dto.Response, error) {
	virtualEndpoint := fmt.Sprintf("10.66.66.%d/32", hostID)

	dns := ""
	if DNS {
		dns = "1.1.1.1, 8.8.8.8"
	}

	reqStruct := dto.AddPeerRequest{
		ID:              telegramID,
		VirtualEndpoint: virtualEndpoint,
		DNS:             dns,
	}
	// parse request
	reqBytes, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	data := bytes.NewReader(reqBytes)

	url := fmt.Sprintf("%s/peers", c.url)

	// get response
	resp, err := c.http.Post(url, "application/json", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return responseDecode(resp)
}

func (c *client) DeletePeer(publicKey string) error {

	url := fmt.Sprintf("%s/peers", c.url)
	reqStruct := dto.DelPeerRequest{PublicKey: publicKey}

	reqBytes, err := json.Marshal(reqStruct)
	if err != nil {
		return err
	}

	buf := bytes.NewReader(reqBytes)
	req, err := http.NewRequest("DELETE", url, buf)
	if err != nil {
		return err
	}

	// send request
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = responseDecode(resp)
	return err
}

func (c *client) DownloadConfFile(telegramID int64) ([]byte, error) {
	url := fmt.Sprintf("%s/peers/%d/config", c.url, telegramID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to download config file, status %d", resp.StatusCode)
		return nil, err
	}

	// read body to buffer
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return data, nil
}

func responseDecode(resp *http.Response) (*dto.Response, error) {
	respStruct := dto.Response{}
	err := json.NewDecoder(resp.Body).Decode(&respStruct)
	if err != nil {
		return nil, err
	}

	if len(respStruct.Error) != 0 {
		err := fmt.Errorf("%v, status code: %v", respStruct.Error, resp.StatusCode)
		return nil, err
	}
	return &respStruct, nil
}
