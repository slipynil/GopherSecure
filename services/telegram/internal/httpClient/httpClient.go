// Package httpclient предоставляет HTTP клиент для взаимодействия с API сервиса AWG.
// Клиент отвечает за управление WireGuard пирами и получение конфигурационных файлов.
package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"telegram-service/internal/dto"
)

// client представляет HTTP клиент для взаимодействия с AWG сервисом.
type client struct {
	http *http.Client
	url  string
}

// New создает новый HTTP клиент с указанным endpoint для AWG сервиса.
func New(endpoint string) *client {
	return &client{http: new(http.Client), url: endpoint}
}

// AddPeer добавляет новый WireGuard пир на AWG сервисе для пользователя с ID telegramID.
// Параметр hostID используется для создания виртуального IP адреса формата 10.66.66.{hostID}/32.
// Если DNS равен true, устанавливаются публичные DNS серверы 1.1.1.1 и 8.8.8.8.
// Возвращает ответ сервера содержащий публичный ключ пира.
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

// DeletePeer удаляет WireGuard пир с указанным publicKey с AWG сервиса.
// В случае ошибки возвращает описание проблемы.
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

// DownloadConfFile загружает конфигурационный файл WireGuard для пользователя с ID telegramID.
// Возвращает содержимое файла как массив байт.
// Если сервер возвращает статус код отличный от 200, возвращает ошибку.
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

// GetKeys получает оба ключа (public_key и preshared_key) для пира с ID id.
// Возвращает структуру KeysResponse содержащую оба ключа.
func (c *client) GetKeys(id int64) (*dto.KeysResponse, error) {
	url := fmt.Sprintf("%s/peers/%d/keys", c.url, id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to get keys, status %d", resp.StatusCode)
		return nil, err
	}

	// decode response
	apiResp := dto.Response{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Error) != 0 {
		return nil, fmt.Errorf("api error: %s", apiResp.Error)
	}

	// extract keys from response data
	keysResp := &dto.KeysResponse{}
	if dataMap, ok := apiResp.Data.(map[string]interface{}); ok {
		if pubKey, ok := dataMap["public_key"].(string); ok {
			keysResp.PublicKey = pubKey
		}
		if presharedKey, ok := dataMap["preshared_key"].(string); ok {
			keysResp.PresharedKey = presharedKey
		}
	}

	return keysResp, nil
}

// responseDecode декодирует JSON ответ от AWG сервиса в структуру [dto.Response].
// Если ответ содержит ошибку, возвращает ее описание вместе со статус кодом.
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
