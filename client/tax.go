package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dimaxgl/rus-tax-client/api"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

const (
	defaultEndpoint = `https://proverkacheka.nalog.ru:9999`
	defaultName     = `test_name`

	requestContentType = `application/json`

	endpointRegister    = `v1/mobile/users/signup`
	endpointRestore     = `v1/mobile/users/restore`
	endpointLogin       = `v1/mobile/users/login`
	endpointBillCheck   = "v1/ofds/*/inns/*/fss/%s/operations/1/tickets/%s?fiscalSign=%s&date=2018-05-17T17:57:00&sum=%f"
	endpointBillDetails = "v1/inns/*/kkts/*/fss/%s/tickets/%s?fiscalSign=%s&sendToEmail=no"

	deviceHeaderId = `device-id`
	deviceHeaderOs = `device-os`
)

type taxClient struct {
	phone       string
	apiEndpoint string
	cli         *http.Client
	token       string
}

func (c *taxClient) getBasicAuthToken() string {
	auth := fmt.Sprintf("%s:%s", c.phone, c.token)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *taxClient) setAuthHeader(req *http.Request) {
	req.Header.Add(`Authorization`, fmt.Sprintf("Basic %s", c.getBasicAuthToken()))
}

func (c *taxClient) Register(email string) error {
	req := api.TaxRegisterRequest{
		Name:  defaultName,
		Email: email,
		Phone: c.phone,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, `failed to marshal JSON request`)
	}

	resp, err := c.cli.Post(fmt.Sprintf("%s/%s", c.apiEndpoint, endpointRegister), requestContentType, bytes.NewBuffer(reqBytes))
	if err != nil {
		return errors.Wrap(err, `failed to send request`)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `failed to read request body`)
	}

	if resp.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	return nil
}

func (c *taxClient) Login(smsPassword string) (*api.TaxLoginResponse, error) {
	c.token = smsPassword
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.apiEndpoint, endpointLogin), nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to make new request`)
	}

	c.setAuthHeader(req)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to make request`)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, `failed to get request body`)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	var loginResponse api.TaxLoginResponse

	if err = json.Unmarshal(body, &loginResponse); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal JSON response`)
	}

	return &loginResponse, nil
}

func (c *taxClient) Restore() error {
	req := api.TaxRestoreRequest{
		Phone: c.phone,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, `failed to marshal JSON request`)
	}

	resp, err := c.cli.Post(fmt.Sprintf("%s/%s", c.apiEndpoint, endpointRestore), requestContentType, bytes.NewBuffer(reqBytes))
	if err != nil {
		return errors.Wrap(err, `failed to send request`)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, `failed to read request body`)
	}

	if resp.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	return nil
}

func (c *taxClient) BillCheck(fiscalNumber, fiscalDocument, fiscalDocumentAttr string, total float64) error {
	req, err := http.NewRequest(http.MethodGet, c.apiEndpoint+`/`+fmt.Sprintf(endpointBillCheck, fiscalNumber, fiscalDocument, fiscalDocumentAttr, total), nil)
	if err != nil {
		return errors.Wrap(err, `failed to create request`)
	}

	c.setAuthHeader(req)

	resp, err := c.cli.Do(req)
	if err != nil {
		return errors.Wrap(err, `failed to do request`)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, `failed to read request body`)
		}
		return api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	return nil
}

func (c *taxClient) BillDetail(fiscalNumber, fiscalDocument, fiscalDocumentAttr string) (*api.TaxBillCheckResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.apiEndpoint+`/`+fmt.Sprintf(endpointBillDetails, fiscalNumber, fiscalDocument, fiscalDocumentAttr), nil)
	if err != nil {
		return nil, errors.Wrap(err, `failed to create request`)
	}

	c.setAuthHeader(req)
	req.Header.Add(deviceHeaderId, ``)
	req.Header.Add(deviceHeaderOs, ``)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, `failed to do request`)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, `failed to read request body`)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedHTTPStatus{Status: resp.StatusCode, Body: body}
	}

	var checkResponse api.TaxBillCheckResponse

	fmt.Println(string(body))

	if err = json.Unmarshal(body, &checkResponse); err != nil {
		return nil, errors.Wrap(err, `failed to unmarshal JSON response`)
	}

	return &checkResponse, nil
}

func NewTaxClient(phone string, opts ...taxClientOpt) (api.TaxClient, error) {
	var err error

	cli := &taxClient{phone: phone, apiEndpoint: defaultEndpoint, cli: http.DefaultClient}

	for _, opt := range opts {
		if err = opt(cli); err != nil {
			return nil, errors.Wrap(err, `failed to apply opt`)
		}
	}
	return cli, nil
}

type taxClientOpt func(c *taxClient) error

func WithEndpoint(url string) taxClientOpt {
	return func(c *taxClient) error {
		c.apiEndpoint = url
		return nil
	}
}

func WihtHTTPClient(cli *http.Client) taxClientOpt {
	return func(c *taxClient) error {
		c.cli = cli
		return nil
	}
}

func WithToken(token string) taxClientOpt {
	return func(c *taxClient) error {
		c.token = token
		return nil
	}
}
