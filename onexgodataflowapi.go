// ONEx Dataflow API 0.0.1
// License: MIT

package onexgodataflowapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/ghodss/yaml"
	onexdataflowapi "github.com/open-network-experiments/onexgodataflowapi/onexdataflowapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type onexgodataflowapiApi struct {
	api
	grpcClient onexdataflowapi.OpenapiClient
	httpClient httpClient
}

// grpcConnect builds up a grpc connection
func (api *onexgodataflowapiApi) grpcConnect() error {
	if api.grpcClient == nil {
		ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.dialTimeout)
		defer cancelFunc()
		conn, err := grpc.DialContext(ctx, api.grpc.location, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		api.grpcClient = onexdataflowapi.NewOpenapiClient(conn)
		api.grpc.clientConnection = conn
	}
	return nil
}

func (api *onexgodataflowapiApi) grpcClose() error {
	if api.grpc != nil {
		if api.grpc.clientConnection != nil {
			err := api.grpc.clientConnection.Close()
			if err != nil {
				return err
			}
		}
	}
	api.grpcClient = nil
	api.grpc = nil
	return nil
}

func (api *onexgodataflowapiApi) Close() error {
	if api.hasGrpcTransport() {
		err := api.grpcClose()
		return err
	}
	if api.hasHttpTransport() {
		api.http = nil
		api.httpClient.client = nil
	}
	return nil
}

//  NewApi returns a new instance of the top level interface hierarchy
func NewApi() OnexgodataflowapiApi {
	api := onexgodataflowapiApi{}
	return &api
}

// httpConnect builds up a http connection
func (api *onexgodataflowapiApi) httpConnect() error {
	if api.httpClient.client == nil {
		var verify = !api.http.verify
		client := httpClient{
			client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: verify},
				},
			},
			ctx: context.Background(),
		}
		api.httpClient = client
	}
	return nil
}

func (api *onexgodataflowapiApi) httpSendRecv(urlPath string, jsonBody string, method string) (*http.Response, error) {
	err := api.httpConnect()
	if err != nil {
		return nil, err
	}
	httpClient := api.httpClient
	var bodyReader = bytes.NewReader([]byte(jsonBody))
	queryUrl, err := url.Parse(api.http.location)
	if err != nil {
		return nil, err
	}
	queryUrl, _ = queryUrl.Parse(urlPath)
	req, _ := http.NewRequest(method, queryUrl.String(), bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(httpClient.ctx)
	return httpClient.client.Do(req)
}

// OnexgodataflowapiApi the Open Network Experiments Dataflow API and Data Models
type OnexgodataflowapiApi interface {
	Api
	// NewConfig returns a new instance of Config.
	// Config is oNEx dataflow configuration
	NewConfig() Config
	// NewGetConfigDetails returns a new instance of GetConfigDetails.
	// GetConfigDetails is get config request details
	NewGetConfigDetails() GetConfigDetails
	// NewExperimentRequest returns a new instance of ExperimentRequest.
	// ExperimentRequest is experiment request details
	NewExperimentRequest() ExperimentRequest
	// NewControlStartRequest returns a new instance of ControlStartRequest.
	// ControlStartRequest is start request details
	NewControlStartRequest() ControlStartRequest
	// NewControlStatusRequest returns a new instance of ControlStatusRequest.
	// ControlStatusRequest is control.state request details
	NewControlStatusRequest() ControlStatusRequest
	// NewMetricsRequest returns a new instance of MetricsRequest.
	// MetricsRequest is metrics request details
	NewMetricsRequest() MetricsRequest
	// NewSetConfigResponse returns a new instance of SetConfigResponse.
	// SetConfigResponse is description is TBD
	NewSetConfigResponse() SetConfigResponse
	// NewGetConfigResponse returns a new instance of GetConfigResponse.
	// GetConfigResponse is description is TBD
	NewGetConfigResponse() GetConfigResponse
	// NewRunExperimentResponse returns a new instance of RunExperimentResponse.
	// RunExperimentResponse is description is TBD
	NewRunExperimentResponse() RunExperimentResponse
	// NewStartResponse returns a new instance of StartResponse.
	// StartResponse is description is TBD
	NewStartResponse() StartResponse
	// NewGetStatusResponse returns a new instance of GetStatusResponse.
	// GetStatusResponse is description is TBD
	NewGetStatusResponse() GetStatusResponse
	// NewGetMetricsResponse returns a new instance of GetMetricsResponse.
	// GetMetricsResponse is description is TBD
	NewGetMetricsResponse() GetMetricsResponse
	// SetConfig sets the ONEx dataflow configuration
	SetConfig(config Config) (Config, error)
	// GetConfig gets the ONEx dataflow config from the server, as currently configured
	GetConfig(getConfigDetails GetConfigDetails) (Config, error)
	// RunExperiment runs the currently configured dataflow experiment
	RunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error)
	// Start starts the currently configured dataflow experiment
	Start(controlStartRequest ControlStartRequest) (WarningDetails, error)
	// GetStatus gets the control status (e.g. started/completed/error)
	GetStatus(controlStatusRequest ControlStatusRequest) (ControlStatusResponse, error)
	// GetMetrics description is TBD
	GetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error)
}

func (api *onexgodataflowapiApi) NewConfig() Config {
	return NewConfig()
}

func (api *onexgodataflowapiApi) NewGetConfigDetails() GetConfigDetails {
	return NewGetConfigDetails()
}

func (api *onexgodataflowapiApi) NewExperimentRequest() ExperimentRequest {
	return NewExperimentRequest()
}

func (api *onexgodataflowapiApi) NewControlStartRequest() ControlStartRequest {
	return NewControlStartRequest()
}

func (api *onexgodataflowapiApi) NewControlStatusRequest() ControlStatusRequest {
	return NewControlStatusRequest()
}

func (api *onexgodataflowapiApi) NewMetricsRequest() MetricsRequest {
	return NewMetricsRequest()
}

func (api *onexgodataflowapiApi) NewSetConfigResponse() SetConfigResponse {
	return NewSetConfigResponse()
}

func (api *onexgodataflowapiApi) NewGetConfigResponse() GetConfigResponse {
	return NewGetConfigResponse()
}

func (api *onexgodataflowapiApi) NewRunExperimentResponse() RunExperimentResponse {
	return NewRunExperimentResponse()
}

func (api *onexgodataflowapiApi) NewStartResponse() StartResponse {
	return NewStartResponse()
}

func (api *onexgodataflowapiApi) NewGetStatusResponse() GetStatusResponse {
	return NewGetStatusResponse()
}

func (api *onexgodataflowapiApi) NewGetMetricsResponse() GetMetricsResponse {
	return NewGetMetricsResponse()
}

func (api *onexgodataflowapiApi) SetConfig(config Config) (Config, error) {

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpSetConfig(config)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.SetConfigRequest{Config: config.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.SetConfig(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewConfig().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) GetConfig(getConfigDetails GetConfigDetails) (Config, error) {

	err := getConfigDetails.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpGetConfig(getConfigDetails)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.GetConfigRequest{GetConfigDetails: getConfigDetails.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.GetConfig(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewConfig().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) RunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error) {

	err := experimentRequest.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpRunExperiment(experimentRequest)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.RunExperimentRequest{ExperimentRequest: experimentRequest.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.RunExperiment(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewWarningDetails().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) Start(controlStartRequest ControlStartRequest) (WarningDetails, error) {

	err := controlStartRequest.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpStart(controlStartRequest)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.StartRequest{ControlStartRequest: controlStartRequest.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.Start(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewWarningDetails().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) GetStatus(controlStatusRequest ControlStatusRequest) (ControlStatusResponse, error) {

	err := controlStatusRequest.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpGetStatus(controlStatusRequest)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.GetStatusRequest{ControlStatusRequest: controlStatusRequest.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.GetStatus(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewControlStatusResponse().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) GetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error) {

	err := metricsRequest.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpGetMetrics(metricsRequest)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdataflowapi.GetMetricsRequest{MetricsRequest: metricsRequest.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.GetMetrics(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewMetricsResponse().SetMsg(resp.GetStatusCode_200()), nil
	}
	if resp.GetStatusCode_400() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_400())
		return nil, fmt.Errorf(string(data))
	}
	if resp.GetStatusCode_500() != nil {
		data, _ := yaml.Marshal(resp.GetStatusCode_500())
		return nil, fmt.Errorf(string(data))
	}
	return nil, fmt.Errorf("response of 200, 400, 500 has not been implemented")
}

func (api *onexgodataflowapiApi) httpSetConfig(config Config) (Config, error) {
	configJson, err := config.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/config", configJson, "POST")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewSetConfigResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

func (api *onexgodataflowapiApi) httpGetConfig(getConfigDetails GetConfigDetails) (Config, error) {
	getConfigDetailsJson, err := getConfigDetails.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/config", getConfigDetailsJson, "GET")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewGetConfigResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

func (api *onexgodataflowapiApi) httpRunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error) {
	experimentRequestJson, err := experimentRequest.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/control/experiment", experimentRequestJson, "POST")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewRunExperimentResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

func (api *onexgodataflowapiApi) httpStart(controlStartRequest ControlStartRequest) (WarningDetails, error) {
	controlStartRequestJson, err := controlStartRequest.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/control/start", controlStartRequestJson, "POST")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewStartResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

func (api *onexgodataflowapiApi) httpGetStatus(controlStatusRequest ControlStatusRequest) (ControlStatusResponse, error) {
	controlStatusRequestJson, err := controlStatusRequest.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/control/status", controlStatusRequestJson, "GET")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewGetStatusResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

func (api *onexgodataflowapiApi) httpGetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error) {
	metricsRequestJson, err := metricsRequest.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("onex/api/v1/dataflow/results/metrics", metricsRequestJson, "POST")

	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		obj := api.NewGetMetricsResponse().StatusCode200()
		if err := obj.FromJson(string(bodyBytes)); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	if resp.StatusCode == 400 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	if resp.StatusCode == 500 {
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return nil, fmt.Errorf("response not implemented")
}

// ***** Config *****
type config struct {
	obj            *onexdataflowapi.Config
	hostsHolder    ConfigHostIter
	dataflowHolder Dataflow
}

func NewConfig() Config {
	obj := config{obj: &onexdataflowapi.Config{}}
	obj.setDefault()
	return &obj
}

func (obj *config) Msg() *onexdataflowapi.Config {
	return obj.obj
}

func (obj *config) SetMsg(msg *onexdataflowapi.Config) Config {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *config) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *config) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *config) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *config) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *config) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *config) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *config) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *config) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *config) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *config) setNil() {
	obj.hostsHolder = nil
	obj.dataflowHolder = nil
}

// Config is oNEx dataflow configuration
type Config interface {
	Msg() *onexdataflowapi.Config
	SetMsg(*onexdataflowapi.Config) Config
	// ToPbText marshals Config to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals Config to YAML text
	ToYaml() (string, error)
	// ToJson marshals Config to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals Config from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Config from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Config from JSON text
	FromJson(value string) error
	// Validate validates Config
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Hosts returns ConfigHostIter, set in Config
	Hosts() ConfigHostIter
	// Dataflow returns Dataflow, set in Config.
	// Dataflow is description is TBD
	Dataflow() Dataflow
	// SetDataflow assigns Dataflow provided by user to Config.
	// Dataflow is description is TBD
	SetDataflow(value Dataflow) Config
	// HasDataflow checks if Dataflow has been set in Config
	HasDataflow() bool
	setNil()
}

// Hosts returns a []Host
// description is TBD
func (obj *config) Hosts() ConfigHostIter {
	if len(obj.obj.Hosts) == 0 {
		obj.obj.Hosts = []*onexdataflowapi.Host{}
	}
	if obj.hostsHolder == nil {
		obj.hostsHolder = newConfigHostIter().setMsg(obj)
	}
	return obj.hostsHolder
}

type configHostIter struct {
	obj       *config
	hostSlice []Host
}

func newConfigHostIter() ConfigHostIter {
	return &configHostIter{}
}

type ConfigHostIter interface {
	setMsg(*config) ConfigHostIter
	Items() []Host
	Add() Host
	Append(items ...Host) ConfigHostIter
	Set(index int, newObj Host) ConfigHostIter
	Clear() ConfigHostIter
	clearHolderSlice() ConfigHostIter
	appendHolderSlice(item Host) ConfigHostIter
}

func (obj *configHostIter) setMsg(msg *config) ConfigHostIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Hosts {
		obj.appendHolderSlice(&host{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *configHostIter) Items() []Host {
	return obj.hostSlice
}

func (obj *configHostIter) Add() Host {
	newObj := &onexdataflowapi.Host{}
	obj.obj.obj.Hosts = append(obj.obj.obj.Hosts, newObj)
	newLibObj := &host{obj: newObj}
	newLibObj.setDefault()
	obj.hostSlice = append(obj.hostSlice, newLibObj)
	return newLibObj
}

func (obj *configHostIter) Append(items ...Host) ConfigHostIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Hosts = append(obj.obj.obj.Hosts, newObj)
		obj.hostSlice = append(obj.hostSlice, item)
	}
	return obj
}

func (obj *configHostIter) Set(index int, newObj Host) ConfigHostIter {
	obj.obj.obj.Hosts[index] = newObj.Msg()
	obj.hostSlice[index] = newObj
	return obj
}
func (obj *configHostIter) Clear() ConfigHostIter {
	if len(obj.obj.obj.Hosts) > 0 {
		obj.obj.obj.Hosts = []*onexdataflowapi.Host{}
		obj.hostSlice = []Host{}
	}
	return obj
}
func (obj *configHostIter) clearHolderSlice() ConfigHostIter {
	if len(obj.hostSlice) > 0 {
		obj.hostSlice = []Host{}
	}
	return obj
}
func (obj *configHostIter) appendHolderSlice(item Host) ConfigHostIter {
	obj.hostSlice = append(obj.hostSlice, item)
	return obj
}

// Dataflow returns a Dataflow
// description is TBD
func (obj *config) Dataflow() Dataflow {
	if obj.obj.Dataflow == nil {
		obj.obj.Dataflow = NewDataflow().Msg()
	}
	if obj.dataflowHolder == nil {
		obj.dataflowHolder = &dataflow{obj: obj.obj.Dataflow}
	}
	return obj.dataflowHolder
}

// Dataflow returns a Dataflow
// description is TBD
func (obj *config) HasDataflow() bool {
	return obj.obj.Dataflow != nil
}

// SetDataflow sets the Dataflow value in the Config object
// description is TBD
func (obj *config) SetDataflow(value Dataflow) Config {

	obj.dataflowHolder = nil
	obj.obj.Dataflow = value.Msg()

	return obj
}

func (obj *config) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Hosts != nil {

		if set_default {
			obj.Hosts().clearHolderSlice()
			for _, item := range obj.obj.Hosts {
				obj.Hosts().appendHolderSlice(&host{obj: item})
			}
		}
		for _, item := range obj.Hosts().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.Dataflow != nil {
		obj.Dataflow().validateObj(set_default)
	}

}

func (obj *config) setDefault() {

}

// ***** GetConfigDetails *****
type getConfigDetails struct {
	obj *onexdataflowapi.GetConfigDetails
}

func NewGetConfigDetails() GetConfigDetails {
	obj := getConfigDetails{obj: &onexdataflowapi.GetConfigDetails{}}
	obj.setDefault()
	return &obj
}

func (obj *getConfigDetails) Msg() *onexdataflowapi.GetConfigDetails {
	return obj.obj
}

func (obj *getConfigDetails) SetMsg(msg *onexdataflowapi.GetConfigDetails) GetConfigDetails {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getConfigDetails) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *getConfigDetails) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *getConfigDetails) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getConfigDetails) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *getConfigDetails) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getConfigDetails) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *getConfigDetails) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *getConfigDetails) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *getConfigDetails) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// GetConfigDetails is get config request details
type GetConfigDetails interface {
	Msg() *onexdataflowapi.GetConfigDetails
	SetMsg(*onexdataflowapi.GetConfigDetails) GetConfigDetails
	// ToPbText marshals GetConfigDetails to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals GetConfigDetails to YAML text
	ToYaml() (string, error)
	// ToJson marshals GetConfigDetails to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals GetConfigDetails from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetConfigDetails from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetConfigDetails from JSON text
	FromJson(value string) error
	// Validate validates GetConfigDetails
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *getConfigDetails) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *getConfigDetails) setDefault() {

}

// ***** ExperimentRequest *****
type experimentRequest struct {
	obj *onexdataflowapi.ExperimentRequest
}

func NewExperimentRequest() ExperimentRequest {
	obj := experimentRequest{obj: &onexdataflowapi.ExperimentRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *experimentRequest) Msg() *onexdataflowapi.ExperimentRequest {
	return obj.obj
}

func (obj *experimentRequest) SetMsg(msg *onexdataflowapi.ExperimentRequest) ExperimentRequest {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *experimentRequest) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *experimentRequest) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *experimentRequest) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *experimentRequest) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *experimentRequest) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *experimentRequest) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *experimentRequest) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *experimentRequest) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *experimentRequest) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// ExperimentRequest is experiment request details
type ExperimentRequest interface {
	Msg() *onexdataflowapi.ExperimentRequest
	SetMsg(*onexdataflowapi.ExperimentRequest) ExperimentRequest
	// ToPbText marshals ExperimentRequest to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ExperimentRequest to YAML text
	ToYaml() (string, error)
	// ToJson marshals ExperimentRequest to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ExperimentRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ExperimentRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ExperimentRequest from JSON text
	FromJson(value string) error
	// Validate validates ExperimentRequest
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *experimentRequest) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *experimentRequest) setDefault() {

}

// ***** ControlStartRequest *****
type controlStartRequest struct {
	obj *onexdataflowapi.ControlStartRequest
}

func NewControlStartRequest() ControlStartRequest {
	obj := controlStartRequest{obj: &onexdataflowapi.ControlStartRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *controlStartRequest) Msg() *onexdataflowapi.ControlStartRequest {
	return obj.obj
}

func (obj *controlStartRequest) SetMsg(msg *onexdataflowapi.ControlStartRequest) ControlStartRequest {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *controlStartRequest) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *controlStartRequest) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *controlStartRequest) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStartRequest) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *controlStartRequest) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStartRequest) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *controlStartRequest) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *controlStartRequest) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *controlStartRequest) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// ControlStartRequest is start request details
type ControlStartRequest interface {
	Msg() *onexdataflowapi.ControlStartRequest
	SetMsg(*onexdataflowapi.ControlStartRequest) ControlStartRequest
	// ToPbText marshals ControlStartRequest to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ControlStartRequest to YAML text
	ToYaml() (string, error)
	// ToJson marshals ControlStartRequest to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ControlStartRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ControlStartRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ControlStartRequest from JSON text
	FromJson(value string) error
	// Validate validates ControlStartRequest
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *controlStartRequest) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *controlStartRequest) setDefault() {

}

// ***** ControlStatusRequest *****
type controlStatusRequest struct {
	obj *onexdataflowapi.ControlStatusRequest
}

func NewControlStatusRequest() ControlStatusRequest {
	obj := controlStatusRequest{obj: &onexdataflowapi.ControlStatusRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *controlStatusRequest) Msg() *onexdataflowapi.ControlStatusRequest {
	return obj.obj
}

func (obj *controlStatusRequest) SetMsg(msg *onexdataflowapi.ControlStatusRequest) ControlStatusRequest {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *controlStatusRequest) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *controlStatusRequest) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *controlStatusRequest) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStatusRequest) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *controlStatusRequest) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStatusRequest) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *controlStatusRequest) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *controlStatusRequest) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *controlStatusRequest) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// ControlStatusRequest is control.state request details
type ControlStatusRequest interface {
	Msg() *onexdataflowapi.ControlStatusRequest
	SetMsg(*onexdataflowapi.ControlStatusRequest) ControlStatusRequest
	// ToPbText marshals ControlStatusRequest to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ControlStatusRequest to YAML text
	ToYaml() (string, error)
	// ToJson marshals ControlStatusRequest to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ControlStatusRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ControlStatusRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ControlStatusRequest from JSON text
	FromJson(value string) error
	// Validate validates ControlStatusRequest
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *controlStatusRequest) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *controlStatusRequest) setDefault() {

}

// ***** MetricsRequest *****
type metricsRequest struct {
	obj *onexdataflowapi.MetricsRequest
}

func NewMetricsRequest() MetricsRequest {
	obj := metricsRequest{obj: &onexdataflowapi.MetricsRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsRequest) Msg() *onexdataflowapi.MetricsRequest {
	return obj.obj
}

func (obj *metricsRequest) SetMsg(msg *onexdataflowapi.MetricsRequest) MetricsRequest {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsRequest) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *metricsRequest) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsRequest) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsRequest) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *metricsRequest) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsRequest) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *metricsRequest) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *metricsRequest) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *metricsRequest) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// MetricsRequest is metrics request details
type MetricsRequest interface {
	Msg() *onexdataflowapi.MetricsRequest
	SetMsg(*onexdataflowapi.MetricsRequest) MetricsRequest
	// ToPbText marshals MetricsRequest to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals MetricsRequest to YAML text
	ToYaml() (string, error)
	// ToJson marshals MetricsRequest to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals MetricsRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsRequest from JSON text
	FromJson(value string) error
	// Validate validates MetricsRequest
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *metricsRequest) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *metricsRequest) setDefault() {

}

// ***** SetConfigResponse *****
type setConfigResponse struct {
	obj                  *onexdataflowapi.SetConfigResponse
	statusCode_200Holder Config
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewSetConfigResponse() SetConfigResponse {
	obj := setConfigResponse{obj: &onexdataflowapi.SetConfigResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *setConfigResponse) Msg() *onexdataflowapi.SetConfigResponse {
	return obj.obj
}

func (obj *setConfigResponse) SetMsg(msg *onexdataflowapi.SetConfigResponse) SetConfigResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *setConfigResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *setConfigResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *setConfigResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *setConfigResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *setConfigResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *setConfigResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *setConfigResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *setConfigResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *setConfigResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *setConfigResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// SetConfigResponse is description is TBD
type SetConfigResponse interface {
	Msg() *onexdataflowapi.SetConfigResponse
	SetMsg(*onexdataflowapi.SetConfigResponse) SetConfigResponse
	// ToPbText marshals SetConfigResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals SetConfigResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals SetConfigResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals SetConfigResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals SetConfigResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals SetConfigResponse from JSON text
	FromJson(value string) error
	// Validate validates SetConfigResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode200 returns Config, set in SetConfigResponse.
	// Config is oNEx dataflow configuration
	StatusCode200() Config
	// SetStatusCode200 assigns Config provided by user to SetConfigResponse.
	// Config is oNEx dataflow configuration
	SetStatusCode200(value Config) SetConfigResponse
	// HasStatusCode200 checks if StatusCode200 has been set in SetConfigResponse
	HasStatusCode200() bool
	// StatusCode400 returns ErrorDetails, set in SetConfigResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to SetConfigResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) SetConfigResponse
	// HasStatusCode400 checks if StatusCode400 has been set in SetConfigResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in SetConfigResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to SetConfigResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) SetConfigResponse
	// HasStatusCode500 checks if StatusCode500 has been set in SetConfigResponse
	HasStatusCode500() bool
	setNil()
}

// StatusCode200 returns a Config
// description is TBD
func (obj *setConfigResponse) StatusCode200() Config {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewConfig().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &config{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a Config
// description is TBD
func (obj *setConfigResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the Config value in the SetConfigResponse object
// description is TBD
func (obj *setConfigResponse) SetStatusCode200(value Config) SetConfigResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *setConfigResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *setConfigResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the SetConfigResponse object
// description is TBD
func (obj *setConfigResponse) SetStatusCode400(value ErrorDetails) SetConfigResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *setConfigResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *setConfigResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the SetConfigResponse object
// description is TBD
func (obj *setConfigResponse) SetStatusCode500(value ErrorDetails) SetConfigResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

func (obj *setConfigResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

}

func (obj *setConfigResponse) setDefault() {

}

// ***** GetConfigResponse *****
type getConfigResponse struct {
	obj                  *onexdataflowapi.GetConfigResponse
	statusCode_200Holder Config
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewGetConfigResponse() GetConfigResponse {
	obj := getConfigResponse{obj: &onexdataflowapi.GetConfigResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *getConfigResponse) Msg() *onexdataflowapi.GetConfigResponse {
	return obj.obj
}

func (obj *getConfigResponse) SetMsg(msg *onexdataflowapi.GetConfigResponse) GetConfigResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getConfigResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *getConfigResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *getConfigResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getConfigResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *getConfigResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getConfigResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *getConfigResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *getConfigResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *getConfigResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *getConfigResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// GetConfigResponse is description is TBD
type GetConfigResponse interface {
	Msg() *onexdataflowapi.GetConfigResponse
	SetMsg(*onexdataflowapi.GetConfigResponse) GetConfigResponse
	// ToPbText marshals GetConfigResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals GetConfigResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals GetConfigResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals GetConfigResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetConfigResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetConfigResponse from JSON text
	FromJson(value string) error
	// Validate validates GetConfigResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode200 returns Config, set in GetConfigResponse.
	// Config is oNEx dataflow configuration
	StatusCode200() Config
	// SetStatusCode200 assigns Config provided by user to GetConfigResponse.
	// Config is oNEx dataflow configuration
	SetStatusCode200(value Config) GetConfigResponse
	// HasStatusCode200 checks if StatusCode200 has been set in GetConfigResponse
	HasStatusCode200() bool
	// StatusCode400 returns ErrorDetails, set in GetConfigResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to GetConfigResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) GetConfigResponse
	// HasStatusCode400 checks if StatusCode400 has been set in GetConfigResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in GetConfigResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to GetConfigResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) GetConfigResponse
	// HasStatusCode500 checks if StatusCode500 has been set in GetConfigResponse
	HasStatusCode500() bool
	setNil()
}

// StatusCode200 returns a Config
// description is TBD
func (obj *getConfigResponse) StatusCode200() Config {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewConfig().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &config{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a Config
// description is TBD
func (obj *getConfigResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the Config value in the GetConfigResponse object
// description is TBD
func (obj *getConfigResponse) SetStatusCode200(value Config) GetConfigResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getConfigResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getConfigResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the GetConfigResponse object
// description is TBD
func (obj *getConfigResponse) SetStatusCode400(value ErrorDetails) GetConfigResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getConfigResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getConfigResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the GetConfigResponse object
// description is TBD
func (obj *getConfigResponse) SetStatusCode500(value ErrorDetails) GetConfigResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

func (obj *getConfigResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

}

func (obj *getConfigResponse) setDefault() {

}

// ***** RunExperimentResponse *****
type runExperimentResponse struct {
	obj                  *onexdataflowapi.RunExperimentResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
	statusCode_200Holder WarningDetails
}

func NewRunExperimentResponse() RunExperimentResponse {
	obj := runExperimentResponse{obj: &onexdataflowapi.RunExperimentResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *runExperimentResponse) Msg() *onexdataflowapi.RunExperimentResponse {
	return obj.obj
}

func (obj *runExperimentResponse) SetMsg(msg *onexdataflowapi.RunExperimentResponse) RunExperimentResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *runExperimentResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *runExperimentResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *runExperimentResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *runExperimentResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *runExperimentResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *runExperimentResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *runExperimentResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *runExperimentResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *runExperimentResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *runExperimentResponse) setNil() {
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
	obj.statusCode_200Holder = nil
}

// RunExperimentResponse is description is TBD
type RunExperimentResponse interface {
	Msg() *onexdataflowapi.RunExperimentResponse
	SetMsg(*onexdataflowapi.RunExperimentResponse) RunExperimentResponse
	// ToPbText marshals RunExperimentResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals RunExperimentResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals RunExperimentResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals RunExperimentResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals RunExperimentResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals RunExperimentResponse from JSON text
	FromJson(value string) error
	// Validate validates RunExperimentResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode400 returns ErrorDetails, set in RunExperimentResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to RunExperimentResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) RunExperimentResponse
	// HasStatusCode400 checks if StatusCode400 has been set in RunExperimentResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in RunExperimentResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to RunExperimentResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) RunExperimentResponse
	// HasStatusCode500 checks if StatusCode500 has been set in RunExperimentResponse
	HasStatusCode500() bool
	// StatusCode200 returns WarningDetails, set in RunExperimentResponse.
	// WarningDetails is description is TBD
	StatusCode200() WarningDetails
	// SetStatusCode200 assigns WarningDetails provided by user to RunExperimentResponse.
	// WarningDetails is description is TBD
	SetStatusCode200(value WarningDetails) RunExperimentResponse
	// HasStatusCode200 checks if StatusCode200 has been set in RunExperimentResponse
	HasStatusCode200() bool
	setNil()
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *runExperimentResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *runExperimentResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the RunExperimentResponse object
// description is TBD
func (obj *runExperimentResponse) SetStatusCode400(value ErrorDetails) RunExperimentResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *runExperimentResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *runExperimentResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the RunExperimentResponse object
// description is TBD
func (obj *runExperimentResponse) SetStatusCode500(value ErrorDetails) RunExperimentResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

// StatusCode200 returns a WarningDetails
// description is TBD
func (obj *runExperimentResponse) StatusCode200() WarningDetails {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewWarningDetails().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &warningDetails{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a WarningDetails
// description is TBD
func (obj *runExperimentResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the WarningDetails value in the RunExperimentResponse object
// description is TBD
func (obj *runExperimentResponse) SetStatusCode200(value WarningDetails) RunExperimentResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

func (obj *runExperimentResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

}

func (obj *runExperimentResponse) setDefault() {

}

// ***** StartResponse *****
type startResponse struct {
	obj                  *onexdataflowapi.StartResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
	statusCode_200Holder WarningDetails
}

func NewStartResponse() StartResponse {
	obj := startResponse{obj: &onexdataflowapi.StartResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *startResponse) Msg() *onexdataflowapi.StartResponse {
	return obj.obj
}

func (obj *startResponse) SetMsg(msg *onexdataflowapi.StartResponse) StartResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *startResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *startResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *startResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *startResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *startResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *startResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *startResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *startResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *startResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *startResponse) setNil() {
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
	obj.statusCode_200Holder = nil
}

// StartResponse is description is TBD
type StartResponse interface {
	Msg() *onexdataflowapi.StartResponse
	SetMsg(*onexdataflowapi.StartResponse) StartResponse
	// ToPbText marshals StartResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals StartResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals StartResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals StartResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals StartResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals StartResponse from JSON text
	FromJson(value string) error
	// Validate validates StartResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode400 returns ErrorDetails, set in StartResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to StartResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) StartResponse
	// HasStatusCode400 checks if StatusCode400 has been set in StartResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in StartResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to StartResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) StartResponse
	// HasStatusCode500 checks if StatusCode500 has been set in StartResponse
	HasStatusCode500() bool
	// StatusCode200 returns WarningDetails, set in StartResponse.
	// WarningDetails is description is TBD
	StatusCode200() WarningDetails
	// SetStatusCode200 assigns WarningDetails provided by user to StartResponse.
	// WarningDetails is description is TBD
	SetStatusCode200(value WarningDetails) StartResponse
	// HasStatusCode200 checks if StatusCode200 has been set in StartResponse
	HasStatusCode200() bool
	setNil()
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *startResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *startResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the StartResponse object
// description is TBD
func (obj *startResponse) SetStatusCode400(value ErrorDetails) StartResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *startResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *startResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the StartResponse object
// description is TBD
func (obj *startResponse) SetStatusCode500(value ErrorDetails) StartResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

// StatusCode200 returns a WarningDetails
// description is TBD
func (obj *startResponse) StatusCode200() WarningDetails {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewWarningDetails().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &warningDetails{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a WarningDetails
// description is TBD
func (obj *startResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the WarningDetails value in the StartResponse object
// description is TBD
func (obj *startResponse) SetStatusCode200(value WarningDetails) StartResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

func (obj *startResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

}

func (obj *startResponse) setDefault() {

}

// ***** GetStatusResponse *****
type getStatusResponse struct {
	obj                  *onexdataflowapi.GetStatusResponse
	statusCode_200Holder ControlStatusResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewGetStatusResponse() GetStatusResponse {
	obj := getStatusResponse{obj: &onexdataflowapi.GetStatusResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *getStatusResponse) Msg() *onexdataflowapi.GetStatusResponse {
	return obj.obj
}

func (obj *getStatusResponse) SetMsg(msg *onexdataflowapi.GetStatusResponse) GetStatusResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getStatusResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *getStatusResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *getStatusResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getStatusResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *getStatusResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getStatusResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *getStatusResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *getStatusResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *getStatusResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *getStatusResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// GetStatusResponse is description is TBD
type GetStatusResponse interface {
	Msg() *onexdataflowapi.GetStatusResponse
	SetMsg(*onexdataflowapi.GetStatusResponse) GetStatusResponse
	// ToPbText marshals GetStatusResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals GetStatusResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals GetStatusResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals GetStatusResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetStatusResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetStatusResponse from JSON text
	FromJson(value string) error
	// Validate validates GetStatusResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode200 returns ControlStatusResponse, set in GetStatusResponse.
	// ControlStatusResponse is control/state response details
	StatusCode200() ControlStatusResponse
	// SetStatusCode200 assigns ControlStatusResponse provided by user to GetStatusResponse.
	// ControlStatusResponse is control/state response details
	SetStatusCode200(value ControlStatusResponse) GetStatusResponse
	// HasStatusCode200 checks if StatusCode200 has been set in GetStatusResponse
	HasStatusCode200() bool
	// StatusCode400 returns ErrorDetails, set in GetStatusResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to GetStatusResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) GetStatusResponse
	// HasStatusCode400 checks if StatusCode400 has been set in GetStatusResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in GetStatusResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to GetStatusResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) GetStatusResponse
	// HasStatusCode500 checks if StatusCode500 has been set in GetStatusResponse
	HasStatusCode500() bool
	setNil()
}

// StatusCode200 returns a ControlStatusResponse
// description is TBD
func (obj *getStatusResponse) StatusCode200() ControlStatusResponse {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewControlStatusResponse().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &controlStatusResponse{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a ControlStatusResponse
// description is TBD
func (obj *getStatusResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the ControlStatusResponse value in the GetStatusResponse object
// description is TBD
func (obj *getStatusResponse) SetStatusCode200(value ControlStatusResponse) GetStatusResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getStatusResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getStatusResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the GetStatusResponse object
// description is TBD
func (obj *getStatusResponse) SetStatusCode400(value ErrorDetails) GetStatusResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getStatusResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getStatusResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the GetStatusResponse object
// description is TBD
func (obj *getStatusResponse) SetStatusCode500(value ErrorDetails) GetStatusResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

func (obj *getStatusResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

}

func (obj *getStatusResponse) setDefault() {

}

// ***** GetMetricsResponse *****
type getMetricsResponse struct {
	obj                  *onexdataflowapi.GetMetricsResponse
	statusCode_200Holder MetricsResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewGetMetricsResponse() GetMetricsResponse {
	obj := getMetricsResponse{obj: &onexdataflowapi.GetMetricsResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *getMetricsResponse) Msg() *onexdataflowapi.GetMetricsResponse {
	return obj.obj
}

func (obj *getMetricsResponse) SetMsg(msg *onexdataflowapi.GetMetricsResponse) GetMetricsResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getMetricsResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *getMetricsResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *getMetricsResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getMetricsResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *getMetricsResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *getMetricsResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *getMetricsResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *getMetricsResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *getMetricsResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *getMetricsResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// GetMetricsResponse is description is TBD
type GetMetricsResponse interface {
	Msg() *onexdataflowapi.GetMetricsResponse
	SetMsg(*onexdataflowapi.GetMetricsResponse) GetMetricsResponse
	// ToPbText marshals GetMetricsResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals GetMetricsResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals GetMetricsResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals GetMetricsResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetMetricsResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetMetricsResponse from JSON text
	FromJson(value string) error
	// Validate validates GetMetricsResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode200 returns MetricsResponse, set in GetMetricsResponse.
	// MetricsResponse is metrics response details
	StatusCode200() MetricsResponse
	// SetStatusCode200 assigns MetricsResponse provided by user to GetMetricsResponse.
	// MetricsResponse is metrics response details
	SetStatusCode200(value MetricsResponse) GetMetricsResponse
	// HasStatusCode200 checks if StatusCode200 has been set in GetMetricsResponse
	HasStatusCode200() bool
	// StatusCode400 returns ErrorDetails, set in GetMetricsResponse.
	// ErrorDetails is description is TBD
	StatusCode400() ErrorDetails
	// SetStatusCode400 assigns ErrorDetails provided by user to GetMetricsResponse.
	// ErrorDetails is description is TBD
	SetStatusCode400(value ErrorDetails) GetMetricsResponse
	// HasStatusCode400 checks if StatusCode400 has been set in GetMetricsResponse
	HasStatusCode400() bool
	// StatusCode500 returns ErrorDetails, set in GetMetricsResponse.
	// ErrorDetails is description is TBD
	StatusCode500() ErrorDetails
	// SetStatusCode500 assigns ErrorDetails provided by user to GetMetricsResponse.
	// ErrorDetails is description is TBD
	SetStatusCode500(value ErrorDetails) GetMetricsResponse
	// HasStatusCode500 checks if StatusCode500 has been set in GetMetricsResponse
	HasStatusCode500() bool
	setNil()
}

// StatusCode200 returns a MetricsResponse
// description is TBD
func (obj *getMetricsResponse) StatusCode200() MetricsResponse {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewMetricsResponse().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &metricsResponse{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a MetricsResponse
// description is TBD
func (obj *getMetricsResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the MetricsResponse value in the GetMetricsResponse object
// description is TBD
func (obj *getMetricsResponse) SetStatusCode200(value MetricsResponse) GetMetricsResponse {

	obj.statusCode_200Holder = nil
	obj.obj.StatusCode_200 = value.Msg()

	return obj
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getMetricsResponse) StatusCode400() ErrorDetails {
	if obj.obj.StatusCode_400 == nil {
		obj.obj.StatusCode_400 = NewErrorDetails().Msg()
	}
	if obj.statusCode_400Holder == nil {
		obj.statusCode_400Holder = &errorDetails{obj: obj.obj.StatusCode_400}
	}
	return obj.statusCode_400Holder
}

// StatusCode400 returns a ErrorDetails
// description is TBD
func (obj *getMetricsResponse) HasStatusCode400() bool {
	return obj.obj.StatusCode_400 != nil
}

// SetStatusCode400 sets the ErrorDetails value in the GetMetricsResponse object
// description is TBD
func (obj *getMetricsResponse) SetStatusCode400(value ErrorDetails) GetMetricsResponse {

	obj.statusCode_400Holder = nil
	obj.obj.StatusCode_400 = value.Msg()

	return obj
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getMetricsResponse) StatusCode500() ErrorDetails {
	if obj.obj.StatusCode_500 == nil {
		obj.obj.StatusCode_500 = NewErrorDetails().Msg()
	}
	if obj.statusCode_500Holder == nil {
		obj.statusCode_500Holder = &errorDetails{obj: obj.obj.StatusCode_500}
	}
	return obj.statusCode_500Holder
}

// StatusCode500 returns a ErrorDetails
// description is TBD
func (obj *getMetricsResponse) HasStatusCode500() bool {
	return obj.obj.StatusCode_500 != nil
}

// SetStatusCode500 sets the ErrorDetails value in the GetMetricsResponse object
// description is TBD
func (obj *getMetricsResponse) SetStatusCode500(value ErrorDetails) GetMetricsResponse {

	obj.statusCode_500Holder = nil
	obj.obj.StatusCode_500 = value.Msg()

	return obj
}

func (obj *getMetricsResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.StatusCode_200 != nil {
		obj.StatusCode200().validateObj(set_default)
	}

	if obj.obj.StatusCode_400 != nil {
		obj.StatusCode400().validateObj(set_default)
	}

	if obj.obj.StatusCode_500 != nil {
		obj.StatusCode500().validateObj(set_default)
	}

}

func (obj *getMetricsResponse) setDefault() {

}

// ***** Host *****
type host struct {
	obj *onexdataflowapi.Host
}

func NewHost() Host {
	obj := host{obj: &onexdataflowapi.Host{}}
	obj.setDefault()
	return &obj
}

func (obj *host) Msg() *onexdataflowapi.Host {
	return obj.obj
}

func (obj *host) SetMsg(msg *onexdataflowapi.Host) Host {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *host) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *host) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *host) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *host) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *host) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *host) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *host) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *host) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *host) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// Host is description is TBD
type Host interface {
	Msg() *onexdataflowapi.Host
	SetMsg(*onexdataflowapi.Host) Host
	// ToPbText marshals Host to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals Host to YAML text
	ToYaml() (string, error)
	// ToJson marshals Host to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals Host from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Host from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Host from JSON text
	FromJson(value string) error
	// Validate validates Host
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in Host.
	Name() string
	// SetName assigns string provided by user to Host
	SetName(value string) Host
	// Address returns string, set in Host.
	Address() string
	// SetAddress assigns string provided by user to Host
	SetAddress(value string) Host
	// Prefix returns int32, set in Host.
	Prefix() int32
	// SetPrefix assigns int32 provided by user to Host
	SetPrefix(value int32) Host
	// HasPrefix checks if Prefix has been set in Host
	HasPrefix() bool
	// L1ProfileName returns string, set in Host.
	L1ProfileName() string
	// SetL1ProfileName assigns string provided by user to Host
	SetL1ProfileName(value string) Host
	// HasL1ProfileName checks if L1ProfileName has been set in Host
	HasL1ProfileName() bool
	// Annotations returns string, set in Host.
	Annotations() string
	// SetAnnotations assigns string provided by user to Host
	SetAnnotations(value string) Host
	// HasAnnotations checks if Annotations has been set in Host
	HasAnnotations() bool
}

// Name returns a string
// The name, uniquely identifying the host
func (obj *host) Name() string {

	return obj.obj.Name
}

// SetName sets the string value in the Host object
// The name, uniquely identifying the host
func (obj *host) SetName(value string) Host {

	obj.obj.Name = value
	return obj
}

// Address returns a string
// The test address of the host
func (obj *host) Address() string {

	return obj.obj.Address
}

// SetAddress sets the string value in the Host object
// The test address of the host
func (obj *host) SetAddress(value string) Host {

	obj.obj.Address = value
	return obj
}

// Prefix returns a int32
// The prefix of the host
func (obj *host) Prefix() int32 {

	return *obj.obj.Prefix

}

// Prefix returns a int32
// The prefix of the host
func (obj *host) HasPrefix() bool {
	return obj.obj.Prefix != nil
}

// SetPrefix sets the int32 value in the Host object
// The prefix of the host
func (obj *host) SetPrefix(value int32) Host {

	obj.obj.Prefix = &value
	return obj
}

// L1ProfileName returns a string
// The layer 1 settings profile associated with the host/front panel port.
//
// x-constraint:
// - ../l1settings/l1_profiles.yaml#/components/schemas/L1SettingsProfile/properties/name
//
func (obj *host) L1ProfileName() string {

	return *obj.obj.L1ProfileName

}

// L1ProfileName returns a string
// The layer 1 settings profile associated with the host/front panel port.
//
// x-constraint:
// - ../l1settings/l1_profiles.yaml#/components/schemas/L1SettingsProfile/properties/name
//
func (obj *host) HasL1ProfileName() bool {
	return obj.obj.L1ProfileName != nil
}

// SetL1ProfileName sets the string value in the Host object
// The layer 1 settings profile associated with the host/front panel port.
//
// x-constraint:
// - ../l1settings/l1_profiles.yaml#/components/schemas/L1SettingsProfile/properties/name
//
func (obj *host) SetL1ProfileName(value string) Host {

	obj.obj.L1ProfileName = &value
	return obj
}

// Annotations returns a string
// description is TBD
func (obj *host) Annotations() string {

	return *obj.obj.Annotations

}

// Annotations returns a string
// description is TBD
func (obj *host) HasAnnotations() bool {
	return obj.obj.Annotations != nil
}

// SetAnnotations sets the string value in the Host object
// description is TBD
func (obj *host) SetAnnotations(value string) Host {

	obj.obj.Annotations = &value
	return obj
}

func (obj *host) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// Name is required
	if obj.obj.Name == "" {
		validation = append(validation, "Name is required field on interface Host")
	}

	// Address is required
	if obj.obj.Address == "" {
		validation = append(validation, "Address is required field on interface Host")
	}
}

func (obj *host) setDefault() {
	if obj.obj.Prefix == nil {
		obj.SetPrefix(24)
	}

}

// ***** Dataflow *****
type dataflow struct {
	obj                  *onexdataflowapi.Dataflow
	hostManagementHolder DataflowDataflowHostManagementIter
	workloadHolder       DataflowDataflowWorkloadItemIter
	flowProfilesHolder   DataflowDataflowFlowProfileIter
}

func NewDataflow() Dataflow {
	obj := dataflow{obj: &onexdataflowapi.Dataflow{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflow) Msg() *onexdataflowapi.Dataflow {
	return obj.obj
}

func (obj *dataflow) SetMsg(msg *onexdataflowapi.Dataflow) Dataflow {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflow) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflow) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflow) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflow) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflow) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflow) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflow) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflow) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflow) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflow) setNil() {
	obj.hostManagementHolder = nil
	obj.workloadHolder = nil
	obj.flowProfilesHolder = nil
}

// Dataflow is description is TBD
type Dataflow interface {
	Msg() *onexdataflowapi.Dataflow
	SetMsg(*onexdataflowapi.Dataflow) Dataflow
	// ToPbText marshals Dataflow to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals Dataflow to YAML text
	ToYaml() (string, error)
	// ToJson marshals Dataflow to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals Dataflow from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Dataflow from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Dataflow from JSON text
	FromJson(value string) error
	// Validate validates Dataflow
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// HostManagement returns DataflowDataflowHostManagementIter, set in Dataflow
	HostManagement() DataflowDataflowHostManagementIter
	// Workload returns DataflowDataflowWorkloadItemIter, set in Dataflow
	Workload() DataflowDataflowWorkloadItemIter
	// FlowProfiles returns DataflowDataflowFlowProfileIter, set in Dataflow
	FlowProfiles() DataflowDataflowFlowProfileIter
	setNil()
}

// HostManagement returns a []DataflowHostManagement
// description is TBD
func (obj *dataflow) HostManagement() DataflowDataflowHostManagementIter {
	if len(obj.obj.HostManagement) == 0 {
		obj.obj.HostManagement = []*onexdataflowapi.DataflowHostManagement{}
	}
	if obj.hostManagementHolder == nil {
		obj.hostManagementHolder = newDataflowDataflowHostManagementIter().setMsg(obj)
	}
	return obj.hostManagementHolder
}

type dataflowDataflowHostManagementIter struct {
	obj                         *dataflow
	dataflowHostManagementSlice []DataflowHostManagement
}

func newDataflowDataflowHostManagementIter() DataflowDataflowHostManagementIter {
	return &dataflowDataflowHostManagementIter{}
}

type DataflowDataflowHostManagementIter interface {
	setMsg(*dataflow) DataflowDataflowHostManagementIter
	Items() []DataflowHostManagement
	Add() DataflowHostManagement
	Append(items ...DataflowHostManagement) DataflowDataflowHostManagementIter
	Set(index int, newObj DataflowHostManagement) DataflowDataflowHostManagementIter
	Clear() DataflowDataflowHostManagementIter
	clearHolderSlice() DataflowDataflowHostManagementIter
	appendHolderSlice(item DataflowHostManagement) DataflowDataflowHostManagementIter
}

func (obj *dataflowDataflowHostManagementIter) setMsg(msg *dataflow) DataflowDataflowHostManagementIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.HostManagement {
		obj.appendHolderSlice(&dataflowHostManagement{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *dataflowDataflowHostManagementIter) Items() []DataflowHostManagement {
	return obj.dataflowHostManagementSlice
}

func (obj *dataflowDataflowHostManagementIter) Add() DataflowHostManagement {
	newObj := &onexdataflowapi.DataflowHostManagement{}
	obj.obj.obj.HostManagement = append(obj.obj.obj.HostManagement, newObj)
	newLibObj := &dataflowHostManagement{obj: newObj}
	newLibObj.setDefault()
	obj.dataflowHostManagementSlice = append(obj.dataflowHostManagementSlice, newLibObj)
	return newLibObj
}

func (obj *dataflowDataflowHostManagementIter) Append(items ...DataflowHostManagement) DataflowDataflowHostManagementIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.HostManagement = append(obj.obj.obj.HostManagement, newObj)
		obj.dataflowHostManagementSlice = append(obj.dataflowHostManagementSlice, item)
	}
	return obj
}

func (obj *dataflowDataflowHostManagementIter) Set(index int, newObj DataflowHostManagement) DataflowDataflowHostManagementIter {
	obj.obj.obj.HostManagement[index] = newObj.Msg()
	obj.dataflowHostManagementSlice[index] = newObj
	return obj
}
func (obj *dataflowDataflowHostManagementIter) Clear() DataflowDataflowHostManagementIter {
	if len(obj.obj.obj.HostManagement) > 0 {
		obj.obj.obj.HostManagement = []*onexdataflowapi.DataflowHostManagement{}
		obj.dataflowHostManagementSlice = []DataflowHostManagement{}
	}
	return obj
}
func (obj *dataflowDataflowHostManagementIter) clearHolderSlice() DataflowDataflowHostManagementIter {
	if len(obj.dataflowHostManagementSlice) > 0 {
		obj.dataflowHostManagementSlice = []DataflowHostManagement{}
	}
	return obj
}
func (obj *dataflowDataflowHostManagementIter) appendHolderSlice(item DataflowHostManagement) DataflowDataflowHostManagementIter {
	obj.dataflowHostManagementSlice = append(obj.dataflowHostManagementSlice, item)
	return obj
}

// Workload returns a []DataflowWorkloadItem
// The workload items making up the dataflow
func (obj *dataflow) Workload() DataflowDataflowWorkloadItemIter {
	if len(obj.obj.Workload) == 0 {
		obj.obj.Workload = []*onexdataflowapi.DataflowWorkloadItem{}
	}
	if obj.workloadHolder == nil {
		obj.workloadHolder = newDataflowDataflowWorkloadItemIter().setMsg(obj)
	}
	return obj.workloadHolder
}

type dataflowDataflowWorkloadItemIter struct {
	obj                       *dataflow
	dataflowWorkloadItemSlice []DataflowWorkloadItem
}

func newDataflowDataflowWorkloadItemIter() DataflowDataflowWorkloadItemIter {
	return &dataflowDataflowWorkloadItemIter{}
}

type DataflowDataflowWorkloadItemIter interface {
	setMsg(*dataflow) DataflowDataflowWorkloadItemIter
	Items() []DataflowWorkloadItem
	Add() DataflowWorkloadItem
	Append(items ...DataflowWorkloadItem) DataflowDataflowWorkloadItemIter
	Set(index int, newObj DataflowWorkloadItem) DataflowDataflowWorkloadItemIter
	Clear() DataflowDataflowWorkloadItemIter
	clearHolderSlice() DataflowDataflowWorkloadItemIter
	appendHolderSlice(item DataflowWorkloadItem) DataflowDataflowWorkloadItemIter
}

func (obj *dataflowDataflowWorkloadItemIter) setMsg(msg *dataflow) DataflowDataflowWorkloadItemIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Workload {
		obj.appendHolderSlice(&dataflowWorkloadItem{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *dataflowDataflowWorkloadItemIter) Items() []DataflowWorkloadItem {
	return obj.dataflowWorkloadItemSlice
}

func (obj *dataflowDataflowWorkloadItemIter) Add() DataflowWorkloadItem {
	newObj := &onexdataflowapi.DataflowWorkloadItem{}
	obj.obj.obj.Workload = append(obj.obj.obj.Workload, newObj)
	newLibObj := &dataflowWorkloadItem{obj: newObj}
	newLibObj.setDefault()
	obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, newLibObj)
	return newLibObj
}

func (obj *dataflowDataflowWorkloadItemIter) Append(items ...DataflowWorkloadItem) DataflowDataflowWorkloadItemIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Workload = append(obj.obj.obj.Workload, newObj)
		obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, item)
	}
	return obj
}

func (obj *dataflowDataflowWorkloadItemIter) Set(index int, newObj DataflowWorkloadItem) DataflowDataflowWorkloadItemIter {
	obj.obj.obj.Workload[index] = newObj.Msg()
	obj.dataflowWorkloadItemSlice[index] = newObj
	return obj
}
func (obj *dataflowDataflowWorkloadItemIter) Clear() DataflowDataflowWorkloadItemIter {
	if len(obj.obj.obj.Workload) > 0 {
		obj.obj.obj.Workload = []*onexdataflowapi.DataflowWorkloadItem{}
		obj.dataflowWorkloadItemSlice = []DataflowWorkloadItem{}
	}
	return obj
}
func (obj *dataflowDataflowWorkloadItemIter) clearHolderSlice() DataflowDataflowWorkloadItemIter {
	if len(obj.dataflowWorkloadItemSlice) > 0 {
		obj.dataflowWorkloadItemSlice = []DataflowWorkloadItem{}
	}
	return obj
}
func (obj *dataflowDataflowWorkloadItemIter) appendHolderSlice(item DataflowWorkloadItem) DataflowDataflowWorkloadItemIter {
	obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, item)
	return obj
}

// FlowProfiles returns a []DataflowFlowProfile
// foo
func (obj *dataflow) FlowProfiles() DataflowDataflowFlowProfileIter {
	if len(obj.obj.FlowProfiles) == 0 {
		obj.obj.FlowProfiles = []*onexdataflowapi.DataflowFlowProfile{}
	}
	if obj.flowProfilesHolder == nil {
		obj.flowProfilesHolder = newDataflowDataflowFlowProfileIter().setMsg(obj)
	}
	return obj.flowProfilesHolder
}

type dataflowDataflowFlowProfileIter struct {
	obj                      *dataflow
	dataflowFlowProfileSlice []DataflowFlowProfile
}

func newDataflowDataflowFlowProfileIter() DataflowDataflowFlowProfileIter {
	return &dataflowDataflowFlowProfileIter{}
}

type DataflowDataflowFlowProfileIter interface {
	setMsg(*dataflow) DataflowDataflowFlowProfileIter
	Items() []DataflowFlowProfile
	Add() DataflowFlowProfile
	Append(items ...DataflowFlowProfile) DataflowDataflowFlowProfileIter
	Set(index int, newObj DataflowFlowProfile) DataflowDataflowFlowProfileIter
	Clear() DataflowDataflowFlowProfileIter
	clearHolderSlice() DataflowDataflowFlowProfileIter
	appendHolderSlice(item DataflowFlowProfile) DataflowDataflowFlowProfileIter
}

func (obj *dataflowDataflowFlowProfileIter) setMsg(msg *dataflow) DataflowDataflowFlowProfileIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.FlowProfiles {
		obj.appendHolderSlice(&dataflowFlowProfile{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *dataflowDataflowFlowProfileIter) Items() []DataflowFlowProfile {
	return obj.dataflowFlowProfileSlice
}

func (obj *dataflowDataflowFlowProfileIter) Add() DataflowFlowProfile {
	newObj := &onexdataflowapi.DataflowFlowProfile{}
	obj.obj.obj.FlowProfiles = append(obj.obj.obj.FlowProfiles, newObj)
	newLibObj := &dataflowFlowProfile{obj: newObj}
	newLibObj.setDefault()
	obj.dataflowFlowProfileSlice = append(obj.dataflowFlowProfileSlice, newLibObj)
	return newLibObj
}

func (obj *dataflowDataflowFlowProfileIter) Append(items ...DataflowFlowProfile) DataflowDataflowFlowProfileIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.FlowProfiles = append(obj.obj.obj.FlowProfiles, newObj)
		obj.dataflowFlowProfileSlice = append(obj.dataflowFlowProfileSlice, item)
	}
	return obj
}

func (obj *dataflowDataflowFlowProfileIter) Set(index int, newObj DataflowFlowProfile) DataflowDataflowFlowProfileIter {
	obj.obj.obj.FlowProfiles[index] = newObj.Msg()
	obj.dataflowFlowProfileSlice[index] = newObj
	return obj
}
func (obj *dataflowDataflowFlowProfileIter) Clear() DataflowDataflowFlowProfileIter {
	if len(obj.obj.obj.FlowProfiles) > 0 {
		obj.obj.obj.FlowProfiles = []*onexdataflowapi.DataflowFlowProfile{}
		obj.dataflowFlowProfileSlice = []DataflowFlowProfile{}
	}
	return obj
}
func (obj *dataflowDataflowFlowProfileIter) clearHolderSlice() DataflowDataflowFlowProfileIter {
	if len(obj.dataflowFlowProfileSlice) > 0 {
		obj.dataflowFlowProfileSlice = []DataflowFlowProfile{}
	}
	return obj
}
func (obj *dataflowDataflowFlowProfileIter) appendHolderSlice(item DataflowFlowProfile) DataflowDataflowFlowProfileIter {
	obj.dataflowFlowProfileSlice = append(obj.dataflowFlowProfileSlice, item)
	return obj
}

func (obj *dataflow) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.HostManagement != nil {

		if set_default {
			obj.HostManagement().clearHolderSlice()
			for _, item := range obj.obj.HostManagement {
				obj.HostManagement().appendHolderSlice(&dataflowHostManagement{obj: item})
			}
		}
		for _, item := range obj.HostManagement().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.Workload != nil {

		if set_default {
			obj.Workload().clearHolderSlice()
			for _, item := range obj.obj.Workload {
				obj.Workload().appendHolderSlice(&dataflowWorkloadItem{obj: item})
			}
		}
		for _, item := range obj.Workload().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.FlowProfiles != nil {

		if set_default {
			obj.FlowProfiles().clearHolderSlice()
			for _, item := range obj.obj.FlowProfiles {
				obj.FlowProfiles().appendHolderSlice(&dataflowFlowProfile{obj: item})
			}
		}
		for _, item := range obj.FlowProfiles().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *dataflow) setDefault() {

}

// ***** ErrorDetails *****
type errorDetails struct {
	obj          *onexdataflowapi.ErrorDetails
	errorsHolder ErrorDetailsErrorItemIter
}

func NewErrorDetails() ErrorDetails {
	obj := errorDetails{obj: &onexdataflowapi.ErrorDetails{}}
	obj.setDefault()
	return &obj
}

func (obj *errorDetails) Msg() *onexdataflowapi.ErrorDetails {
	return obj.obj
}

func (obj *errorDetails) SetMsg(msg *onexdataflowapi.ErrorDetails) ErrorDetails {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *errorDetails) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *errorDetails) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *errorDetails) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *errorDetails) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *errorDetails) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *errorDetails) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *errorDetails) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *errorDetails) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *errorDetails) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *errorDetails) setNil() {
	obj.errorsHolder = nil
}

// ErrorDetails is description is TBD
type ErrorDetails interface {
	Msg() *onexdataflowapi.ErrorDetails
	SetMsg(*onexdataflowapi.ErrorDetails) ErrorDetails
	// ToPbText marshals ErrorDetails to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ErrorDetails to YAML text
	ToYaml() (string, error)
	// ToJson marshals ErrorDetails to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ErrorDetails from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ErrorDetails from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ErrorDetails from JSON text
	FromJson(value string) error
	// Validate validates ErrorDetails
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Errors returns ErrorDetailsErrorItemIter, set in ErrorDetails
	Errors() ErrorDetailsErrorItemIter
	setNil()
}

// Errors returns a []ErrorItem
// description is TBD
func (obj *errorDetails) Errors() ErrorDetailsErrorItemIter {
	if len(obj.obj.Errors) == 0 {
		obj.obj.Errors = []*onexdataflowapi.ErrorItem{}
	}
	if obj.errorsHolder == nil {
		obj.errorsHolder = newErrorDetailsErrorItemIter().setMsg(obj)
	}
	return obj.errorsHolder
}

type errorDetailsErrorItemIter struct {
	obj            *errorDetails
	errorItemSlice []ErrorItem
}

func newErrorDetailsErrorItemIter() ErrorDetailsErrorItemIter {
	return &errorDetailsErrorItemIter{}
}

type ErrorDetailsErrorItemIter interface {
	setMsg(*errorDetails) ErrorDetailsErrorItemIter
	Items() []ErrorItem
	Add() ErrorItem
	Append(items ...ErrorItem) ErrorDetailsErrorItemIter
	Set(index int, newObj ErrorItem) ErrorDetailsErrorItemIter
	Clear() ErrorDetailsErrorItemIter
	clearHolderSlice() ErrorDetailsErrorItemIter
	appendHolderSlice(item ErrorItem) ErrorDetailsErrorItemIter
}

func (obj *errorDetailsErrorItemIter) setMsg(msg *errorDetails) ErrorDetailsErrorItemIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Errors {
		obj.appendHolderSlice(&errorItem{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *errorDetailsErrorItemIter) Items() []ErrorItem {
	return obj.errorItemSlice
}

func (obj *errorDetailsErrorItemIter) Add() ErrorItem {
	newObj := &onexdataflowapi.ErrorItem{}
	obj.obj.obj.Errors = append(obj.obj.obj.Errors, newObj)
	newLibObj := &errorItem{obj: newObj}
	newLibObj.setDefault()
	obj.errorItemSlice = append(obj.errorItemSlice, newLibObj)
	return newLibObj
}

func (obj *errorDetailsErrorItemIter) Append(items ...ErrorItem) ErrorDetailsErrorItemIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Errors = append(obj.obj.obj.Errors, newObj)
		obj.errorItemSlice = append(obj.errorItemSlice, item)
	}
	return obj
}

func (obj *errorDetailsErrorItemIter) Set(index int, newObj ErrorItem) ErrorDetailsErrorItemIter {
	obj.obj.obj.Errors[index] = newObj.Msg()
	obj.errorItemSlice[index] = newObj
	return obj
}
func (obj *errorDetailsErrorItemIter) Clear() ErrorDetailsErrorItemIter {
	if len(obj.obj.obj.Errors) > 0 {
		obj.obj.obj.Errors = []*onexdataflowapi.ErrorItem{}
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *errorDetailsErrorItemIter) clearHolderSlice() ErrorDetailsErrorItemIter {
	if len(obj.errorItemSlice) > 0 {
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *errorDetailsErrorItemIter) appendHolderSlice(item ErrorItem) ErrorDetailsErrorItemIter {
	obj.errorItemSlice = append(obj.errorItemSlice, item)
	return obj
}

func (obj *errorDetails) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Errors != nil {

		if set_default {
			obj.Errors().clearHolderSlice()
			for _, item := range obj.obj.Errors {
				obj.Errors().appendHolderSlice(&errorItem{obj: item})
			}
		}
		for _, item := range obj.Errors().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *errorDetails) setDefault() {

}

// ***** WarningDetails *****
type warningDetails struct {
	obj            *onexdataflowapi.WarningDetails
	warningsHolder WarningDetailsErrorItemIter
}

func NewWarningDetails() WarningDetails {
	obj := warningDetails{obj: &onexdataflowapi.WarningDetails{}}
	obj.setDefault()
	return &obj
}

func (obj *warningDetails) Msg() *onexdataflowapi.WarningDetails {
	return obj.obj
}

func (obj *warningDetails) SetMsg(msg *onexdataflowapi.WarningDetails) WarningDetails {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *warningDetails) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *warningDetails) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *warningDetails) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *warningDetails) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *warningDetails) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *warningDetails) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *warningDetails) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *warningDetails) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *warningDetails) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *warningDetails) setNil() {
	obj.warningsHolder = nil
}

// WarningDetails is description is TBD
type WarningDetails interface {
	Msg() *onexdataflowapi.WarningDetails
	SetMsg(*onexdataflowapi.WarningDetails) WarningDetails
	// ToPbText marshals WarningDetails to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals WarningDetails to YAML text
	ToYaml() (string, error)
	// ToJson marshals WarningDetails to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals WarningDetails from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals WarningDetails from YAML text
	FromYaml(value string) error
	// FromJson unmarshals WarningDetails from JSON text
	FromJson(value string) error
	// Validate validates WarningDetails
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Warnings returns WarningDetailsErrorItemIter, set in WarningDetails
	Warnings() WarningDetailsErrorItemIter
	setNil()
}

// Warnings returns a []ErrorItem
// description is TBD
func (obj *warningDetails) Warnings() WarningDetailsErrorItemIter {
	if len(obj.obj.Warnings) == 0 {
		obj.obj.Warnings = []*onexdataflowapi.ErrorItem{}
	}
	if obj.warningsHolder == nil {
		obj.warningsHolder = newWarningDetailsErrorItemIter().setMsg(obj)
	}
	return obj.warningsHolder
}

type warningDetailsErrorItemIter struct {
	obj            *warningDetails
	errorItemSlice []ErrorItem
}

func newWarningDetailsErrorItemIter() WarningDetailsErrorItemIter {
	return &warningDetailsErrorItemIter{}
}

type WarningDetailsErrorItemIter interface {
	setMsg(*warningDetails) WarningDetailsErrorItemIter
	Items() []ErrorItem
	Add() ErrorItem
	Append(items ...ErrorItem) WarningDetailsErrorItemIter
	Set(index int, newObj ErrorItem) WarningDetailsErrorItemIter
	Clear() WarningDetailsErrorItemIter
	clearHolderSlice() WarningDetailsErrorItemIter
	appendHolderSlice(item ErrorItem) WarningDetailsErrorItemIter
}

func (obj *warningDetailsErrorItemIter) setMsg(msg *warningDetails) WarningDetailsErrorItemIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Warnings {
		obj.appendHolderSlice(&errorItem{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *warningDetailsErrorItemIter) Items() []ErrorItem {
	return obj.errorItemSlice
}

func (obj *warningDetailsErrorItemIter) Add() ErrorItem {
	newObj := &onexdataflowapi.ErrorItem{}
	obj.obj.obj.Warnings = append(obj.obj.obj.Warnings, newObj)
	newLibObj := &errorItem{obj: newObj}
	newLibObj.setDefault()
	obj.errorItemSlice = append(obj.errorItemSlice, newLibObj)
	return newLibObj
}

func (obj *warningDetailsErrorItemIter) Append(items ...ErrorItem) WarningDetailsErrorItemIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Warnings = append(obj.obj.obj.Warnings, newObj)
		obj.errorItemSlice = append(obj.errorItemSlice, item)
	}
	return obj
}

func (obj *warningDetailsErrorItemIter) Set(index int, newObj ErrorItem) WarningDetailsErrorItemIter {
	obj.obj.obj.Warnings[index] = newObj.Msg()
	obj.errorItemSlice[index] = newObj
	return obj
}
func (obj *warningDetailsErrorItemIter) Clear() WarningDetailsErrorItemIter {
	if len(obj.obj.obj.Warnings) > 0 {
		obj.obj.obj.Warnings = []*onexdataflowapi.ErrorItem{}
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *warningDetailsErrorItemIter) clearHolderSlice() WarningDetailsErrorItemIter {
	if len(obj.errorItemSlice) > 0 {
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *warningDetailsErrorItemIter) appendHolderSlice(item ErrorItem) WarningDetailsErrorItemIter {
	obj.errorItemSlice = append(obj.errorItemSlice, item)
	return obj
}

func (obj *warningDetails) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Warnings != nil {

		if set_default {
			obj.Warnings().clearHolderSlice()
			for _, item := range obj.obj.Warnings {
				obj.Warnings().appendHolderSlice(&errorItem{obj: item})
			}
		}
		for _, item := range obj.Warnings().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *warningDetails) setDefault() {

}

// ***** ControlStatusResponse *****
type controlStatusResponse struct {
	obj          *onexdataflowapi.ControlStatusResponse
	errorsHolder ControlStatusResponseErrorItemIter
}

func NewControlStatusResponse() ControlStatusResponse {
	obj := controlStatusResponse{obj: &onexdataflowapi.ControlStatusResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *controlStatusResponse) Msg() *onexdataflowapi.ControlStatusResponse {
	return obj.obj
}

func (obj *controlStatusResponse) SetMsg(msg *onexdataflowapi.ControlStatusResponse) ControlStatusResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *controlStatusResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *controlStatusResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *controlStatusResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStatusResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *controlStatusResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *controlStatusResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *controlStatusResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *controlStatusResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *controlStatusResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *controlStatusResponse) setNil() {
	obj.errorsHolder = nil
}

// ControlStatusResponse is control/state response details
type ControlStatusResponse interface {
	Msg() *onexdataflowapi.ControlStatusResponse
	SetMsg(*onexdataflowapi.ControlStatusResponse) ControlStatusResponse
	// ToPbText marshals ControlStatusResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ControlStatusResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals ControlStatusResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ControlStatusResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ControlStatusResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ControlStatusResponse from JSON text
	FromJson(value string) error
	// Validate validates ControlStatusResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// DataflowStatus returns ControlStatusResponseDataflowStatusEnum, set in ControlStatusResponse
	DataflowStatus() ControlStatusResponseDataflowStatusEnum
	// SetDataflowStatus assigns ControlStatusResponseDataflowStatusEnum provided by user to ControlStatusResponse
	SetDataflowStatus(value ControlStatusResponseDataflowStatusEnum) ControlStatusResponse
	// HasDataflowStatus checks if DataflowStatus has been set in ControlStatusResponse
	HasDataflowStatus() bool
	// Errors returns ControlStatusResponseErrorItemIter, set in ControlStatusResponse
	Errors() ControlStatusResponseErrorItemIter
	setNil()
}

type ControlStatusResponseDataflowStatusEnum string

//  Enum of DataflowStatus on ControlStatusResponse
var ControlStatusResponseDataflowStatus = struct {
	STARTED   ControlStatusResponseDataflowStatusEnum
	COMPLETED ControlStatusResponseDataflowStatusEnum
	ERROR     ControlStatusResponseDataflowStatusEnum
}{
	STARTED:   ControlStatusResponseDataflowStatusEnum("started"),
	COMPLETED: ControlStatusResponseDataflowStatusEnum("completed"),
	ERROR:     ControlStatusResponseDataflowStatusEnum("error"),
}

func (obj *controlStatusResponse) DataflowStatus() ControlStatusResponseDataflowStatusEnum {
	return ControlStatusResponseDataflowStatusEnum(obj.obj.DataflowStatus.Enum().String())
}

// DataflowStatus returns a string
// dataflow status:
// started - data flow traffic is running
// completed - all traffic flows completed, metrics are available
// error - an error occurred
func (obj *controlStatusResponse) HasDataflowStatus() bool {
	return obj.obj.DataflowStatus != nil
}

func (obj *controlStatusResponse) SetDataflowStatus(value ControlStatusResponseDataflowStatusEnum) ControlStatusResponse {
	intValue, ok := onexdataflowapi.ControlStatusResponse_DataflowStatus_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ControlStatusResponseDataflowStatusEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.ControlStatusResponse_DataflowStatus_Enum(intValue)
	obj.obj.DataflowStatus = &enumValue

	return obj
}

// Errors returns a []ErrorItem
// description is TBD
func (obj *controlStatusResponse) Errors() ControlStatusResponseErrorItemIter {
	if len(obj.obj.Errors) == 0 {
		obj.obj.Errors = []*onexdataflowapi.ErrorItem{}
	}
	if obj.errorsHolder == nil {
		obj.errorsHolder = newControlStatusResponseErrorItemIter().setMsg(obj)
	}
	return obj.errorsHolder
}

type controlStatusResponseErrorItemIter struct {
	obj            *controlStatusResponse
	errorItemSlice []ErrorItem
}

func newControlStatusResponseErrorItemIter() ControlStatusResponseErrorItemIter {
	return &controlStatusResponseErrorItemIter{}
}

type ControlStatusResponseErrorItemIter interface {
	setMsg(*controlStatusResponse) ControlStatusResponseErrorItemIter
	Items() []ErrorItem
	Add() ErrorItem
	Append(items ...ErrorItem) ControlStatusResponseErrorItemIter
	Set(index int, newObj ErrorItem) ControlStatusResponseErrorItemIter
	Clear() ControlStatusResponseErrorItemIter
	clearHolderSlice() ControlStatusResponseErrorItemIter
	appendHolderSlice(item ErrorItem) ControlStatusResponseErrorItemIter
}

func (obj *controlStatusResponseErrorItemIter) setMsg(msg *controlStatusResponse) ControlStatusResponseErrorItemIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Errors {
		obj.appendHolderSlice(&errorItem{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *controlStatusResponseErrorItemIter) Items() []ErrorItem {
	return obj.errorItemSlice
}

func (obj *controlStatusResponseErrorItemIter) Add() ErrorItem {
	newObj := &onexdataflowapi.ErrorItem{}
	obj.obj.obj.Errors = append(obj.obj.obj.Errors, newObj)
	newLibObj := &errorItem{obj: newObj}
	newLibObj.setDefault()
	obj.errorItemSlice = append(obj.errorItemSlice, newLibObj)
	return newLibObj
}

func (obj *controlStatusResponseErrorItemIter) Append(items ...ErrorItem) ControlStatusResponseErrorItemIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Errors = append(obj.obj.obj.Errors, newObj)
		obj.errorItemSlice = append(obj.errorItemSlice, item)
	}
	return obj
}

func (obj *controlStatusResponseErrorItemIter) Set(index int, newObj ErrorItem) ControlStatusResponseErrorItemIter {
	obj.obj.obj.Errors[index] = newObj.Msg()
	obj.errorItemSlice[index] = newObj
	return obj
}
func (obj *controlStatusResponseErrorItemIter) Clear() ControlStatusResponseErrorItemIter {
	if len(obj.obj.obj.Errors) > 0 {
		obj.obj.obj.Errors = []*onexdataflowapi.ErrorItem{}
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *controlStatusResponseErrorItemIter) clearHolderSlice() ControlStatusResponseErrorItemIter {
	if len(obj.errorItemSlice) > 0 {
		obj.errorItemSlice = []ErrorItem{}
	}
	return obj
}
func (obj *controlStatusResponseErrorItemIter) appendHolderSlice(item ErrorItem) ControlStatusResponseErrorItemIter {
	obj.errorItemSlice = append(obj.errorItemSlice, item)
	return obj
}

func (obj *controlStatusResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Errors != nil {

		if set_default {
			obj.Errors().clearHolderSlice()
			for _, item := range obj.obj.Errors {
				obj.Errors().appendHolderSlice(&errorItem{obj: item})
			}
		}
		for _, item := range obj.Errors().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *controlStatusResponse) setDefault() {

}

// ***** MetricsResponse *****
type metricsResponse struct {
	obj               *onexdataflowapi.MetricsResponse
	flowResultsHolder MetricsResponseMetricsResponseFlowResultIter
}

func NewMetricsResponse() MetricsResponse {
	obj := metricsResponse{obj: &onexdataflowapi.MetricsResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponse) Msg() *onexdataflowapi.MetricsResponse {
	return obj.obj
}

func (obj *metricsResponse) SetMsg(msg *onexdataflowapi.MetricsResponse) MetricsResponse {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponse) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *metricsResponse) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsResponse) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponse) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *metricsResponse) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponse) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *metricsResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *metricsResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *metricsResponse) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *metricsResponse) setNil() {
	obj.flowResultsHolder = nil
}

// MetricsResponse is metrics response details
type MetricsResponse interface {
	Msg() *onexdataflowapi.MetricsResponse
	SetMsg(*onexdataflowapi.MetricsResponse) MetricsResponse
	// ToPbText marshals MetricsResponse to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals MetricsResponse to YAML text
	ToYaml() (string, error)
	// ToJson marshals MetricsResponse to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals MetricsResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponse from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponse
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Jct returns int64, set in MetricsResponse.
	Jct() int64
	// SetJct assigns int64 provided by user to MetricsResponse
	SetJct(value int64) MetricsResponse
	// HasJct checks if Jct has been set in MetricsResponse
	HasJct() bool
	// FlowResults returns MetricsResponseMetricsResponseFlowResultIter, set in MetricsResponse
	FlowResults() MetricsResponseMetricsResponseFlowResultIter
	setNil()
}

// Jct returns a int64
// job completion time in micro seconds
func (obj *metricsResponse) Jct() int64 {

	return *obj.obj.Jct

}

// Jct returns a int64
// job completion time in micro seconds
func (obj *metricsResponse) HasJct() bool {
	return obj.obj.Jct != nil
}

// SetJct sets the int64 value in the MetricsResponse object
// job completion time in micro seconds
func (obj *metricsResponse) SetJct(value int64) MetricsResponse {

	obj.obj.Jct = &value
	return obj
}

// FlowResults returns a []MetricsResponseFlowResult
// description is TBD
func (obj *metricsResponse) FlowResults() MetricsResponseMetricsResponseFlowResultIter {
	if len(obj.obj.FlowResults) == 0 {
		obj.obj.FlowResults = []*onexdataflowapi.MetricsResponseFlowResult{}
	}
	if obj.flowResultsHolder == nil {
		obj.flowResultsHolder = newMetricsResponseMetricsResponseFlowResultIter().setMsg(obj)
	}
	return obj.flowResultsHolder
}

type metricsResponseMetricsResponseFlowResultIter struct {
	obj                            *metricsResponse
	metricsResponseFlowResultSlice []MetricsResponseFlowResult
}

func newMetricsResponseMetricsResponseFlowResultIter() MetricsResponseMetricsResponseFlowResultIter {
	return &metricsResponseMetricsResponseFlowResultIter{}
}

type MetricsResponseMetricsResponseFlowResultIter interface {
	setMsg(*metricsResponse) MetricsResponseMetricsResponseFlowResultIter
	Items() []MetricsResponseFlowResult
	Add() MetricsResponseFlowResult
	Append(items ...MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter
	Set(index int, newObj MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter
	Clear() MetricsResponseMetricsResponseFlowResultIter
	clearHolderSlice() MetricsResponseMetricsResponseFlowResultIter
	appendHolderSlice(item MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter
}

func (obj *metricsResponseMetricsResponseFlowResultIter) setMsg(msg *metricsResponse) MetricsResponseMetricsResponseFlowResultIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.FlowResults {
		obj.appendHolderSlice(&metricsResponseFlowResult{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *metricsResponseMetricsResponseFlowResultIter) Items() []MetricsResponseFlowResult {
	return obj.metricsResponseFlowResultSlice
}

func (obj *metricsResponseMetricsResponseFlowResultIter) Add() MetricsResponseFlowResult {
	newObj := &onexdataflowapi.MetricsResponseFlowResult{}
	obj.obj.obj.FlowResults = append(obj.obj.obj.FlowResults, newObj)
	newLibObj := &metricsResponseFlowResult{obj: newObj}
	newLibObj.setDefault()
	obj.metricsResponseFlowResultSlice = append(obj.metricsResponseFlowResultSlice, newLibObj)
	return newLibObj
}

func (obj *metricsResponseMetricsResponseFlowResultIter) Append(items ...MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.FlowResults = append(obj.obj.obj.FlowResults, newObj)
		obj.metricsResponseFlowResultSlice = append(obj.metricsResponseFlowResultSlice, item)
	}
	return obj
}

func (obj *metricsResponseMetricsResponseFlowResultIter) Set(index int, newObj MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter {
	obj.obj.obj.FlowResults[index] = newObj.Msg()
	obj.metricsResponseFlowResultSlice[index] = newObj
	return obj
}
func (obj *metricsResponseMetricsResponseFlowResultIter) Clear() MetricsResponseMetricsResponseFlowResultIter {
	if len(obj.obj.obj.FlowResults) > 0 {
		obj.obj.obj.FlowResults = []*onexdataflowapi.MetricsResponseFlowResult{}
		obj.metricsResponseFlowResultSlice = []MetricsResponseFlowResult{}
	}
	return obj
}
func (obj *metricsResponseMetricsResponseFlowResultIter) clearHolderSlice() MetricsResponseMetricsResponseFlowResultIter {
	if len(obj.metricsResponseFlowResultSlice) > 0 {
		obj.metricsResponseFlowResultSlice = []MetricsResponseFlowResult{}
	}
	return obj
}
func (obj *metricsResponseMetricsResponseFlowResultIter) appendHolderSlice(item MetricsResponseFlowResult) MetricsResponseMetricsResponseFlowResultIter {
	obj.metricsResponseFlowResultSlice = append(obj.metricsResponseFlowResultSlice, item)
	return obj
}

func (obj *metricsResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.FlowResults != nil {

		if set_default {
			obj.FlowResults().clearHolderSlice()
			for _, item := range obj.obj.FlowResults {
				obj.FlowResults().appendHolderSlice(&metricsResponseFlowResult{obj: item})
			}
		}
		for _, item := range obj.FlowResults().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *metricsResponse) setDefault() {

}

// ***** DataflowHostManagement *****
type dataflowHostManagement struct {
	obj *onexdataflowapi.DataflowHostManagement
}

func NewDataflowHostManagement() DataflowHostManagement {
	obj := dataflowHostManagement{obj: &onexdataflowapi.DataflowHostManagement{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowHostManagement) Msg() *onexdataflowapi.DataflowHostManagement {
	return obj.obj
}

func (obj *dataflowHostManagement) SetMsg(msg *onexdataflowapi.DataflowHostManagement) DataflowHostManagement {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowHostManagement) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowHostManagement) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowHostManagement) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowHostManagement) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowHostManagement) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowHostManagement) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowHostManagement) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowHostManagement) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowHostManagement) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowHostManagement is auxillary host information needed to run dataflow experiments
type DataflowHostManagement interface {
	Msg() *onexdataflowapi.DataflowHostManagement
	SetMsg(*onexdataflowapi.DataflowHostManagement) DataflowHostManagement
	// ToPbText marshals DataflowHostManagement to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowHostManagement to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowHostManagement to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowHostManagement from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowHostManagement from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowHostManagement from JSON text
	FromJson(value string) error
	// Validate validates DataflowHostManagement
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// HostName returns string, set in DataflowHostManagement.
	HostName() string
	// SetHostName assigns string provided by user to DataflowHostManagement
	SetHostName(value string) DataflowHostManagement
	// EthNicProfileName returns string, set in DataflowHostManagement.
	EthNicProfileName() string
	// SetEthNicProfileName assigns string provided by user to DataflowHostManagement
	SetEthNicProfileName(value string) DataflowHostManagement
	// HasEthNicProfileName checks if EthNicProfileName has been set in DataflowHostManagement
	HasEthNicProfileName() bool
}

// HostName returns a string
// TBD
//
// x-constraint:
// - #components/schemas/Host/properties/name
//
func (obj *dataflowHostManagement) HostName() string {

	return obj.obj.HostName
}

// SetHostName sets the string value in the DataflowHostManagement object
// TBD
//
// x-constraint:
// - #components/schemas/Host/properties/name
//
func (obj *dataflowHostManagement) SetHostName(value string) DataflowHostManagement {

	obj.obj.HostName = value
	return obj
}

// EthNicProfileName returns a string
// The nic parameters profile associated with the host.
//
// x-constraint:
// - #/components/schemas/Profiles.Dataflow.HostManagement.EthNicSetting/properties/name
//
func (obj *dataflowHostManagement) EthNicProfileName() string {

	return *obj.obj.EthNicProfileName

}

// EthNicProfileName returns a string
// The nic parameters profile associated with the host.
//
// x-constraint:
// - #/components/schemas/Profiles.Dataflow.HostManagement.EthNicSetting/properties/name
//
func (obj *dataflowHostManagement) HasEthNicProfileName() bool {
	return obj.obj.EthNicProfileName != nil
}

// SetEthNicProfileName sets the string value in the DataflowHostManagement object
// The nic parameters profile associated with the host.
//
// x-constraint:
// - #/components/schemas/Profiles.Dataflow.HostManagement.EthNicSetting/properties/name
//
func (obj *dataflowHostManagement) SetEthNicProfileName(value string) DataflowHostManagement {

	obj.obj.EthNicProfileName = &value
	return obj
}

func (obj *dataflowHostManagement) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// HostName is required
	if obj.obj.HostName == "" {
		validation = append(validation, "HostName is required field on interface DataflowHostManagement")
	}
}

func (obj *dataflowHostManagement) setDefault() {

}

// ***** DataflowWorkloadItem *****
type dataflowWorkloadItem struct {
	obj             *onexdataflowapi.DataflowWorkloadItem
	scatterHolder   DataflowScatterWorkload
	gatherHolder    DataflowGatherWorkload
	loopHolder      DataflowLoopWorkload
	computeHolder   DataflowComputeWorkload
	allReduceHolder DataflowAllReduceWorkload
	broadcastHolder DataflowBroadcastWorkload
	allToAllHolder  DataflowAlltoallWorkload
}

func NewDataflowWorkloadItem() DataflowWorkloadItem {
	obj := dataflowWorkloadItem{obj: &onexdataflowapi.DataflowWorkloadItem{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowWorkloadItem) Msg() *onexdataflowapi.DataflowWorkloadItem {
	return obj.obj
}

func (obj *dataflowWorkloadItem) SetMsg(msg *onexdataflowapi.DataflowWorkloadItem) DataflowWorkloadItem {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowWorkloadItem) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowWorkloadItem) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowWorkloadItem) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowWorkloadItem) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowWorkloadItem) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowWorkloadItem) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowWorkloadItem) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowWorkloadItem) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowWorkloadItem) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowWorkloadItem) setNil() {
	obj.scatterHolder = nil
	obj.gatherHolder = nil
	obj.loopHolder = nil
	obj.computeHolder = nil
	obj.allReduceHolder = nil
	obj.broadcastHolder = nil
	obj.allToAllHolder = nil
}

// DataflowWorkloadItem is description is TBD
type DataflowWorkloadItem interface {
	Msg() *onexdataflowapi.DataflowWorkloadItem
	SetMsg(*onexdataflowapi.DataflowWorkloadItem) DataflowWorkloadItem
	// ToPbText marshals DataflowWorkloadItem to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowWorkloadItem to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowWorkloadItem to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowWorkloadItem from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowWorkloadItem from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowWorkloadItem from JSON text
	FromJson(value string) error
	// Validate validates DataflowWorkloadItem
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in DataflowWorkloadItem.
	Name() string
	// SetName assigns string provided by user to DataflowWorkloadItem
	SetName(value string) DataflowWorkloadItem
	// Choice returns DataflowWorkloadItemChoiceEnum, set in DataflowWorkloadItem
	Choice() DataflowWorkloadItemChoiceEnum
	// SetChoice assigns DataflowWorkloadItemChoiceEnum provided by user to DataflowWorkloadItem
	SetChoice(value DataflowWorkloadItemChoiceEnum) DataflowWorkloadItem
	// Scatter returns DataflowScatterWorkload, set in DataflowWorkloadItem.
	// DataflowScatterWorkload is description is TBD
	Scatter() DataflowScatterWorkload
	// SetScatter assigns DataflowScatterWorkload provided by user to DataflowWorkloadItem.
	// DataflowScatterWorkload is description is TBD
	SetScatter(value DataflowScatterWorkload) DataflowWorkloadItem
	// HasScatter checks if Scatter has been set in DataflowWorkloadItem
	HasScatter() bool
	// Gather returns DataflowGatherWorkload, set in DataflowWorkloadItem.
	// DataflowGatherWorkload is description is TBD
	Gather() DataflowGatherWorkload
	// SetGather assigns DataflowGatherWorkload provided by user to DataflowWorkloadItem.
	// DataflowGatherWorkload is description is TBD
	SetGather(value DataflowGatherWorkload) DataflowWorkloadItem
	// HasGather checks if Gather has been set in DataflowWorkloadItem
	HasGather() bool
	// Loop returns DataflowLoopWorkload, set in DataflowWorkloadItem.
	// DataflowLoopWorkload is description is TBD
	Loop() DataflowLoopWorkload
	// SetLoop assigns DataflowLoopWorkload provided by user to DataflowWorkloadItem.
	// DataflowLoopWorkload is description is TBD
	SetLoop(value DataflowLoopWorkload) DataflowWorkloadItem
	// HasLoop checks if Loop has been set in DataflowWorkloadItem
	HasLoop() bool
	// Compute returns DataflowComputeWorkload, set in DataflowWorkloadItem.
	// DataflowComputeWorkload is description is TBD
	Compute() DataflowComputeWorkload
	// SetCompute assigns DataflowComputeWorkload provided by user to DataflowWorkloadItem.
	// DataflowComputeWorkload is description is TBD
	SetCompute(value DataflowComputeWorkload) DataflowWorkloadItem
	// HasCompute checks if Compute has been set in DataflowWorkloadItem
	HasCompute() bool
	// AllReduce returns DataflowAllReduceWorkload, set in DataflowWorkloadItem.
	// DataflowAllReduceWorkload is description is TBD
	AllReduce() DataflowAllReduceWorkload
	// SetAllReduce assigns DataflowAllReduceWorkload provided by user to DataflowWorkloadItem.
	// DataflowAllReduceWorkload is description is TBD
	SetAllReduce(value DataflowAllReduceWorkload) DataflowWorkloadItem
	// HasAllReduce checks if AllReduce has been set in DataflowWorkloadItem
	HasAllReduce() bool
	// Broadcast returns DataflowBroadcastWorkload, set in DataflowWorkloadItem.
	// DataflowBroadcastWorkload is description is TBD
	Broadcast() DataflowBroadcastWorkload
	// SetBroadcast assigns DataflowBroadcastWorkload provided by user to DataflowWorkloadItem.
	// DataflowBroadcastWorkload is description is TBD
	SetBroadcast(value DataflowBroadcastWorkload) DataflowWorkloadItem
	// HasBroadcast checks if Broadcast has been set in DataflowWorkloadItem
	HasBroadcast() bool
	// AllToAll returns DataflowAlltoallWorkload, set in DataflowWorkloadItem.
	// DataflowAlltoallWorkload is creates full-mesh flows between all nodes
	AllToAll() DataflowAlltoallWorkload
	// SetAllToAll assigns DataflowAlltoallWorkload provided by user to DataflowWorkloadItem.
	// DataflowAlltoallWorkload is creates full-mesh flows between all nodes
	SetAllToAll(value DataflowAlltoallWorkload) DataflowWorkloadItem
	// HasAllToAll checks if AllToAll has been set in DataflowWorkloadItem
	HasAllToAll() bool
	setNil()
}

// Name returns a string
// uniquely identifies the workload item
func (obj *dataflowWorkloadItem) Name() string {

	return obj.obj.Name
}

// SetName sets the string value in the DataflowWorkloadItem object
// uniquely identifies the workload item
func (obj *dataflowWorkloadItem) SetName(value string) DataflowWorkloadItem {

	obj.obj.Name = value
	return obj
}

type DataflowWorkloadItemChoiceEnum string

//  Enum of Choice on DataflowWorkloadItem
var DataflowWorkloadItemChoice = struct {
	SCATTER    DataflowWorkloadItemChoiceEnum
	GATHER     DataflowWorkloadItemChoiceEnum
	ALL_REDUCE DataflowWorkloadItemChoiceEnum
	LOOP       DataflowWorkloadItemChoiceEnum
	COMPUTE    DataflowWorkloadItemChoiceEnum
	BROADCAST  DataflowWorkloadItemChoiceEnum
	ALL_TO_ALL DataflowWorkloadItemChoiceEnum
}{
	SCATTER:    DataflowWorkloadItemChoiceEnum("scatter"),
	GATHER:     DataflowWorkloadItemChoiceEnum("gather"),
	ALL_REDUCE: DataflowWorkloadItemChoiceEnum("all_reduce"),
	LOOP:       DataflowWorkloadItemChoiceEnum("loop"),
	COMPUTE:    DataflowWorkloadItemChoiceEnum("compute"),
	BROADCAST:  DataflowWorkloadItemChoiceEnum("broadcast"),
	ALL_TO_ALL: DataflowWorkloadItemChoiceEnum("all_to_all"),
}

func (obj *dataflowWorkloadItem) Choice() DataflowWorkloadItemChoiceEnum {
	return DataflowWorkloadItemChoiceEnum(obj.obj.Choice.Enum().String())
}

func (obj *dataflowWorkloadItem) SetChoice(value DataflowWorkloadItemChoiceEnum) DataflowWorkloadItem {
	intValue, ok := onexdataflowapi.DataflowWorkloadItem_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowWorkloadItemChoiceEnum", string(value)))
		return obj
	}
	obj.obj.Choice = onexdataflowapi.DataflowWorkloadItem_Choice_Enum(intValue)
	obj.obj.AllToAll = nil
	obj.allToAllHolder = nil
	obj.obj.Broadcast = nil
	obj.broadcastHolder = nil
	obj.obj.AllReduce = nil
	obj.allReduceHolder = nil
	obj.obj.Compute = nil
	obj.computeHolder = nil
	obj.obj.Loop = nil
	obj.loopHolder = nil
	obj.obj.Gather = nil
	obj.gatherHolder = nil
	obj.obj.Scatter = nil
	obj.scatterHolder = nil

	if value == DataflowWorkloadItemChoice.SCATTER {
		obj.obj.Scatter = NewDataflowScatterWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.GATHER {
		obj.obj.Gather = NewDataflowGatherWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.LOOP {
		obj.obj.Loop = NewDataflowLoopWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.COMPUTE {
		obj.obj.Compute = NewDataflowComputeWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.ALL_REDUCE {
		obj.obj.AllReduce = NewDataflowAllReduceWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.BROADCAST {
		obj.obj.Broadcast = NewDataflowBroadcastWorkload().Msg()
	}

	if value == DataflowWorkloadItemChoice.ALL_TO_ALL {
		obj.obj.AllToAll = NewDataflowAlltoallWorkload().Msg()
	}

	return obj
}

// Scatter returns a DataflowScatterWorkload
// description is TBD
func (obj *dataflowWorkloadItem) Scatter() DataflowScatterWorkload {
	if obj.obj.Scatter == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.SCATTER)
	}
	if obj.scatterHolder == nil {
		obj.scatterHolder = &dataflowScatterWorkload{obj: obj.obj.Scatter}
	}
	return obj.scatterHolder
}

// Scatter returns a DataflowScatterWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasScatter() bool {
	return obj.obj.Scatter != nil
}

// SetScatter sets the DataflowScatterWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetScatter(value DataflowScatterWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.SCATTER)
	obj.scatterHolder = nil
	obj.obj.Scatter = value.Msg()

	return obj
}

// Gather returns a DataflowGatherWorkload
// description is TBD
func (obj *dataflowWorkloadItem) Gather() DataflowGatherWorkload {
	if obj.obj.Gather == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.GATHER)
	}
	if obj.gatherHolder == nil {
		obj.gatherHolder = &dataflowGatherWorkload{obj: obj.obj.Gather}
	}
	return obj.gatherHolder
}

// Gather returns a DataflowGatherWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasGather() bool {
	return obj.obj.Gather != nil
}

// SetGather sets the DataflowGatherWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetGather(value DataflowGatherWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.GATHER)
	obj.gatherHolder = nil
	obj.obj.Gather = value.Msg()

	return obj
}

// Loop returns a DataflowLoopWorkload
// description is TBD
func (obj *dataflowWorkloadItem) Loop() DataflowLoopWorkload {
	if obj.obj.Loop == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.LOOP)
	}
	if obj.loopHolder == nil {
		obj.loopHolder = &dataflowLoopWorkload{obj: obj.obj.Loop}
	}
	return obj.loopHolder
}

// Loop returns a DataflowLoopWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasLoop() bool {
	return obj.obj.Loop != nil
}

// SetLoop sets the DataflowLoopWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetLoop(value DataflowLoopWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.LOOP)
	obj.loopHolder = nil
	obj.obj.Loop = value.Msg()

	return obj
}

// Compute returns a DataflowComputeWorkload
// description is TBD
func (obj *dataflowWorkloadItem) Compute() DataflowComputeWorkload {
	if obj.obj.Compute == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.COMPUTE)
	}
	if obj.computeHolder == nil {
		obj.computeHolder = &dataflowComputeWorkload{obj: obj.obj.Compute}
	}
	return obj.computeHolder
}

// Compute returns a DataflowComputeWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasCompute() bool {
	return obj.obj.Compute != nil
}

// SetCompute sets the DataflowComputeWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetCompute(value DataflowComputeWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.COMPUTE)
	obj.computeHolder = nil
	obj.obj.Compute = value.Msg()

	return obj
}

// AllReduce returns a DataflowAllReduceWorkload
// description is TBD
func (obj *dataflowWorkloadItem) AllReduce() DataflowAllReduceWorkload {
	if obj.obj.AllReduce == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.ALL_REDUCE)
	}
	if obj.allReduceHolder == nil {
		obj.allReduceHolder = &dataflowAllReduceWorkload{obj: obj.obj.AllReduce}
	}
	return obj.allReduceHolder
}

// AllReduce returns a DataflowAllReduceWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasAllReduce() bool {
	return obj.obj.AllReduce != nil
}

// SetAllReduce sets the DataflowAllReduceWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetAllReduce(value DataflowAllReduceWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.ALL_REDUCE)
	obj.allReduceHolder = nil
	obj.obj.AllReduce = value.Msg()

	return obj
}

// Broadcast returns a DataflowBroadcastWorkload
// description is TBD
func (obj *dataflowWorkloadItem) Broadcast() DataflowBroadcastWorkload {
	if obj.obj.Broadcast == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.BROADCAST)
	}
	if obj.broadcastHolder == nil {
		obj.broadcastHolder = &dataflowBroadcastWorkload{obj: obj.obj.Broadcast}
	}
	return obj.broadcastHolder
}

// Broadcast returns a DataflowBroadcastWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasBroadcast() bool {
	return obj.obj.Broadcast != nil
}

// SetBroadcast sets the DataflowBroadcastWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetBroadcast(value DataflowBroadcastWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.BROADCAST)
	obj.broadcastHolder = nil
	obj.obj.Broadcast = value.Msg()

	return obj
}

// AllToAll returns a DataflowAlltoallWorkload
// description is TBD
func (obj *dataflowWorkloadItem) AllToAll() DataflowAlltoallWorkload {
	if obj.obj.AllToAll == nil {
		obj.SetChoice(DataflowWorkloadItemChoice.ALL_TO_ALL)
	}
	if obj.allToAllHolder == nil {
		obj.allToAllHolder = &dataflowAlltoallWorkload{obj: obj.obj.AllToAll}
	}
	return obj.allToAllHolder
}

// AllToAll returns a DataflowAlltoallWorkload
// description is TBD
func (obj *dataflowWorkloadItem) HasAllToAll() bool {
	return obj.obj.AllToAll != nil
}

// SetAllToAll sets the DataflowAlltoallWorkload value in the DataflowWorkloadItem object
// description is TBD
func (obj *dataflowWorkloadItem) SetAllToAll(value DataflowAlltoallWorkload) DataflowWorkloadItem {
	obj.SetChoice(DataflowWorkloadItemChoice.ALL_TO_ALL)
	obj.allToAllHolder = nil
	obj.obj.AllToAll = value.Msg()

	return obj
}

func (obj *dataflowWorkloadItem) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// Name is required
	if obj.obj.Name == "" {
		validation = append(validation, "Name is required field on interface DataflowWorkloadItem")
	}

	// Choice is required
	if obj.obj.Choice.Number() == 0 {
		validation = append(validation, "Choice is required field on interface DataflowWorkloadItem")
	}

	if obj.obj.Scatter != nil {
		obj.Scatter().validateObj(set_default)
	}

	if obj.obj.Gather != nil {
		obj.Gather().validateObj(set_default)
	}

	if obj.obj.Loop != nil {
		obj.Loop().validateObj(set_default)
	}

	if obj.obj.Compute != nil {
		obj.Compute().validateObj(set_default)
	}

	if obj.obj.AllReduce != nil {
		obj.AllReduce().validateObj(set_default)
	}

	if obj.obj.Broadcast != nil {
		obj.Broadcast().validateObj(set_default)
	}

	if obj.obj.AllToAll != nil {
		obj.AllToAll().validateObj(set_default)
	}

}

func (obj *dataflowWorkloadItem) setDefault() {

}

// ***** DataflowFlowProfile *****
type dataflowFlowProfile struct {
	obj         *onexdataflowapi.DataflowFlowProfile
	rdmaHolder  DataflowFlowProfileRdmaStack
	tcpipHolder DataflowFlowProfileTcpIpStack
}

func NewDataflowFlowProfile() DataflowFlowProfile {
	obj := dataflowFlowProfile{obj: &onexdataflowapi.DataflowFlowProfile{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfile) Msg() *onexdataflowapi.DataflowFlowProfile {
	return obj.obj
}

func (obj *dataflowFlowProfile) SetMsg(msg *onexdataflowapi.DataflowFlowProfile) DataflowFlowProfile {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfile) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfile) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfile) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfile) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfile) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfile) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfile) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfile) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfile) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowFlowProfile) setNil() {
	obj.rdmaHolder = nil
	obj.tcpipHolder = nil
}

// DataflowFlowProfile is description is TBD
type DataflowFlowProfile interface {
	Msg() *onexdataflowapi.DataflowFlowProfile
	SetMsg(*onexdataflowapi.DataflowFlowProfile) DataflowFlowProfile
	// ToPbText marshals DataflowFlowProfile to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfile to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfile to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfile from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfile from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfile from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfile
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in DataflowFlowProfile.
	Name() string
	// SetName assigns string provided by user to DataflowFlowProfile
	SetName(value string) DataflowFlowProfile
	// DataSize returns int32, set in DataflowFlowProfile.
	DataSize() int32
	// SetDataSize assigns int32 provided by user to DataflowFlowProfile
	SetDataSize(value int32) DataflowFlowProfile
	// Bidirectional returns bool, set in DataflowFlowProfile.
	Bidirectional() bool
	// SetBidirectional assigns bool provided by user to DataflowFlowProfile
	SetBidirectional(value bool) DataflowFlowProfile
	// HasBidirectional checks if Bidirectional has been set in DataflowFlowProfile
	HasBidirectional() bool
	// Iterations returns int32, set in DataflowFlowProfile.
	Iterations() int32
	// SetIterations assigns int32 provided by user to DataflowFlowProfile
	SetIterations(value int32) DataflowFlowProfile
	// HasIterations checks if Iterations has been set in DataflowFlowProfile
	HasIterations() bool
	// Choice returns DataflowFlowProfileChoiceEnum, set in DataflowFlowProfile
	Choice() DataflowFlowProfileChoiceEnum
	// SetChoice assigns DataflowFlowProfileChoiceEnum provided by user to DataflowFlowProfile
	SetChoice(value DataflowFlowProfileChoiceEnum) DataflowFlowProfile
	// HasChoice checks if Choice has been set in DataflowFlowProfile
	HasChoice() bool
	// Rdma returns DataflowFlowProfileRdmaStack, set in DataflowFlowProfile.
	// DataflowFlowProfileRdmaStack is description is TBD
	Rdma() DataflowFlowProfileRdmaStack
	// SetRdma assigns DataflowFlowProfileRdmaStack provided by user to DataflowFlowProfile.
	// DataflowFlowProfileRdmaStack is description is TBD
	SetRdma(value DataflowFlowProfileRdmaStack) DataflowFlowProfile
	// HasRdma checks if Rdma has been set in DataflowFlowProfile
	HasRdma() bool
	// Tcpip returns DataflowFlowProfileTcpIpStack, set in DataflowFlowProfile.
	// DataflowFlowProfileTcpIpStack is description is TBD
	Tcpip() DataflowFlowProfileTcpIpStack
	// SetTcpip assigns DataflowFlowProfileTcpIpStack provided by user to DataflowFlowProfile.
	// DataflowFlowProfileTcpIpStack is description is TBD
	SetTcpip(value DataflowFlowProfileTcpIpStack) DataflowFlowProfile
	// HasTcpip checks if Tcpip has been set in DataflowFlowProfile
	HasTcpip() bool
	setNil()
}

// Name returns a string
// description is TBD
func (obj *dataflowFlowProfile) Name() string {

	return obj.obj.Name
}

// SetName sets the string value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetName(value string) DataflowFlowProfile {

	obj.obj.Name = value
	return obj
}

// DataSize returns a int32
// description is TBD
func (obj *dataflowFlowProfile) DataSize() int32 {

	return obj.obj.DataSize
}

// SetDataSize sets the int32 value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetDataSize(value int32) DataflowFlowProfile {

	obj.obj.DataSize = value
	return obj
}

// Bidirectional returns a bool
// whether data is sent both ways
func (obj *dataflowFlowProfile) Bidirectional() bool {

	return *obj.obj.Bidirectional

}

// Bidirectional returns a bool
// whether data is sent both ways
func (obj *dataflowFlowProfile) HasBidirectional() bool {
	return obj.obj.Bidirectional != nil
}

// SetBidirectional sets the bool value in the DataflowFlowProfile object
// whether data is sent both ways
func (obj *dataflowFlowProfile) SetBidirectional(value bool) DataflowFlowProfile {

	obj.obj.Bidirectional = &value
	return obj
}

// Iterations returns a int32
// how many times to send the message
func (obj *dataflowFlowProfile) Iterations() int32 {

	return *obj.obj.Iterations

}

// Iterations returns a int32
// how many times to send the message
func (obj *dataflowFlowProfile) HasIterations() bool {
	return obj.obj.Iterations != nil
}

// SetIterations sets the int32 value in the DataflowFlowProfile object
// how many times to send the message
func (obj *dataflowFlowProfile) SetIterations(value int32) DataflowFlowProfile {

	obj.obj.Iterations = &value
	return obj
}

type DataflowFlowProfileChoiceEnum string

//  Enum of Choice on DataflowFlowProfile
var DataflowFlowProfileChoice = struct {
	RDMA  DataflowFlowProfileChoiceEnum
	TCPIP DataflowFlowProfileChoiceEnum
}{
	RDMA:  DataflowFlowProfileChoiceEnum("rdma"),
	TCPIP: DataflowFlowProfileChoiceEnum("tcpip"),
}

func (obj *dataflowFlowProfile) Choice() DataflowFlowProfileChoiceEnum {
	return DataflowFlowProfileChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// RDMA traffic or traditional TCP/IP
func (obj *dataflowFlowProfile) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *dataflowFlowProfile) SetChoice(value DataflowFlowProfileChoiceEnum) DataflowFlowProfile {
	intValue, ok := onexdataflowapi.DataflowFlowProfile_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfile_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Tcpip = nil
	obj.tcpipHolder = nil
	obj.obj.Rdma = nil
	obj.rdmaHolder = nil

	if value == DataflowFlowProfileChoice.RDMA {
		obj.obj.Rdma = NewDataflowFlowProfileRdmaStack().Msg()
	}

	if value == DataflowFlowProfileChoice.TCPIP {
		obj.obj.Tcpip = NewDataflowFlowProfileTcpIpStack().Msg()
	}

	return obj
}

// Rdma returns a DataflowFlowProfileRdmaStack
// description is TBD
func (obj *dataflowFlowProfile) Rdma() DataflowFlowProfileRdmaStack {
	if obj.obj.Rdma == nil {
		obj.SetChoice(DataflowFlowProfileChoice.RDMA)
	}
	if obj.rdmaHolder == nil {
		obj.rdmaHolder = &dataflowFlowProfileRdmaStack{obj: obj.obj.Rdma}
	}
	return obj.rdmaHolder
}

// Rdma returns a DataflowFlowProfileRdmaStack
// description is TBD
func (obj *dataflowFlowProfile) HasRdma() bool {
	return obj.obj.Rdma != nil
}

// SetRdma sets the DataflowFlowProfileRdmaStack value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetRdma(value DataflowFlowProfileRdmaStack) DataflowFlowProfile {
	obj.SetChoice(DataflowFlowProfileChoice.RDMA)
	obj.rdmaHolder = nil
	obj.obj.Rdma = value.Msg()

	return obj
}

// Tcpip returns a DataflowFlowProfileTcpIpStack
// description is TBD
func (obj *dataflowFlowProfile) Tcpip() DataflowFlowProfileTcpIpStack {
	if obj.obj.Tcpip == nil {
		obj.SetChoice(DataflowFlowProfileChoice.TCPIP)
	}
	if obj.tcpipHolder == nil {
		obj.tcpipHolder = &dataflowFlowProfileTcpIpStack{obj: obj.obj.Tcpip}
	}
	return obj.tcpipHolder
}

// Tcpip returns a DataflowFlowProfileTcpIpStack
// description is TBD
func (obj *dataflowFlowProfile) HasTcpip() bool {
	return obj.obj.Tcpip != nil
}

// SetTcpip sets the DataflowFlowProfileTcpIpStack value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetTcpip(value DataflowFlowProfileTcpIpStack) DataflowFlowProfile {
	obj.SetChoice(DataflowFlowProfileChoice.TCPIP)
	obj.tcpipHolder = nil
	obj.obj.Tcpip = value.Msg()

	return obj
}

func (obj *dataflowFlowProfile) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// Name is required
	if obj.obj.Name == "" {
		validation = append(validation, "Name is required field on interface DataflowFlowProfile")
	}

	if obj.obj.Rdma != nil {
		obj.Rdma().validateObj(set_default)
	}

	if obj.obj.Tcpip != nil {
		obj.Tcpip().validateObj(set_default)
	}

}

func (obj *dataflowFlowProfile) setDefault() {
	if obj.obj.Bidirectional == nil {
		obj.SetBidirectional(false)
	}
	if obj.obj.Iterations == nil {
		obj.SetIterations(1)
	}

}

// ***** ErrorItem *****
type errorItem struct {
	obj *onexdataflowapi.ErrorItem
}

func NewErrorItem() ErrorItem {
	obj := errorItem{obj: &onexdataflowapi.ErrorItem{}}
	obj.setDefault()
	return &obj
}

func (obj *errorItem) Msg() *onexdataflowapi.ErrorItem {
	return obj.obj
}

func (obj *errorItem) SetMsg(msg *onexdataflowapi.ErrorItem) ErrorItem {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *errorItem) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *errorItem) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *errorItem) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *errorItem) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *errorItem) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *errorItem) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *errorItem) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *errorItem) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *errorItem) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// ErrorItem is description is TBD
type ErrorItem interface {
	Msg() *onexdataflowapi.ErrorItem
	SetMsg(*onexdataflowapi.ErrorItem) ErrorItem
	// ToPbText marshals ErrorItem to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals ErrorItem to YAML text
	ToYaml() (string, error)
	// ToJson marshals ErrorItem to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals ErrorItem from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ErrorItem from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ErrorItem from JSON text
	FromJson(value string) error
	// Validate validates ErrorItem
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Message returns string, set in ErrorItem.
	Message() string
	// SetMessage assigns string provided by user to ErrorItem
	SetMessage(value string) ErrorItem
	// HasMessage checks if Message has been set in ErrorItem
	HasMessage() bool
	// Code returns int32, set in ErrorItem.
	Code() int32
	// SetCode assigns int32 provided by user to ErrorItem
	SetCode(value int32) ErrorItem
	// HasCode checks if Code has been set in ErrorItem
	HasCode() bool
	// Detail returns string, set in ErrorItem.
	Detail() string
	// SetDetail assigns string provided by user to ErrorItem
	SetDetail(value string) ErrorItem
	// HasDetail checks if Detail has been set in ErrorItem
	HasDetail() bool
}

// Message returns a string
// description is TBD
func (obj *errorItem) Message() string {

	return *obj.obj.Message

}

// Message returns a string
// description is TBD
func (obj *errorItem) HasMessage() bool {
	return obj.obj.Message != nil
}

// SetMessage sets the string value in the ErrorItem object
// description is TBD
func (obj *errorItem) SetMessage(value string) ErrorItem {

	obj.obj.Message = &value
	return obj
}

// Code returns a int32
// description is TBD
func (obj *errorItem) Code() int32 {

	return *obj.obj.Code

}

// Code returns a int32
// description is TBD
func (obj *errorItem) HasCode() bool {
	return obj.obj.Code != nil
}

// SetCode sets the int32 value in the ErrorItem object
// description is TBD
func (obj *errorItem) SetCode(value int32) ErrorItem {

	obj.obj.Code = &value
	return obj
}

// Detail returns a string
// description is TBD
func (obj *errorItem) Detail() string {

	return *obj.obj.Detail

}

// Detail returns a string
// description is TBD
func (obj *errorItem) HasDetail() bool {
	return obj.obj.Detail != nil
}

// SetDetail sets the string value in the ErrorItem object
// description is TBD
func (obj *errorItem) SetDetail(value string) ErrorItem {

	obj.obj.Detail = &value
	return obj
}

func (obj *errorItem) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *errorItem) setDefault() {

}

// ***** MetricsResponseFlowResult *****
type metricsResponseFlowResult struct {
	obj                    *onexdataflowapi.MetricsResponseFlowResult
	tcpInfoInitiatorHolder MetricsResponseFlowResultTcpInfo
	tcpInfoResponderHolder MetricsResponseFlowResultTcpInfo
}

func NewMetricsResponseFlowResult() MetricsResponseFlowResult {
	obj := metricsResponseFlowResult{obj: &onexdataflowapi.MetricsResponseFlowResult{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponseFlowResult) Msg() *onexdataflowapi.MetricsResponseFlowResult {
	return obj.obj
}

func (obj *metricsResponseFlowResult) SetMsg(msg *onexdataflowapi.MetricsResponseFlowResult) MetricsResponseFlowResult {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponseFlowResult) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *metricsResponseFlowResult) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsResponseFlowResult) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponseFlowResult) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *metricsResponseFlowResult) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponseFlowResult) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *metricsResponseFlowResult) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *metricsResponseFlowResult) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *metricsResponseFlowResult) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *metricsResponseFlowResult) setNil() {
	obj.tcpInfoInitiatorHolder = nil
	obj.tcpInfoResponderHolder = nil
}

// MetricsResponseFlowResult is result for a single data flow
type MetricsResponseFlowResult interface {
	Msg() *onexdataflowapi.MetricsResponseFlowResult
	SetMsg(*onexdataflowapi.MetricsResponseFlowResult) MetricsResponseFlowResult
	// ToPbText marshals MetricsResponseFlowResult to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals MetricsResponseFlowResult to YAML text
	ToYaml() (string, error)
	// ToJson marshals MetricsResponseFlowResult to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals MetricsResponseFlowResult from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponseFlowResult from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponseFlowResult from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponseFlowResult
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// WorkloadName returns string, set in MetricsResponseFlowResult.
	WorkloadName() string
	// SetWorkloadName assigns string provided by user to MetricsResponseFlowResult
	SetWorkloadName(value string) MetricsResponseFlowResult
	// HasWorkloadName checks if WorkloadName has been set in MetricsResponseFlowResult
	HasWorkloadName() bool
	// FromHostName returns string, set in MetricsResponseFlowResult.
	FromHostName() string
	// SetFromHostName assigns string provided by user to MetricsResponseFlowResult
	SetFromHostName(value string) MetricsResponseFlowResult
	// HasFromHostName checks if FromHostName has been set in MetricsResponseFlowResult
	HasFromHostName() bool
	// ToHostName returns string, set in MetricsResponseFlowResult.
	ToHostName() string
	// SetToHostName assigns string provided by user to MetricsResponseFlowResult
	SetToHostName(value string) MetricsResponseFlowResult
	// HasToHostName checks if ToHostName has been set in MetricsResponseFlowResult
	HasToHostName() bool
	// Fct returns int64, set in MetricsResponseFlowResult.
	Fct() int64
	// SetFct assigns int64 provided by user to MetricsResponseFlowResult
	SetFct(value int64) MetricsResponseFlowResult
	// HasFct checks if Fct has been set in MetricsResponseFlowResult
	HasFct() bool
	// FirstTimestamp returns int64, set in MetricsResponseFlowResult.
	FirstTimestamp() int64
	// SetFirstTimestamp assigns int64 provided by user to MetricsResponseFlowResult
	SetFirstTimestamp(value int64) MetricsResponseFlowResult
	// HasFirstTimestamp checks if FirstTimestamp has been set in MetricsResponseFlowResult
	HasFirstTimestamp() bool
	// LastTimestamp returns int64, set in MetricsResponseFlowResult.
	LastTimestamp() int64
	// SetLastTimestamp assigns int64 provided by user to MetricsResponseFlowResult
	SetLastTimestamp(value int64) MetricsResponseFlowResult
	// HasLastTimestamp checks if LastTimestamp has been set in MetricsResponseFlowResult
	HasLastTimestamp() bool
	// BytesTx returns int64, set in MetricsResponseFlowResult.
	BytesTx() int64
	// SetBytesTx assigns int64 provided by user to MetricsResponseFlowResult
	SetBytesTx(value int64) MetricsResponseFlowResult
	// HasBytesTx checks if BytesTx has been set in MetricsResponseFlowResult
	HasBytesTx() bool
	// BytesRx returns int64, set in MetricsResponseFlowResult.
	BytesRx() int64
	// SetBytesRx assigns int64 provided by user to MetricsResponseFlowResult
	SetBytesRx(value int64) MetricsResponseFlowResult
	// HasBytesRx checks if BytesRx has been set in MetricsResponseFlowResult
	HasBytesRx() bool
	// TcpInfoInitiator returns MetricsResponseFlowResultTcpInfo, set in MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	TcpInfoInitiator() MetricsResponseFlowResultTcpInfo
	// SetTcpInfoInitiator assigns MetricsResponseFlowResultTcpInfo provided by user to MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	SetTcpInfoInitiator(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult
	// HasTcpInfoInitiator checks if TcpInfoInitiator has been set in MetricsResponseFlowResult
	HasTcpInfoInitiator() bool
	// TcpInfoResponder returns MetricsResponseFlowResultTcpInfo, set in MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	TcpInfoResponder() MetricsResponseFlowResultTcpInfo
	// SetTcpInfoResponder assigns MetricsResponseFlowResultTcpInfo provided by user to MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	SetTcpInfoResponder(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult
	// HasTcpInfoResponder checks if TcpInfoResponder has been set in MetricsResponseFlowResult
	HasTcpInfoResponder() bool
	setNil()
}

// WorkloadName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) WorkloadName() string {

	return *obj.obj.WorkloadName

}

// WorkloadName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) HasWorkloadName() bool {
	return obj.obj.WorkloadName != nil
}

// SetWorkloadName sets the string value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetWorkloadName(value string) MetricsResponseFlowResult {

	obj.obj.WorkloadName = &value
	return obj
}

// FromHostName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) FromHostName() string {

	return *obj.obj.FromHostName

}

// FromHostName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) HasFromHostName() bool {
	return obj.obj.FromHostName != nil
}

// SetFromHostName sets the string value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetFromHostName(value string) MetricsResponseFlowResult {

	obj.obj.FromHostName = &value
	return obj
}

// ToHostName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) ToHostName() string {

	return *obj.obj.ToHostName

}

// ToHostName returns a string
// description is TBD
func (obj *metricsResponseFlowResult) HasToHostName() bool {
	return obj.obj.ToHostName != nil
}

// SetToHostName sets the string value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetToHostName(value string) MetricsResponseFlowResult {

	obj.obj.ToHostName = &value
	return obj
}

// Fct returns a int64
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) Fct() int64 {

	return *obj.obj.Fct

}

// Fct returns a int64
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) HasFct() bool {
	return obj.obj.Fct != nil
}

// SetFct sets the int64 value in the MetricsResponseFlowResult object
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) SetFct(value int64) MetricsResponseFlowResult {

	obj.obj.Fct = &value
	return obj
}

// FirstTimestamp returns a int64
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) FirstTimestamp() int64 {

	return *obj.obj.FirstTimestamp

}

// FirstTimestamp returns a int64
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) HasFirstTimestamp() bool {
	return obj.obj.FirstTimestamp != nil
}

// SetFirstTimestamp sets the int64 value in the MetricsResponseFlowResult object
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) SetFirstTimestamp(value int64) MetricsResponseFlowResult {

	obj.obj.FirstTimestamp = &value
	return obj
}

// LastTimestamp returns a int64
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) LastTimestamp() int64 {

	return *obj.obj.LastTimestamp

}

// LastTimestamp returns a int64
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) HasLastTimestamp() bool {
	return obj.obj.LastTimestamp != nil
}

// SetLastTimestamp sets the int64 value in the MetricsResponseFlowResult object
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) SetLastTimestamp(value int64) MetricsResponseFlowResult {

	obj.obj.LastTimestamp = &value
	return obj
}

// BytesTx returns a int64
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) BytesTx() int64 {

	return *obj.obj.BytesTx

}

// BytesTx returns a int64
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) HasBytesTx() bool {
	return obj.obj.BytesTx != nil
}

// SetBytesTx sets the int64 value in the MetricsResponseFlowResult object
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) SetBytesTx(value int64) MetricsResponseFlowResult {

	obj.obj.BytesTx = &value
	return obj
}

// BytesRx returns a int64
// bytes received by src from dst
func (obj *metricsResponseFlowResult) BytesRx() int64 {

	return *obj.obj.BytesRx

}

// BytesRx returns a int64
// bytes received by src from dst
func (obj *metricsResponseFlowResult) HasBytesRx() bool {
	return obj.obj.BytesRx != nil
}

// SetBytesRx sets the int64 value in the MetricsResponseFlowResult object
// bytes received by src from dst
func (obj *metricsResponseFlowResult) SetBytesRx(value int64) MetricsResponseFlowResult {

	obj.obj.BytesRx = &value
	return obj
}

// TcpInfoInitiator returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) TcpInfoInitiator() MetricsResponseFlowResultTcpInfo {
	if obj.obj.TcpInfoInitiator == nil {
		obj.obj.TcpInfoInitiator = NewMetricsResponseFlowResultTcpInfo().Msg()
	}
	if obj.tcpInfoInitiatorHolder == nil {
		obj.tcpInfoInitiatorHolder = &metricsResponseFlowResultTcpInfo{obj: obj.obj.TcpInfoInitiator}
	}
	return obj.tcpInfoInitiatorHolder
}

// TcpInfoInitiator returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) HasTcpInfoInitiator() bool {
	return obj.obj.TcpInfoInitiator != nil
}

// SetTcpInfoInitiator sets the MetricsResponseFlowResultTcpInfo value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetTcpInfoInitiator(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult {

	obj.tcpInfoInitiatorHolder = nil
	obj.obj.TcpInfoInitiator = value.Msg()

	return obj
}

// TcpInfoResponder returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) TcpInfoResponder() MetricsResponseFlowResultTcpInfo {
	if obj.obj.TcpInfoResponder == nil {
		obj.obj.TcpInfoResponder = NewMetricsResponseFlowResultTcpInfo().Msg()
	}
	if obj.tcpInfoResponderHolder == nil {
		obj.tcpInfoResponderHolder = &metricsResponseFlowResultTcpInfo{obj: obj.obj.TcpInfoResponder}
	}
	return obj.tcpInfoResponderHolder
}

// TcpInfoResponder returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) HasTcpInfoResponder() bool {
	return obj.obj.TcpInfoResponder != nil
}

// SetTcpInfoResponder sets the MetricsResponseFlowResultTcpInfo value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetTcpInfoResponder(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult {

	obj.tcpInfoResponderHolder = nil
	obj.obj.TcpInfoResponder = value.Msg()

	return obj
}

func (obj *metricsResponseFlowResult) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.TcpInfoInitiator != nil {
		obj.TcpInfoInitiator().validateObj(set_default)
	}

	if obj.obj.TcpInfoResponder != nil {
		obj.TcpInfoResponder().validateObj(set_default)
	}

}

func (obj *metricsResponseFlowResult) setDefault() {

}

// ***** DataflowScatterWorkload *****
type dataflowScatterWorkload struct {
	obj *onexdataflowapi.DataflowScatterWorkload
}

func NewDataflowScatterWorkload() DataflowScatterWorkload {
	obj := dataflowScatterWorkload{obj: &onexdataflowapi.DataflowScatterWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowScatterWorkload) Msg() *onexdataflowapi.DataflowScatterWorkload {
	return obj.obj
}

func (obj *dataflowScatterWorkload) SetMsg(msg *onexdataflowapi.DataflowScatterWorkload) DataflowScatterWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowScatterWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowScatterWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowScatterWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowScatterWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowScatterWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowScatterWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowScatterWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowScatterWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowScatterWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowScatterWorkload is description is TBD
type DataflowScatterWorkload interface {
	Msg() *onexdataflowapi.DataflowScatterWorkload
	SetMsg(*onexdataflowapi.DataflowScatterWorkload) DataflowScatterWorkload
	// ToPbText marshals DataflowScatterWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowScatterWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowScatterWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowScatterWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowScatterWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowScatterWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowScatterWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Sources returns []string, set in DataflowScatterWorkload.
	Sources() []string
	// SetSources assigns []string provided by user to DataflowScatterWorkload
	SetSources(value []string) DataflowScatterWorkload
	// Destinations returns []string, set in DataflowScatterWorkload.
	Destinations() []string
	// SetDestinations assigns []string provided by user to DataflowScatterWorkload
	SetDestinations(value []string) DataflowScatterWorkload
	// FlowProfileName returns string, set in DataflowScatterWorkload.
	FlowProfileName() string
	// SetFlowProfileName assigns string provided by user to DataflowScatterWorkload
	SetFlowProfileName(value string) DataflowScatterWorkload
	// HasFlowProfileName checks if FlowProfileName has been set in DataflowScatterWorkload
	HasFlowProfileName() bool
}

// Sources returns a []string
// list of host names, indicating the originator of the data
func (obj *dataflowScatterWorkload) Sources() []string {
	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	return obj.obj.Sources
}

// SetSources sets the []string value in the DataflowScatterWorkload object
// list of host names, indicating the originator of the data
func (obj *dataflowScatterWorkload) SetSources(value []string) DataflowScatterWorkload {

	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	obj.obj.Sources = value

	return obj
}

// Destinations returns a []string
// list of host names, indicating the destination of the data
func (obj *dataflowScatterWorkload) Destinations() []string {
	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	return obj.obj.Destinations
}

// SetDestinations sets the []string value in the DataflowScatterWorkload object
// list of host names, indicating the destination of the data
func (obj *dataflowScatterWorkload) SetDestinations(value []string) DataflowScatterWorkload {

	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	obj.obj.Destinations = value

	return obj
}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowScatterWorkload) FlowProfileName() string {

	return *obj.obj.FlowProfileName

}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowScatterWorkload) HasFlowProfileName() bool {
	return obj.obj.FlowProfileName != nil
}

// SetFlowProfileName sets the string value in the DataflowScatterWorkload object
// flow profile reference
func (obj *dataflowScatterWorkload) SetFlowProfileName(value string) DataflowScatterWorkload {

	obj.obj.FlowProfileName = &value
	return obj
}

func (obj *dataflowScatterWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowScatterWorkload) setDefault() {

}

// ***** DataflowGatherWorkload *****
type dataflowGatherWorkload struct {
	obj *onexdataflowapi.DataflowGatherWorkload
}

func NewDataflowGatherWorkload() DataflowGatherWorkload {
	obj := dataflowGatherWorkload{obj: &onexdataflowapi.DataflowGatherWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowGatherWorkload) Msg() *onexdataflowapi.DataflowGatherWorkload {
	return obj.obj
}

func (obj *dataflowGatherWorkload) SetMsg(msg *onexdataflowapi.DataflowGatherWorkload) DataflowGatherWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowGatherWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowGatherWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowGatherWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowGatherWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowGatherWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowGatherWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowGatherWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowGatherWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowGatherWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowGatherWorkload is description is TBD
type DataflowGatherWorkload interface {
	Msg() *onexdataflowapi.DataflowGatherWorkload
	SetMsg(*onexdataflowapi.DataflowGatherWorkload) DataflowGatherWorkload
	// ToPbText marshals DataflowGatherWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowGatherWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowGatherWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowGatherWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowGatherWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowGatherWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowGatherWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Sources returns []string, set in DataflowGatherWorkload.
	Sources() []string
	// SetSources assigns []string provided by user to DataflowGatherWorkload
	SetSources(value []string) DataflowGatherWorkload
	// Destinations returns []string, set in DataflowGatherWorkload.
	Destinations() []string
	// SetDestinations assigns []string provided by user to DataflowGatherWorkload
	SetDestinations(value []string) DataflowGatherWorkload
	// FlowProfileName returns string, set in DataflowGatherWorkload.
	FlowProfileName() string
	// SetFlowProfileName assigns string provided by user to DataflowGatherWorkload
	SetFlowProfileName(value string) DataflowGatherWorkload
	// HasFlowProfileName checks if FlowProfileName has been set in DataflowGatherWorkload
	HasFlowProfileName() bool
}

// Sources returns a []string
// list of host names, indicating the originator of the data
func (obj *dataflowGatherWorkload) Sources() []string {
	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	return obj.obj.Sources
}

// SetSources sets the []string value in the DataflowGatherWorkload object
// list of host names, indicating the originator of the data
func (obj *dataflowGatherWorkload) SetSources(value []string) DataflowGatherWorkload {

	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	obj.obj.Sources = value

	return obj
}

// Destinations returns a []string
// list of host names, indicating the destination of the data
func (obj *dataflowGatherWorkload) Destinations() []string {
	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	return obj.obj.Destinations
}

// SetDestinations sets the []string value in the DataflowGatherWorkload object
// list of host names, indicating the destination of the data
func (obj *dataflowGatherWorkload) SetDestinations(value []string) DataflowGatherWorkload {

	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	obj.obj.Destinations = value

	return obj
}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowGatherWorkload) FlowProfileName() string {

	return *obj.obj.FlowProfileName

}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowGatherWorkload) HasFlowProfileName() bool {
	return obj.obj.FlowProfileName != nil
}

// SetFlowProfileName sets the string value in the DataflowGatherWorkload object
// flow profile reference
func (obj *dataflowGatherWorkload) SetFlowProfileName(value string) DataflowGatherWorkload {

	obj.obj.FlowProfileName = &value
	return obj
}

func (obj *dataflowGatherWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowGatherWorkload) setDefault() {

}

// ***** DataflowLoopWorkload *****
type dataflowLoopWorkload struct {
	obj            *onexdataflowapi.DataflowLoopWorkload
	childrenHolder DataflowLoopWorkloadDataflowWorkloadItemIter
}

func NewDataflowLoopWorkload() DataflowLoopWorkload {
	obj := dataflowLoopWorkload{obj: &onexdataflowapi.DataflowLoopWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowLoopWorkload) Msg() *onexdataflowapi.DataflowLoopWorkload {
	return obj.obj
}

func (obj *dataflowLoopWorkload) SetMsg(msg *onexdataflowapi.DataflowLoopWorkload) DataflowLoopWorkload {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowLoopWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowLoopWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowLoopWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowLoopWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowLoopWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowLoopWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowLoopWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowLoopWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowLoopWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowLoopWorkload) setNil() {
	obj.childrenHolder = nil
}

// DataflowLoopWorkload is description is TBD
type DataflowLoopWorkload interface {
	Msg() *onexdataflowapi.DataflowLoopWorkload
	SetMsg(*onexdataflowapi.DataflowLoopWorkload) DataflowLoopWorkload
	// ToPbText marshals DataflowLoopWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowLoopWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowLoopWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowLoopWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowLoopWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowLoopWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowLoopWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Iterations returns int32, set in DataflowLoopWorkload.
	Iterations() int32
	// SetIterations assigns int32 provided by user to DataflowLoopWorkload
	SetIterations(value int32) DataflowLoopWorkload
	// HasIterations checks if Iterations has been set in DataflowLoopWorkload
	HasIterations() bool
	// Children returns DataflowLoopWorkloadDataflowWorkloadItemIter, set in DataflowLoopWorkload
	Children() DataflowLoopWorkloadDataflowWorkloadItemIter
	setNil()
}

// Iterations returns a int32
// number of iterations in the loop
func (obj *dataflowLoopWorkload) Iterations() int32 {

	return *obj.obj.Iterations

}

// Iterations returns a int32
// number of iterations in the loop
func (obj *dataflowLoopWorkload) HasIterations() bool {
	return obj.obj.Iterations != nil
}

// SetIterations sets the int32 value in the DataflowLoopWorkload object
// number of iterations in the loop
func (obj *dataflowLoopWorkload) SetIterations(value int32) DataflowLoopWorkload {

	obj.obj.Iterations = &value
	return obj
}

// Children returns a []DataflowWorkloadItem
// list of workload items that are executed in this loop
func (obj *dataflowLoopWorkload) Children() DataflowLoopWorkloadDataflowWorkloadItemIter {
	if len(obj.obj.Children) == 0 {
		obj.obj.Children = []*onexdataflowapi.DataflowWorkloadItem{}
	}
	if obj.childrenHolder == nil {
		obj.childrenHolder = newDataflowLoopWorkloadDataflowWorkloadItemIter().setMsg(obj)
	}
	return obj.childrenHolder
}

type dataflowLoopWorkloadDataflowWorkloadItemIter struct {
	obj                       *dataflowLoopWorkload
	dataflowWorkloadItemSlice []DataflowWorkloadItem
}

func newDataflowLoopWorkloadDataflowWorkloadItemIter() DataflowLoopWorkloadDataflowWorkloadItemIter {
	return &dataflowLoopWorkloadDataflowWorkloadItemIter{}
}

type DataflowLoopWorkloadDataflowWorkloadItemIter interface {
	setMsg(*dataflowLoopWorkload) DataflowLoopWorkloadDataflowWorkloadItemIter
	Items() []DataflowWorkloadItem
	Add() DataflowWorkloadItem
	Append(items ...DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter
	Set(index int, newObj DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter
	Clear() DataflowLoopWorkloadDataflowWorkloadItemIter
	clearHolderSlice() DataflowLoopWorkloadDataflowWorkloadItemIter
	appendHolderSlice(item DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter
}

func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) setMsg(msg *dataflowLoopWorkload) DataflowLoopWorkloadDataflowWorkloadItemIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Children {
		obj.appendHolderSlice(&dataflowWorkloadItem{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) Items() []DataflowWorkloadItem {
	return obj.dataflowWorkloadItemSlice
}

func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) Add() DataflowWorkloadItem {
	newObj := &onexdataflowapi.DataflowWorkloadItem{}
	obj.obj.obj.Children = append(obj.obj.obj.Children, newObj)
	newLibObj := &dataflowWorkloadItem{obj: newObj}
	newLibObj.setDefault()
	obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, newLibObj)
	return newLibObj
}

func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) Append(items ...DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Children = append(obj.obj.obj.Children, newObj)
		obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, item)
	}
	return obj
}

func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) Set(index int, newObj DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter {
	obj.obj.obj.Children[index] = newObj.Msg()
	obj.dataflowWorkloadItemSlice[index] = newObj
	return obj
}
func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) Clear() DataflowLoopWorkloadDataflowWorkloadItemIter {
	if len(obj.obj.obj.Children) > 0 {
		obj.obj.obj.Children = []*onexdataflowapi.DataflowWorkloadItem{}
		obj.dataflowWorkloadItemSlice = []DataflowWorkloadItem{}
	}
	return obj
}
func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) clearHolderSlice() DataflowLoopWorkloadDataflowWorkloadItemIter {
	if len(obj.dataflowWorkloadItemSlice) > 0 {
		obj.dataflowWorkloadItemSlice = []DataflowWorkloadItem{}
	}
	return obj
}
func (obj *dataflowLoopWorkloadDataflowWorkloadItemIter) appendHolderSlice(item DataflowWorkloadItem) DataflowLoopWorkloadDataflowWorkloadItemIter {
	obj.dataflowWorkloadItemSlice = append(obj.dataflowWorkloadItemSlice, item)
	return obj
}

func (obj *dataflowLoopWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Children != nil {

		if set_default {
			obj.Children().clearHolderSlice()
			for _, item := range obj.obj.Children {
				obj.Children().appendHolderSlice(&dataflowWorkloadItem{obj: item})
			}
		}
		for _, item := range obj.Children().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *dataflowLoopWorkload) setDefault() {

}

// ***** DataflowComputeWorkload *****
type dataflowComputeWorkload struct {
	obj             *onexdataflowapi.DataflowComputeWorkload
	simulatedHolder DataflowSimulatedComputeWorkload
}

func NewDataflowComputeWorkload() DataflowComputeWorkload {
	obj := dataflowComputeWorkload{obj: &onexdataflowapi.DataflowComputeWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowComputeWorkload) Msg() *onexdataflowapi.DataflowComputeWorkload {
	return obj.obj
}

func (obj *dataflowComputeWorkload) SetMsg(msg *onexdataflowapi.DataflowComputeWorkload) DataflowComputeWorkload {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowComputeWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowComputeWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowComputeWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowComputeWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowComputeWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowComputeWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowComputeWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowComputeWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowComputeWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowComputeWorkload) setNil() {
	obj.simulatedHolder = nil
}

// DataflowComputeWorkload is description is TBD
type DataflowComputeWorkload interface {
	Msg() *onexdataflowapi.DataflowComputeWorkload
	SetMsg(*onexdataflowapi.DataflowComputeWorkload) DataflowComputeWorkload
	// ToPbText marshals DataflowComputeWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowComputeWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowComputeWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowComputeWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowComputeWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowComputeWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowComputeWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Nodes returns []string, set in DataflowComputeWorkload.
	Nodes() []string
	// SetNodes assigns []string provided by user to DataflowComputeWorkload
	SetNodes(value []string) DataflowComputeWorkload
	// Choice returns DataflowComputeWorkloadChoiceEnum, set in DataflowComputeWorkload
	Choice() DataflowComputeWorkloadChoiceEnum
	// SetChoice assigns DataflowComputeWorkloadChoiceEnum provided by user to DataflowComputeWorkload
	SetChoice(value DataflowComputeWorkloadChoiceEnum) DataflowComputeWorkload
	// HasChoice checks if Choice has been set in DataflowComputeWorkload
	HasChoice() bool
	// Simulated returns DataflowSimulatedComputeWorkload, set in DataflowComputeWorkload.
	// DataflowSimulatedComputeWorkload is description is TBD
	Simulated() DataflowSimulatedComputeWorkload
	// SetSimulated assigns DataflowSimulatedComputeWorkload provided by user to DataflowComputeWorkload.
	// DataflowSimulatedComputeWorkload is description is TBD
	SetSimulated(value DataflowSimulatedComputeWorkload) DataflowComputeWorkload
	// HasSimulated checks if Simulated has been set in DataflowComputeWorkload
	HasSimulated() bool
	setNil()
}

// Nodes returns a []string
// description is TBD
func (obj *dataflowComputeWorkload) Nodes() []string {
	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	return obj.obj.Nodes
}

// SetNodes sets the []string value in the DataflowComputeWorkload object
// description is TBD
func (obj *dataflowComputeWorkload) SetNodes(value []string) DataflowComputeWorkload {

	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	obj.obj.Nodes = value

	return obj
}

type DataflowComputeWorkloadChoiceEnum string

//  Enum of Choice on DataflowComputeWorkload
var DataflowComputeWorkloadChoice = struct {
	SIMULATED DataflowComputeWorkloadChoiceEnum
}{
	SIMULATED: DataflowComputeWorkloadChoiceEnum("simulated"),
}

func (obj *dataflowComputeWorkload) Choice() DataflowComputeWorkloadChoiceEnum {
	return DataflowComputeWorkloadChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// type of compute
func (obj *dataflowComputeWorkload) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *dataflowComputeWorkload) SetChoice(value DataflowComputeWorkloadChoiceEnum) DataflowComputeWorkload {
	intValue, ok := onexdataflowapi.DataflowComputeWorkload_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowComputeWorkloadChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowComputeWorkload_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Simulated = nil
	obj.simulatedHolder = nil

	if value == DataflowComputeWorkloadChoice.SIMULATED {
		obj.obj.Simulated = NewDataflowSimulatedComputeWorkload().Msg()
	}

	return obj
}

// Simulated returns a DataflowSimulatedComputeWorkload
// description is TBD
func (obj *dataflowComputeWorkload) Simulated() DataflowSimulatedComputeWorkload {
	if obj.obj.Simulated == nil {
		obj.SetChoice(DataflowComputeWorkloadChoice.SIMULATED)
	}
	if obj.simulatedHolder == nil {
		obj.simulatedHolder = &dataflowSimulatedComputeWorkload{obj: obj.obj.Simulated}
	}
	return obj.simulatedHolder
}

// Simulated returns a DataflowSimulatedComputeWorkload
// description is TBD
func (obj *dataflowComputeWorkload) HasSimulated() bool {
	return obj.obj.Simulated != nil
}

// SetSimulated sets the DataflowSimulatedComputeWorkload value in the DataflowComputeWorkload object
// description is TBD
func (obj *dataflowComputeWorkload) SetSimulated(value DataflowSimulatedComputeWorkload) DataflowComputeWorkload {
	obj.SetChoice(DataflowComputeWorkloadChoice.SIMULATED)
	obj.simulatedHolder = nil
	obj.obj.Simulated = value.Msg()

	return obj
}

func (obj *dataflowComputeWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Simulated != nil {
		obj.Simulated().validateObj(set_default)
	}

}

func (obj *dataflowComputeWorkload) setDefault() {

}

// ***** DataflowAllReduceWorkload *****
type dataflowAllReduceWorkload struct {
	obj *onexdataflowapi.DataflowAllReduceWorkload
}

func NewDataflowAllReduceWorkload() DataflowAllReduceWorkload {
	obj := dataflowAllReduceWorkload{obj: &onexdataflowapi.DataflowAllReduceWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowAllReduceWorkload) Msg() *onexdataflowapi.DataflowAllReduceWorkload {
	return obj.obj
}

func (obj *dataflowAllReduceWorkload) SetMsg(msg *onexdataflowapi.DataflowAllReduceWorkload) DataflowAllReduceWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowAllReduceWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowAllReduceWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowAllReduceWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowAllReduceWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowAllReduceWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowAllReduceWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowAllReduceWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowAllReduceWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowAllReduceWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowAllReduceWorkload is description is TBD
type DataflowAllReduceWorkload interface {
	Msg() *onexdataflowapi.DataflowAllReduceWorkload
	SetMsg(*onexdataflowapi.DataflowAllReduceWorkload) DataflowAllReduceWorkload
	// ToPbText marshals DataflowAllReduceWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowAllReduceWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowAllReduceWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowAllReduceWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowAllReduceWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowAllReduceWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowAllReduceWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Nodes returns []string, set in DataflowAllReduceWorkload.
	Nodes() []string
	// SetNodes assigns []string provided by user to DataflowAllReduceWorkload
	SetNodes(value []string) DataflowAllReduceWorkload
	// FlowProfileName returns string, set in DataflowAllReduceWorkload.
	FlowProfileName() string
	// SetFlowProfileName assigns string provided by user to DataflowAllReduceWorkload
	SetFlowProfileName(value string) DataflowAllReduceWorkload
	// HasFlowProfileName checks if FlowProfileName has been set in DataflowAllReduceWorkload
	HasFlowProfileName() bool
	// Type returns DataflowAllReduceWorkloadTypeEnum, set in DataflowAllReduceWorkload
	Type() DataflowAllReduceWorkloadTypeEnum
	// SetType assigns DataflowAllReduceWorkloadTypeEnum provided by user to DataflowAllReduceWorkload
	SetType(value DataflowAllReduceWorkloadTypeEnum) DataflowAllReduceWorkload
	// HasType checks if Type has been set in DataflowAllReduceWorkload
	HasType() bool
}

// Nodes returns a []string
// description is TBD
func (obj *dataflowAllReduceWorkload) Nodes() []string {
	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	return obj.obj.Nodes
}

// SetNodes sets the []string value in the DataflowAllReduceWorkload object
// description is TBD
func (obj *dataflowAllReduceWorkload) SetNodes(value []string) DataflowAllReduceWorkload {

	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	obj.obj.Nodes = value

	return obj
}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowAllReduceWorkload) FlowProfileName() string {

	return *obj.obj.FlowProfileName

}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowAllReduceWorkload) HasFlowProfileName() bool {
	return obj.obj.FlowProfileName != nil
}

// SetFlowProfileName sets the string value in the DataflowAllReduceWorkload object
// flow profile reference
func (obj *dataflowAllReduceWorkload) SetFlowProfileName(value string) DataflowAllReduceWorkload {

	obj.obj.FlowProfileName = &value
	return obj
}

type DataflowAllReduceWorkloadTypeEnum string

//  Enum of Type on DataflowAllReduceWorkload
var DataflowAllReduceWorkloadType = struct {
	RING      DataflowAllReduceWorkloadTypeEnum
	TREE      DataflowAllReduceWorkloadTypeEnum
	BUTTERFLY DataflowAllReduceWorkloadTypeEnum
}{
	RING:      DataflowAllReduceWorkloadTypeEnum("ring"),
	TREE:      DataflowAllReduceWorkloadTypeEnum("tree"),
	BUTTERFLY: DataflowAllReduceWorkloadTypeEnum("butterfly"),
}

func (obj *dataflowAllReduceWorkload) Type() DataflowAllReduceWorkloadTypeEnum {
	return DataflowAllReduceWorkloadTypeEnum(obj.obj.Type.Enum().String())
}

// Type returns a string
// type of all reduce
func (obj *dataflowAllReduceWorkload) HasType() bool {
	return obj.obj.Type != nil
}

func (obj *dataflowAllReduceWorkload) SetType(value DataflowAllReduceWorkloadTypeEnum) DataflowAllReduceWorkload {
	intValue, ok := onexdataflowapi.DataflowAllReduceWorkload_Type_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowAllReduceWorkloadTypeEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowAllReduceWorkload_Type_Enum(intValue)
	obj.obj.Type = &enumValue

	return obj
}

func (obj *dataflowAllReduceWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowAllReduceWorkload) setDefault() {
	if obj.obj.Type == nil {
		obj.SetType(DataflowAllReduceWorkloadType.RING)

	}

}

// ***** DataflowBroadcastWorkload *****
type dataflowBroadcastWorkload struct {
	obj *onexdataflowapi.DataflowBroadcastWorkload
}

func NewDataflowBroadcastWorkload() DataflowBroadcastWorkload {
	obj := dataflowBroadcastWorkload{obj: &onexdataflowapi.DataflowBroadcastWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowBroadcastWorkload) Msg() *onexdataflowapi.DataflowBroadcastWorkload {
	return obj.obj
}

func (obj *dataflowBroadcastWorkload) SetMsg(msg *onexdataflowapi.DataflowBroadcastWorkload) DataflowBroadcastWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowBroadcastWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowBroadcastWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowBroadcastWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowBroadcastWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowBroadcastWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowBroadcastWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowBroadcastWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowBroadcastWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowBroadcastWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowBroadcastWorkload is description is TBD
type DataflowBroadcastWorkload interface {
	Msg() *onexdataflowapi.DataflowBroadcastWorkload
	SetMsg(*onexdataflowapi.DataflowBroadcastWorkload) DataflowBroadcastWorkload
	// ToPbText marshals DataflowBroadcastWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowBroadcastWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowBroadcastWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowBroadcastWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowBroadcastWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowBroadcastWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowBroadcastWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Sources returns []string, set in DataflowBroadcastWorkload.
	Sources() []string
	// SetSources assigns []string provided by user to DataflowBroadcastWorkload
	SetSources(value []string) DataflowBroadcastWorkload
	// Destinations returns []string, set in DataflowBroadcastWorkload.
	Destinations() []string
	// SetDestinations assigns []string provided by user to DataflowBroadcastWorkload
	SetDestinations(value []string) DataflowBroadcastWorkload
	// FlowProfileName returns string, set in DataflowBroadcastWorkload.
	FlowProfileName() string
	// SetFlowProfileName assigns string provided by user to DataflowBroadcastWorkload
	SetFlowProfileName(value string) DataflowBroadcastWorkload
	// HasFlowProfileName checks if FlowProfileName has been set in DataflowBroadcastWorkload
	HasFlowProfileName() bool
}

// Sources returns a []string
// list of host names, indicating the originator of the data
func (obj *dataflowBroadcastWorkload) Sources() []string {
	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	return obj.obj.Sources
}

// SetSources sets the []string value in the DataflowBroadcastWorkload object
// list of host names, indicating the originator of the data
func (obj *dataflowBroadcastWorkload) SetSources(value []string) DataflowBroadcastWorkload {

	if obj.obj.Sources == nil {
		obj.obj.Sources = make([]string, 0)
	}
	obj.obj.Sources = value

	return obj
}

// Destinations returns a []string
// list of host names, indicating the destination of the data
func (obj *dataflowBroadcastWorkload) Destinations() []string {
	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	return obj.obj.Destinations
}

// SetDestinations sets the []string value in the DataflowBroadcastWorkload object
// list of host names, indicating the destination of the data
func (obj *dataflowBroadcastWorkload) SetDestinations(value []string) DataflowBroadcastWorkload {

	if obj.obj.Destinations == nil {
		obj.obj.Destinations = make([]string, 0)
	}
	obj.obj.Destinations = value

	return obj
}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowBroadcastWorkload) FlowProfileName() string {

	return *obj.obj.FlowProfileName

}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowBroadcastWorkload) HasFlowProfileName() bool {
	return obj.obj.FlowProfileName != nil
}

// SetFlowProfileName sets the string value in the DataflowBroadcastWorkload object
// flow profile reference
func (obj *dataflowBroadcastWorkload) SetFlowProfileName(value string) DataflowBroadcastWorkload {

	obj.obj.FlowProfileName = &value
	return obj
}

func (obj *dataflowBroadcastWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowBroadcastWorkload) setDefault() {

}

// ***** DataflowAlltoallWorkload *****
type dataflowAlltoallWorkload struct {
	obj *onexdataflowapi.DataflowAlltoallWorkload
}

func NewDataflowAlltoallWorkload() DataflowAlltoallWorkload {
	obj := dataflowAlltoallWorkload{obj: &onexdataflowapi.DataflowAlltoallWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowAlltoallWorkload) Msg() *onexdataflowapi.DataflowAlltoallWorkload {
	return obj.obj
}

func (obj *dataflowAlltoallWorkload) SetMsg(msg *onexdataflowapi.DataflowAlltoallWorkload) DataflowAlltoallWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowAlltoallWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowAlltoallWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowAlltoallWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowAlltoallWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowAlltoallWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowAlltoallWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowAlltoallWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowAlltoallWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowAlltoallWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowAlltoallWorkload is creates full-mesh flows between all nodes
type DataflowAlltoallWorkload interface {
	Msg() *onexdataflowapi.DataflowAlltoallWorkload
	SetMsg(*onexdataflowapi.DataflowAlltoallWorkload) DataflowAlltoallWorkload
	// ToPbText marshals DataflowAlltoallWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowAlltoallWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowAlltoallWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowAlltoallWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowAlltoallWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowAlltoallWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowAlltoallWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Nodes returns []string, set in DataflowAlltoallWorkload.
	Nodes() []string
	// SetNodes assigns []string provided by user to DataflowAlltoallWorkload
	SetNodes(value []string) DataflowAlltoallWorkload
	// FlowProfileName returns string, set in DataflowAlltoallWorkload.
	FlowProfileName() string
	// SetFlowProfileName assigns string provided by user to DataflowAlltoallWorkload
	SetFlowProfileName(value string) DataflowAlltoallWorkload
	// HasFlowProfileName checks if FlowProfileName has been set in DataflowAlltoallWorkload
	HasFlowProfileName() bool
}

// Nodes returns a []string
// description is TBD
func (obj *dataflowAlltoallWorkload) Nodes() []string {
	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	return obj.obj.Nodes
}

// SetNodes sets the []string value in the DataflowAlltoallWorkload object
// description is TBD
func (obj *dataflowAlltoallWorkload) SetNodes(value []string) DataflowAlltoallWorkload {

	if obj.obj.Nodes == nil {
		obj.obj.Nodes = make([]string, 0)
	}
	obj.obj.Nodes = value

	return obj
}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowAlltoallWorkload) FlowProfileName() string {

	return *obj.obj.FlowProfileName

}

// FlowProfileName returns a string
// flow profile reference
func (obj *dataflowAlltoallWorkload) HasFlowProfileName() bool {
	return obj.obj.FlowProfileName != nil
}

// SetFlowProfileName sets the string value in the DataflowAlltoallWorkload object
// flow profile reference
func (obj *dataflowAlltoallWorkload) SetFlowProfileName(value string) DataflowAlltoallWorkload {

	obj.obj.FlowProfileName = &value
	return obj
}

func (obj *dataflowAlltoallWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowAlltoallWorkload) setDefault() {

}

// ***** DataflowFlowProfileRdmaStack *****
type dataflowFlowProfileRdmaStack struct {
	obj          *onexdataflowapi.DataflowFlowProfileRdmaStack
	rocev2Holder DataflowFlowProfileRdmaStackRoceV2
}

func NewDataflowFlowProfileRdmaStack() DataflowFlowProfileRdmaStack {
	obj := dataflowFlowProfileRdmaStack{obj: &onexdataflowapi.DataflowFlowProfileRdmaStack{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileRdmaStack) Msg() *onexdataflowapi.DataflowFlowProfileRdmaStack {
	return obj.obj
}

func (obj *dataflowFlowProfileRdmaStack) SetMsg(msg *onexdataflowapi.DataflowFlowProfileRdmaStack) DataflowFlowProfileRdmaStack {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileRdmaStack) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileRdmaStack) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileRdmaStack) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileRdmaStack) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileRdmaStack) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileRdmaStack) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileRdmaStack) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileRdmaStack) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileRdmaStack) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowFlowProfileRdmaStack) setNil() {
	obj.rocev2Holder = nil
}

// DataflowFlowProfileRdmaStack is description is TBD
type DataflowFlowProfileRdmaStack interface {
	Msg() *onexdataflowapi.DataflowFlowProfileRdmaStack
	SetMsg(*onexdataflowapi.DataflowFlowProfileRdmaStack) DataflowFlowProfileRdmaStack
	// ToPbText marshals DataflowFlowProfileRdmaStack to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileRdmaStack to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileRdmaStack to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileRdmaStack from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileRdmaStack from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileRdmaStack from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileRdmaStack
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Choice returns DataflowFlowProfileRdmaStackChoiceEnum, set in DataflowFlowProfileRdmaStack
	Choice() DataflowFlowProfileRdmaStackChoiceEnum
	// SetChoice assigns DataflowFlowProfileRdmaStackChoiceEnum provided by user to DataflowFlowProfileRdmaStack
	SetChoice(value DataflowFlowProfileRdmaStackChoiceEnum) DataflowFlowProfileRdmaStack
	// HasChoice checks if Choice has been set in DataflowFlowProfileRdmaStack
	HasChoice() bool
	// Rocev2 returns DataflowFlowProfileRdmaStackRoceV2, set in DataflowFlowProfileRdmaStack.
	// DataflowFlowProfileRdmaStackRoceV2 is description is TBD
	Rocev2() DataflowFlowProfileRdmaStackRoceV2
	// SetRocev2 assigns DataflowFlowProfileRdmaStackRoceV2 provided by user to DataflowFlowProfileRdmaStack.
	// DataflowFlowProfileRdmaStackRoceV2 is description is TBD
	SetRocev2(value DataflowFlowProfileRdmaStackRoceV2) DataflowFlowProfileRdmaStack
	// HasRocev2 checks if Rocev2 has been set in DataflowFlowProfileRdmaStack
	HasRocev2() bool
	setNil()
}

type DataflowFlowProfileRdmaStackChoiceEnum string

//  Enum of Choice on DataflowFlowProfileRdmaStack
var DataflowFlowProfileRdmaStackChoice = struct {
	ROCEV2 DataflowFlowProfileRdmaStackChoiceEnum
}{
	ROCEV2: DataflowFlowProfileRdmaStackChoiceEnum("rocev2"),
}

func (obj *dataflowFlowProfileRdmaStack) Choice() DataflowFlowProfileRdmaStackChoiceEnum {
	return DataflowFlowProfileRdmaStackChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *dataflowFlowProfileRdmaStack) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *dataflowFlowProfileRdmaStack) SetChoice(value DataflowFlowProfileRdmaStackChoiceEnum) DataflowFlowProfileRdmaStack {
	intValue, ok := onexdataflowapi.DataflowFlowProfileRdmaStack_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileRdmaStackChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfileRdmaStack_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Rocev2 = nil
	obj.rocev2Holder = nil

	if value == DataflowFlowProfileRdmaStackChoice.ROCEV2 {
		obj.obj.Rocev2 = NewDataflowFlowProfileRdmaStackRoceV2().Msg()
	}

	return obj
}

// Rocev2 returns a DataflowFlowProfileRdmaStackRoceV2
// description is TBD
func (obj *dataflowFlowProfileRdmaStack) Rocev2() DataflowFlowProfileRdmaStackRoceV2 {
	if obj.obj.Rocev2 == nil {
		obj.SetChoice(DataflowFlowProfileRdmaStackChoice.ROCEV2)
	}
	if obj.rocev2Holder == nil {
		obj.rocev2Holder = &dataflowFlowProfileRdmaStackRoceV2{obj: obj.obj.Rocev2}
	}
	return obj.rocev2Holder
}

// Rocev2 returns a DataflowFlowProfileRdmaStackRoceV2
// description is TBD
func (obj *dataflowFlowProfileRdmaStack) HasRocev2() bool {
	return obj.obj.Rocev2 != nil
}

// SetRocev2 sets the DataflowFlowProfileRdmaStackRoceV2 value in the DataflowFlowProfileRdmaStack object
// description is TBD
func (obj *dataflowFlowProfileRdmaStack) SetRocev2(value DataflowFlowProfileRdmaStackRoceV2) DataflowFlowProfileRdmaStack {
	obj.SetChoice(DataflowFlowProfileRdmaStackChoice.ROCEV2)
	obj.rocev2Holder = nil
	obj.obj.Rocev2 = value.Msg()

	return obj
}

func (obj *dataflowFlowProfileRdmaStack) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Rocev2 != nil {
		obj.Rocev2().validateObj(set_default)
	}

}

func (obj *dataflowFlowProfileRdmaStack) setDefault() {

}

// ***** DataflowFlowProfileTcpIpStack *****
type dataflowFlowProfileTcpIpStack struct {
	obj       *onexdataflowapi.DataflowFlowProfileTcpIpStack
	ipHolder  DataflowFlowProfileTcpIpStackIp
	tcpHolder DataflowFlowProfileL4ProtocolTcp
	udpHolder DataflowFlowProfileL4ProtocolUdp
}

func NewDataflowFlowProfileTcpIpStack() DataflowFlowProfileTcpIpStack {
	obj := dataflowFlowProfileTcpIpStack{obj: &onexdataflowapi.DataflowFlowProfileTcpIpStack{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileTcpIpStack) Msg() *onexdataflowapi.DataflowFlowProfileTcpIpStack {
	return obj.obj
}

func (obj *dataflowFlowProfileTcpIpStack) SetMsg(msg *onexdataflowapi.DataflowFlowProfileTcpIpStack) DataflowFlowProfileTcpIpStack {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileTcpIpStack) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileTcpIpStack) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileTcpIpStack) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileTcpIpStack) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileTcpIpStack) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileTcpIpStack) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileTcpIpStack) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileTcpIpStack) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileTcpIpStack) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowFlowProfileTcpIpStack) setNil() {
	obj.ipHolder = nil
	obj.tcpHolder = nil
	obj.udpHolder = nil
}

// DataflowFlowProfileTcpIpStack is description is TBD
type DataflowFlowProfileTcpIpStack interface {
	Msg() *onexdataflowapi.DataflowFlowProfileTcpIpStack
	SetMsg(*onexdataflowapi.DataflowFlowProfileTcpIpStack) DataflowFlowProfileTcpIpStack
	// ToPbText marshals DataflowFlowProfileTcpIpStack to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileTcpIpStack to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileTcpIpStack to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileTcpIpStack from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileTcpIpStack from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileTcpIpStack from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileTcpIpStack
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Ip returns DataflowFlowProfileTcpIpStackIp, set in DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileTcpIpStackIp is description is TBD
	Ip() DataflowFlowProfileTcpIpStackIp
	// SetIp assigns DataflowFlowProfileTcpIpStackIp provided by user to DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileTcpIpStackIp is description is TBD
	SetIp(value DataflowFlowProfileTcpIpStackIp) DataflowFlowProfileTcpIpStack
	// HasIp checks if Ip has been set in DataflowFlowProfileTcpIpStack
	HasIp() bool
	// Choice returns DataflowFlowProfileTcpIpStackChoiceEnum, set in DataflowFlowProfileTcpIpStack
	Choice() DataflowFlowProfileTcpIpStackChoiceEnum
	// SetChoice assigns DataflowFlowProfileTcpIpStackChoiceEnum provided by user to DataflowFlowProfileTcpIpStack
	SetChoice(value DataflowFlowProfileTcpIpStackChoiceEnum) DataflowFlowProfileTcpIpStack
	// HasChoice checks if Choice has been set in DataflowFlowProfileTcpIpStack
	HasChoice() bool
	// Tcp returns DataflowFlowProfileL4ProtocolTcp, set in DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileL4ProtocolTcp is description is TBD
	Tcp() DataflowFlowProfileL4ProtocolTcp
	// SetTcp assigns DataflowFlowProfileL4ProtocolTcp provided by user to DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileL4ProtocolTcp is description is TBD
	SetTcp(value DataflowFlowProfileL4ProtocolTcp) DataflowFlowProfileTcpIpStack
	// HasTcp checks if Tcp has been set in DataflowFlowProfileTcpIpStack
	HasTcp() bool
	// Udp returns DataflowFlowProfileL4ProtocolUdp, set in DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileL4ProtocolUdp is description is TBD
	Udp() DataflowFlowProfileL4ProtocolUdp
	// SetUdp assigns DataflowFlowProfileL4ProtocolUdp provided by user to DataflowFlowProfileTcpIpStack.
	// DataflowFlowProfileL4ProtocolUdp is description is TBD
	SetUdp(value DataflowFlowProfileL4ProtocolUdp) DataflowFlowProfileTcpIpStack
	// HasUdp checks if Udp has been set in DataflowFlowProfileTcpIpStack
	HasUdp() bool
	setNil()
}

// Ip returns a DataflowFlowProfileTcpIpStackIp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) Ip() DataflowFlowProfileTcpIpStackIp {
	if obj.obj.Ip == nil {
		obj.obj.Ip = NewDataflowFlowProfileTcpIpStackIp().Msg()
	}
	if obj.ipHolder == nil {
		obj.ipHolder = &dataflowFlowProfileTcpIpStackIp{obj: obj.obj.Ip}
	}
	return obj.ipHolder
}

// Ip returns a DataflowFlowProfileTcpIpStackIp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) HasIp() bool {
	return obj.obj.Ip != nil
}

// SetIp sets the DataflowFlowProfileTcpIpStackIp value in the DataflowFlowProfileTcpIpStack object
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) SetIp(value DataflowFlowProfileTcpIpStackIp) DataflowFlowProfileTcpIpStack {

	obj.ipHolder = nil
	obj.obj.Ip = value.Msg()

	return obj
}

type DataflowFlowProfileTcpIpStackChoiceEnum string

//  Enum of Choice on DataflowFlowProfileTcpIpStack
var DataflowFlowProfileTcpIpStackChoice = struct {
	TCP DataflowFlowProfileTcpIpStackChoiceEnum
	UDP DataflowFlowProfileTcpIpStackChoiceEnum
}{
	TCP: DataflowFlowProfileTcpIpStackChoiceEnum("tcp"),
	UDP: DataflowFlowProfileTcpIpStackChoiceEnum("udp"),
}

func (obj *dataflowFlowProfileTcpIpStack) Choice() DataflowFlowProfileTcpIpStackChoiceEnum {
	return DataflowFlowProfileTcpIpStackChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// layer4 protocol selection
func (obj *dataflowFlowProfileTcpIpStack) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *dataflowFlowProfileTcpIpStack) SetChoice(value DataflowFlowProfileTcpIpStackChoiceEnum) DataflowFlowProfileTcpIpStack {
	intValue, ok := onexdataflowapi.DataflowFlowProfileTcpIpStack_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileTcpIpStackChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfileTcpIpStack_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Udp = nil
	obj.udpHolder = nil
	obj.obj.Tcp = nil
	obj.tcpHolder = nil

	if value == DataflowFlowProfileTcpIpStackChoice.TCP {
		obj.obj.Tcp = NewDataflowFlowProfileL4ProtocolTcp().Msg()
	}

	if value == DataflowFlowProfileTcpIpStackChoice.UDP {
		obj.obj.Udp = NewDataflowFlowProfileL4ProtocolUdp().Msg()
	}

	return obj
}

// Tcp returns a DataflowFlowProfileL4ProtocolTcp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) Tcp() DataflowFlowProfileL4ProtocolTcp {
	if obj.obj.Tcp == nil {
		obj.SetChoice(DataflowFlowProfileTcpIpStackChoice.TCP)
	}
	if obj.tcpHolder == nil {
		obj.tcpHolder = &dataflowFlowProfileL4ProtocolTcp{obj: obj.obj.Tcp}
	}
	return obj.tcpHolder
}

// Tcp returns a DataflowFlowProfileL4ProtocolTcp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) HasTcp() bool {
	return obj.obj.Tcp != nil
}

// SetTcp sets the DataflowFlowProfileL4ProtocolTcp value in the DataflowFlowProfileTcpIpStack object
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) SetTcp(value DataflowFlowProfileL4ProtocolTcp) DataflowFlowProfileTcpIpStack {
	obj.SetChoice(DataflowFlowProfileTcpIpStackChoice.TCP)
	obj.tcpHolder = nil
	obj.obj.Tcp = value.Msg()

	return obj
}

// Udp returns a DataflowFlowProfileL4ProtocolUdp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) Udp() DataflowFlowProfileL4ProtocolUdp {
	if obj.obj.Udp == nil {
		obj.SetChoice(DataflowFlowProfileTcpIpStackChoice.UDP)
	}
	if obj.udpHolder == nil {
		obj.udpHolder = &dataflowFlowProfileL4ProtocolUdp{obj: obj.obj.Udp}
	}
	return obj.udpHolder
}

// Udp returns a DataflowFlowProfileL4ProtocolUdp
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) HasUdp() bool {
	return obj.obj.Udp != nil
}

// SetUdp sets the DataflowFlowProfileL4ProtocolUdp value in the DataflowFlowProfileTcpIpStack object
// description is TBD
func (obj *dataflowFlowProfileTcpIpStack) SetUdp(value DataflowFlowProfileL4ProtocolUdp) DataflowFlowProfileTcpIpStack {
	obj.SetChoice(DataflowFlowProfileTcpIpStackChoice.UDP)
	obj.udpHolder = nil
	obj.obj.Udp = value.Msg()

	return obj
}

func (obj *dataflowFlowProfileTcpIpStack) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Ip != nil {
		obj.Ip().validateObj(set_default)
	}

	if obj.obj.Tcp != nil {
		obj.Tcp().validateObj(set_default)
	}

	if obj.obj.Udp != nil {
		obj.Udp().validateObj(set_default)
	}

}

func (obj *dataflowFlowProfileTcpIpStack) setDefault() {

}

// ***** MetricsResponseFlowResultTcpInfo *****
type metricsResponseFlowResultTcpInfo struct {
	obj *onexdataflowapi.MetricsResponseFlowResultTcpInfo
}

func NewMetricsResponseFlowResultTcpInfo() MetricsResponseFlowResultTcpInfo {
	obj := metricsResponseFlowResultTcpInfo{obj: &onexdataflowapi.MetricsResponseFlowResultTcpInfo{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponseFlowResultTcpInfo) Msg() *onexdataflowapi.MetricsResponseFlowResultTcpInfo {
	return obj.obj
}

func (obj *metricsResponseFlowResultTcpInfo) SetMsg(msg *onexdataflowapi.MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResultTcpInfo {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponseFlowResultTcpInfo) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *metricsResponseFlowResultTcpInfo) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsResponseFlowResultTcpInfo) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponseFlowResultTcpInfo) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *metricsResponseFlowResultTcpInfo) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *metricsResponseFlowResultTcpInfo) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *metricsResponseFlowResultTcpInfo) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *metricsResponseFlowResultTcpInfo) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *metricsResponseFlowResultTcpInfo) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// MetricsResponseFlowResultTcpInfo is tCP information for this flow
type MetricsResponseFlowResultTcpInfo interface {
	Msg() *onexdataflowapi.MetricsResponseFlowResultTcpInfo
	SetMsg(*onexdataflowapi.MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResultTcpInfo
	// ToPbText marshals MetricsResponseFlowResultTcpInfo to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals MetricsResponseFlowResultTcpInfo to YAML text
	ToYaml() (string, error)
	// ToJson marshals MetricsResponseFlowResultTcpInfo to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals MetricsResponseFlowResultTcpInfo from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponseFlowResultTcpInfo from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponseFlowResultTcpInfo from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponseFlowResultTcpInfo
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Rtt returns int64, set in MetricsResponseFlowResultTcpInfo.
	Rtt() int64
	// SetRtt assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetRtt(value int64) MetricsResponseFlowResultTcpInfo
	// HasRtt checks if Rtt has been set in MetricsResponseFlowResultTcpInfo
	HasRtt() bool
	// RttVariance returns int64, set in MetricsResponseFlowResultTcpInfo.
	RttVariance() int64
	// SetRttVariance assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetRttVariance(value int64) MetricsResponseFlowResultTcpInfo
	// HasRttVariance checks if RttVariance has been set in MetricsResponseFlowResultTcpInfo
	HasRttVariance() bool
	// Retransmissions returns int64, set in MetricsResponseFlowResultTcpInfo.
	Retransmissions() int64
	// SetRetransmissions assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetRetransmissions(value int64) MetricsResponseFlowResultTcpInfo
	// HasRetransmissions checks if Retransmissions has been set in MetricsResponseFlowResultTcpInfo
	HasRetransmissions() bool
	// RetransmissionTimeout returns int64, set in MetricsResponseFlowResultTcpInfo.
	RetransmissionTimeout() int64
	// SetRetransmissionTimeout assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetRetransmissionTimeout(value int64) MetricsResponseFlowResultTcpInfo
	// HasRetransmissionTimeout checks if RetransmissionTimeout has been set in MetricsResponseFlowResultTcpInfo
	HasRetransmissionTimeout() bool
	// CongestionWindow returns int64, set in MetricsResponseFlowResultTcpInfo.
	CongestionWindow() int64
	// SetCongestionWindow assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetCongestionWindow(value int64) MetricsResponseFlowResultTcpInfo
	// HasCongestionWindow checks if CongestionWindow has been set in MetricsResponseFlowResultTcpInfo
	HasCongestionWindow() bool
	// SlowStartThreshold returns int64, set in MetricsResponseFlowResultTcpInfo.
	SlowStartThreshold() int64
	// SetSlowStartThreshold assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetSlowStartThreshold(value int64) MetricsResponseFlowResultTcpInfo
	// HasSlowStartThreshold checks if SlowStartThreshold has been set in MetricsResponseFlowResultTcpInfo
	HasSlowStartThreshold() bool
	// PathMtu returns int64, set in MetricsResponseFlowResultTcpInfo.
	PathMtu() int64
	// SetPathMtu assigns int64 provided by user to MetricsResponseFlowResultTcpInfo
	SetPathMtu(value int64) MetricsResponseFlowResultTcpInfo
	// HasPathMtu checks if PathMtu has been set in MetricsResponseFlowResultTcpInfo
	HasPathMtu() bool
}

// Rtt returns a int64
// average round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) Rtt() int64 {

	return *obj.obj.Rtt

}

// Rtt returns a int64
// average round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) HasRtt() bool {
	return obj.obj.Rtt != nil
}

// SetRtt sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// average round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) SetRtt(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.Rtt = &value
	return obj
}

// RttVariance returns a int64
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) RttVariance() int64 {

	return *obj.obj.RttVariance

}

// RttVariance returns a int64
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) HasRttVariance() bool {
	return obj.obj.RttVariance != nil
}

// SetRttVariance sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) SetRttVariance(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.RttVariance = &value
	return obj
}

// Retransmissions returns a int64
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) Retransmissions() int64 {

	return *obj.obj.Retransmissions

}

// Retransmissions returns a int64
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) HasRetransmissions() bool {
	return obj.obj.Retransmissions != nil
}

// SetRetransmissions sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) SetRetransmissions(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.Retransmissions = &value
	return obj
}

// RetransmissionTimeout returns a int64
// retransmission timeout in micro seconds
func (obj *metricsResponseFlowResultTcpInfo) RetransmissionTimeout() int64 {

	return *obj.obj.RetransmissionTimeout

}

// RetransmissionTimeout returns a int64
// retransmission timeout in micro seconds
func (obj *metricsResponseFlowResultTcpInfo) HasRetransmissionTimeout() bool {
	return obj.obj.RetransmissionTimeout != nil
}

// SetRetransmissionTimeout sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// retransmission timeout in micro seconds
func (obj *metricsResponseFlowResultTcpInfo) SetRetransmissionTimeout(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.RetransmissionTimeout = &value
	return obj
}

// CongestionWindow returns a int64
// congestion windows size in bytes
func (obj *metricsResponseFlowResultTcpInfo) CongestionWindow() int64 {

	return *obj.obj.CongestionWindow

}

// CongestionWindow returns a int64
// congestion windows size in bytes
func (obj *metricsResponseFlowResultTcpInfo) HasCongestionWindow() bool {
	return obj.obj.CongestionWindow != nil
}

// SetCongestionWindow sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// congestion windows size in bytes
func (obj *metricsResponseFlowResultTcpInfo) SetCongestionWindow(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.CongestionWindow = &value
	return obj
}

// SlowStartThreshold returns a int64
// slow start threshold in bytes (max int64 value when wide open)
func (obj *metricsResponseFlowResultTcpInfo) SlowStartThreshold() int64 {

	return *obj.obj.SlowStartThreshold

}

// SlowStartThreshold returns a int64
// slow start threshold in bytes (max int64 value when wide open)
func (obj *metricsResponseFlowResultTcpInfo) HasSlowStartThreshold() bool {
	return obj.obj.SlowStartThreshold != nil
}

// SetSlowStartThreshold sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// slow start threshold in bytes (max int64 value when wide open)
func (obj *metricsResponseFlowResultTcpInfo) SetSlowStartThreshold(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.SlowStartThreshold = &value
	return obj
}

// PathMtu returns a int64
// path MTU
func (obj *metricsResponseFlowResultTcpInfo) PathMtu() int64 {

	return *obj.obj.PathMtu

}

// PathMtu returns a int64
// path MTU
func (obj *metricsResponseFlowResultTcpInfo) HasPathMtu() bool {
	return obj.obj.PathMtu != nil
}

// SetPathMtu sets the int64 value in the MetricsResponseFlowResultTcpInfo object
// path MTU
func (obj *metricsResponseFlowResultTcpInfo) SetPathMtu(value int64) MetricsResponseFlowResultTcpInfo {

	obj.obj.PathMtu = &value
	return obj
}

func (obj *metricsResponseFlowResultTcpInfo) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *metricsResponseFlowResultTcpInfo) setDefault() {

}

// ***** DataflowSimulatedComputeWorkload *****
type dataflowSimulatedComputeWorkload struct {
	obj *onexdataflowapi.DataflowSimulatedComputeWorkload
}

func NewDataflowSimulatedComputeWorkload() DataflowSimulatedComputeWorkload {
	obj := dataflowSimulatedComputeWorkload{obj: &onexdataflowapi.DataflowSimulatedComputeWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowSimulatedComputeWorkload) Msg() *onexdataflowapi.DataflowSimulatedComputeWorkload {
	return obj.obj
}

func (obj *dataflowSimulatedComputeWorkload) SetMsg(msg *onexdataflowapi.DataflowSimulatedComputeWorkload) DataflowSimulatedComputeWorkload {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowSimulatedComputeWorkload) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowSimulatedComputeWorkload) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowSimulatedComputeWorkload) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowSimulatedComputeWorkload) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowSimulatedComputeWorkload) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowSimulatedComputeWorkload) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowSimulatedComputeWorkload) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowSimulatedComputeWorkload) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowSimulatedComputeWorkload) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowSimulatedComputeWorkload is description is TBD
type DataflowSimulatedComputeWorkload interface {
	Msg() *onexdataflowapi.DataflowSimulatedComputeWorkload
	SetMsg(*onexdataflowapi.DataflowSimulatedComputeWorkload) DataflowSimulatedComputeWorkload
	// ToPbText marshals DataflowSimulatedComputeWorkload to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowSimulatedComputeWorkload to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowSimulatedComputeWorkload to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowSimulatedComputeWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowSimulatedComputeWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowSimulatedComputeWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowSimulatedComputeWorkload
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Duration returns float32, set in DataflowSimulatedComputeWorkload.
	Duration() float32
	// SetDuration assigns float32 provided by user to DataflowSimulatedComputeWorkload
	SetDuration(value float32) DataflowSimulatedComputeWorkload
	// HasDuration checks if Duration has been set in DataflowSimulatedComputeWorkload
	HasDuration() bool
}

// Duration returns a float32
// duration of the simulated compute workload in seconds
func (obj *dataflowSimulatedComputeWorkload) Duration() float32 {

	return *obj.obj.Duration

}

// Duration returns a float32
// duration of the simulated compute workload in seconds
func (obj *dataflowSimulatedComputeWorkload) HasDuration() bool {
	return obj.obj.Duration != nil
}

// SetDuration sets the float32 value in the DataflowSimulatedComputeWorkload object
// duration of the simulated compute workload in seconds
func (obj *dataflowSimulatedComputeWorkload) SetDuration(value float32) DataflowSimulatedComputeWorkload {

	obj.obj.Duration = &value
	return obj
}

func (obj *dataflowSimulatedComputeWorkload) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowSimulatedComputeWorkload) setDefault() {

}

// ***** DataflowFlowProfileRdmaStackRoceV2 *****
type dataflowFlowProfileRdmaStackRoceV2 struct {
	obj *onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2
}

func NewDataflowFlowProfileRdmaStackRoceV2() DataflowFlowProfileRdmaStackRoceV2 {
	obj := dataflowFlowProfileRdmaStackRoceV2{obj: &onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) Msg() *onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2 {
	return obj.obj
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) SetMsg(msg *onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2) DataflowFlowProfileRdmaStackRoceV2 {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowFlowProfileRdmaStackRoceV2 is description is TBD
type DataflowFlowProfileRdmaStackRoceV2 interface {
	Msg() *onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2
	SetMsg(*onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2) DataflowFlowProfileRdmaStackRoceV2
	// ToPbText marshals DataflowFlowProfileRdmaStackRoceV2 to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileRdmaStackRoceV2 to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileRdmaStackRoceV2 to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileRdmaStackRoceV2 from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileRdmaStackRoceV2 from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileRdmaStackRoceV2 from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileRdmaStackRoceV2
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Verb returns DataflowFlowProfileRdmaStackRoceV2VerbEnum, set in DataflowFlowProfileRdmaStackRoceV2
	Verb() DataflowFlowProfileRdmaStackRoceV2VerbEnum
	// SetVerb assigns DataflowFlowProfileRdmaStackRoceV2VerbEnum provided by user to DataflowFlowProfileRdmaStackRoceV2
	SetVerb(value DataflowFlowProfileRdmaStackRoceV2VerbEnum) DataflowFlowProfileRdmaStackRoceV2
	// HasVerb checks if Verb has been set in DataflowFlowProfileRdmaStackRoceV2
	HasVerb() bool
}

type DataflowFlowProfileRdmaStackRoceV2VerbEnum string

//  Enum of Verb on DataflowFlowProfileRdmaStackRoceV2
var DataflowFlowProfileRdmaStackRoceV2Verb = struct {
	WRITE DataflowFlowProfileRdmaStackRoceV2VerbEnum
	READ  DataflowFlowProfileRdmaStackRoceV2VerbEnum
}{
	WRITE: DataflowFlowProfileRdmaStackRoceV2VerbEnum("write"),
	READ:  DataflowFlowProfileRdmaStackRoceV2VerbEnum("read"),
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) Verb() DataflowFlowProfileRdmaStackRoceV2VerbEnum {
	return DataflowFlowProfileRdmaStackRoceV2VerbEnum(obj.obj.Verb.Enum().String())
}

// Verb returns a string
// read or write command
func (obj *dataflowFlowProfileRdmaStackRoceV2) HasVerb() bool {
	return obj.obj.Verb != nil
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) SetVerb(value DataflowFlowProfileRdmaStackRoceV2VerbEnum) DataflowFlowProfileRdmaStackRoceV2 {
	intValue, ok := onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2_Verb_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileRdmaStackRoceV2VerbEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfileRdmaStackRoceV2_Verb_Enum(intValue)
	obj.obj.Verb = &enumValue

	return obj
}

func (obj *dataflowFlowProfileRdmaStackRoceV2) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowFlowProfileRdmaStackRoceV2) setDefault() {
	if obj.obj.Verb == nil {
		obj.SetVerb(DataflowFlowProfileRdmaStackRoceV2Verb.WRITE)

	}

}

// ***** DataflowFlowProfileTcpIpStackIp *****
type dataflowFlowProfileTcpIpStackIp struct {
	obj *onexdataflowapi.DataflowFlowProfileTcpIpStackIp
}

func NewDataflowFlowProfileTcpIpStackIp() DataflowFlowProfileTcpIpStackIp {
	obj := dataflowFlowProfileTcpIpStackIp{obj: &onexdataflowapi.DataflowFlowProfileTcpIpStackIp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileTcpIpStackIp) Msg() *onexdataflowapi.DataflowFlowProfileTcpIpStackIp {
	return obj.obj
}

func (obj *dataflowFlowProfileTcpIpStackIp) SetMsg(msg *onexdataflowapi.DataflowFlowProfileTcpIpStackIp) DataflowFlowProfileTcpIpStackIp {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileTcpIpStackIp) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileTcpIpStackIp) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileTcpIpStackIp) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileTcpIpStackIp) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileTcpIpStackIp) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileTcpIpStackIp) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileTcpIpStackIp) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileTcpIpStackIp) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileTcpIpStackIp) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowFlowProfileTcpIpStackIp is description is TBD
type DataflowFlowProfileTcpIpStackIp interface {
	Msg() *onexdataflowapi.DataflowFlowProfileTcpIpStackIp
	SetMsg(*onexdataflowapi.DataflowFlowProfileTcpIpStackIp) DataflowFlowProfileTcpIpStackIp
	// ToPbText marshals DataflowFlowProfileTcpIpStackIp to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileTcpIpStackIp to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileTcpIpStackIp to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileTcpIpStackIp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileTcpIpStackIp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileTcpIpStackIp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileTcpIpStackIp
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Dscp returns int32, set in DataflowFlowProfileTcpIpStackIp.
	Dscp() int32
	// SetDscp assigns int32 provided by user to DataflowFlowProfileTcpIpStackIp
	SetDscp(value int32) DataflowFlowProfileTcpIpStackIp
	// HasDscp checks if Dscp has been set in DataflowFlowProfileTcpIpStackIp
	HasDscp() bool
}

// Dscp returns a int32
// differentiated services code point
func (obj *dataflowFlowProfileTcpIpStackIp) Dscp() int32 {

	return *obj.obj.Dscp

}

// Dscp returns a int32
// differentiated services code point
func (obj *dataflowFlowProfileTcpIpStackIp) HasDscp() bool {
	return obj.obj.Dscp != nil
}

// SetDscp sets the int32 value in the DataflowFlowProfileTcpIpStackIp object
// differentiated services code point
func (obj *dataflowFlowProfileTcpIpStackIp) SetDscp(value int32) DataflowFlowProfileTcpIpStackIp {

	obj.obj.Dscp = &value
	return obj
}

func (obj *dataflowFlowProfileTcpIpStackIp) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowFlowProfileTcpIpStackIp) setDefault() {

}

// ***** DataflowFlowProfileL4ProtocolTcp *****
type dataflowFlowProfileL4ProtocolTcp struct {
	obj                   *onexdataflowapi.DataflowFlowProfileL4ProtocolTcp
	destinationPortHolder L4PortRange
	sourcePortHolder      L4PortRange
}

func NewDataflowFlowProfileL4ProtocolTcp() DataflowFlowProfileL4ProtocolTcp {
	obj := dataflowFlowProfileL4ProtocolTcp{obj: &onexdataflowapi.DataflowFlowProfileL4ProtocolTcp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileL4ProtocolTcp) Msg() *onexdataflowapi.DataflowFlowProfileL4ProtocolTcp {
	return obj.obj
}

func (obj *dataflowFlowProfileL4ProtocolTcp) SetMsg(msg *onexdataflowapi.DataflowFlowProfileL4ProtocolTcp) DataflowFlowProfileL4ProtocolTcp {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileL4ProtocolTcp) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileL4ProtocolTcp) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileL4ProtocolTcp) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileL4ProtocolTcp) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowFlowProfileL4ProtocolTcp) setNil() {
	obj.destinationPortHolder = nil
	obj.sourcePortHolder = nil
}

// DataflowFlowProfileL4ProtocolTcp is description is TBD
type DataflowFlowProfileL4ProtocolTcp interface {
	Msg() *onexdataflowapi.DataflowFlowProfileL4ProtocolTcp
	SetMsg(*onexdataflowapi.DataflowFlowProfileL4ProtocolTcp) DataflowFlowProfileL4ProtocolTcp
	// ToPbText marshals DataflowFlowProfileL4ProtocolTcp to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileL4ProtocolTcp to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileL4ProtocolTcp to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileL4ProtocolTcp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileL4ProtocolTcp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileL4ProtocolTcp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileL4ProtocolTcp
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// CongestionAlgorithm returns DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum, set in DataflowFlowProfileL4ProtocolTcp
	CongestionAlgorithm() DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum
	// SetCongestionAlgorithm assigns DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum provided by user to DataflowFlowProfileL4ProtocolTcp
	SetCongestionAlgorithm(value DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum) DataflowFlowProfileL4ProtocolTcp
	// HasCongestionAlgorithm checks if CongestionAlgorithm has been set in DataflowFlowProfileL4ProtocolTcp
	HasCongestionAlgorithm() bool
	// Initcwnd returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	Initcwnd() int32
	// SetInitcwnd assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetInitcwnd(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasInitcwnd checks if Initcwnd has been set in DataflowFlowProfileL4ProtocolTcp
	HasInitcwnd() bool
	// SendBuf returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	SendBuf() int32
	// SetSendBuf assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetSendBuf(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasSendBuf checks if SendBuf has been set in DataflowFlowProfileL4ProtocolTcp
	HasSendBuf() bool
	// ReceiveBuf returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	ReceiveBuf() int32
	// SetReceiveBuf assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetReceiveBuf(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasReceiveBuf checks if ReceiveBuf has been set in DataflowFlowProfileL4ProtocolTcp
	HasReceiveBuf() bool
	// DelayedAck returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	DelayedAck() int32
	// SetDelayedAck assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetDelayedAck(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasDelayedAck checks if DelayedAck has been set in DataflowFlowProfileL4ProtocolTcp
	HasDelayedAck() bool
	// SelectiveAck returns bool, set in DataflowFlowProfileL4ProtocolTcp.
	SelectiveAck() bool
	// SetSelectiveAck assigns bool provided by user to DataflowFlowProfileL4ProtocolTcp
	SetSelectiveAck(value bool) DataflowFlowProfileL4ProtocolTcp
	// HasSelectiveAck checks if SelectiveAck has been set in DataflowFlowProfileL4ProtocolTcp
	HasSelectiveAck() bool
	// MinRto returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	MinRto() int32
	// SetMinRto assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetMinRto(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasMinRto checks if MinRto has been set in DataflowFlowProfileL4ProtocolTcp
	HasMinRto() bool
	// Mss returns int32, set in DataflowFlowProfileL4ProtocolTcp.
	Mss() int32
	// SetMss assigns int32 provided by user to DataflowFlowProfileL4ProtocolTcp
	SetMss(value int32) DataflowFlowProfileL4ProtocolTcp
	// HasMss checks if Mss has been set in DataflowFlowProfileL4ProtocolTcp
	HasMss() bool
	// Ecn returns bool, set in DataflowFlowProfileL4ProtocolTcp.
	Ecn() bool
	// SetEcn assigns bool provided by user to DataflowFlowProfileL4ProtocolTcp
	SetEcn(value bool) DataflowFlowProfileL4ProtocolTcp
	// HasEcn checks if Ecn has been set in DataflowFlowProfileL4ProtocolTcp
	HasEcn() bool
	// EnableTimestamp returns bool, set in DataflowFlowProfileL4ProtocolTcp.
	EnableTimestamp() bool
	// SetEnableTimestamp assigns bool provided by user to DataflowFlowProfileL4ProtocolTcp
	SetEnableTimestamp(value bool) DataflowFlowProfileL4ProtocolTcp
	// HasEnableTimestamp checks if EnableTimestamp has been set in DataflowFlowProfileL4ProtocolTcp
	HasEnableTimestamp() bool
	// DestinationPort returns L4PortRange, set in DataflowFlowProfileL4ProtocolTcp.
	// L4PortRange is layer4 protocol source or destination port values
	DestinationPort() L4PortRange
	// SetDestinationPort assigns L4PortRange provided by user to DataflowFlowProfileL4ProtocolTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SetDestinationPort(value L4PortRange) DataflowFlowProfileL4ProtocolTcp
	// HasDestinationPort checks if DestinationPort has been set in DataflowFlowProfileL4ProtocolTcp
	HasDestinationPort() bool
	// SourcePort returns L4PortRange, set in DataflowFlowProfileL4ProtocolTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SourcePort() L4PortRange
	// SetSourcePort assigns L4PortRange provided by user to DataflowFlowProfileL4ProtocolTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SetSourcePort(value L4PortRange) DataflowFlowProfileL4ProtocolTcp
	// HasSourcePort checks if SourcePort has been set in DataflowFlowProfileL4ProtocolTcp
	HasSourcePort() bool
	setNil()
}

type DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum string

//  Enum of CongestionAlgorithm on DataflowFlowProfileL4ProtocolTcp
var DataflowFlowProfileL4ProtocolTcpCongestionAlgorithm = struct {
	BBR   DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum
	DCTCP DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum
	CUBIC DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum
	RENO  DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum
}{
	BBR:   DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum("bbr"),
	DCTCP: DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum("dctcp"),
	CUBIC: DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum("cubic"),
	RENO:  DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum("reno"),
}

func (obj *dataflowFlowProfileL4ProtocolTcp) CongestionAlgorithm() DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum {
	return DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum(obj.obj.CongestionAlgorithm.Enum().String())
}

// CongestionAlgorithm returns a string
// The TCP congestion algorithm:
// bbr - Bottleneck Bandwidth and Round-trip propagation time
// dctcp - Data center TCP
// cubic - cubic window increase function
// reno - TCP New Reno
func (obj *dataflowFlowProfileL4ProtocolTcp) HasCongestionAlgorithm() bool {
	return obj.obj.CongestionAlgorithm != nil
}

func (obj *dataflowFlowProfileL4ProtocolTcp) SetCongestionAlgorithm(value DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum) DataflowFlowProfileL4ProtocolTcp {
	intValue, ok := onexdataflowapi.DataflowFlowProfileL4ProtocolTcp_CongestionAlgorithm_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileL4ProtocolTcpCongestionAlgorithmEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfileL4ProtocolTcp_CongestionAlgorithm_Enum(intValue)
	obj.obj.CongestionAlgorithm = &enumValue

	return obj
}

// Initcwnd returns a int32
// initial congestion window
func (obj *dataflowFlowProfileL4ProtocolTcp) Initcwnd() int32 {

	return *obj.obj.Initcwnd

}

// Initcwnd returns a int32
// initial congestion window
func (obj *dataflowFlowProfileL4ProtocolTcp) HasInitcwnd() bool {
	return obj.obj.Initcwnd != nil
}

// SetInitcwnd sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// initial congestion window
func (obj *dataflowFlowProfileL4ProtocolTcp) SetInitcwnd(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.Initcwnd = &value
	return obj
}

// SendBuf returns a int32
// send buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) SendBuf() int32 {

	return *obj.obj.SendBuf

}

// SendBuf returns a int32
// send buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) HasSendBuf() bool {
	return obj.obj.SendBuf != nil
}

// SetSendBuf sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// send buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) SetSendBuf(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.SendBuf = &value
	return obj
}

// ReceiveBuf returns a int32
// receive buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) ReceiveBuf() int32 {

	return *obj.obj.ReceiveBuf

}

// ReceiveBuf returns a int32
// receive buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) HasReceiveBuf() bool {
	return obj.obj.ReceiveBuf != nil
}

// SetReceiveBuf sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// receive buffer size
func (obj *dataflowFlowProfileL4ProtocolTcp) SetReceiveBuf(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.ReceiveBuf = &value
	return obj
}

// DelayedAck returns a int32
// delayed acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) DelayedAck() int32 {

	return *obj.obj.DelayedAck

}

// DelayedAck returns a int32
// delayed acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) HasDelayedAck() bool {
	return obj.obj.DelayedAck != nil
}

// SetDelayedAck sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// delayed acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) SetDelayedAck(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.DelayedAck = &value
	return obj
}

// SelectiveAck returns a bool
// selective acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) SelectiveAck() bool {

	return *obj.obj.SelectiveAck

}

// SelectiveAck returns a bool
// selective acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) HasSelectiveAck() bool {
	return obj.obj.SelectiveAck != nil
}

// SetSelectiveAck sets the bool value in the DataflowFlowProfileL4ProtocolTcp object
// selective acknowledgment
func (obj *dataflowFlowProfileL4ProtocolTcp) SetSelectiveAck(value bool) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.SelectiveAck = &value
	return obj
}

// MinRto returns a int32
// minimum retransmission timeout
func (obj *dataflowFlowProfileL4ProtocolTcp) MinRto() int32 {

	return *obj.obj.MinRto

}

// MinRto returns a int32
// minimum retransmission timeout
func (obj *dataflowFlowProfileL4ProtocolTcp) HasMinRto() bool {
	return obj.obj.MinRto != nil
}

// SetMinRto sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// minimum retransmission timeout
func (obj *dataflowFlowProfileL4ProtocolTcp) SetMinRto(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.MinRto = &value
	return obj
}

// Mss returns a int32
// Maximum Segment Size
func (obj *dataflowFlowProfileL4ProtocolTcp) Mss() int32 {

	return *obj.obj.Mss

}

// Mss returns a int32
// Maximum Segment Size
func (obj *dataflowFlowProfileL4ProtocolTcp) HasMss() bool {
	return obj.obj.Mss != nil
}

// SetMss sets the int32 value in the DataflowFlowProfileL4ProtocolTcp object
// Maximum Segment Size
func (obj *dataflowFlowProfileL4ProtocolTcp) SetMss(value int32) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.Mss = &value
	return obj
}

// Ecn returns a bool
// early congestion notification
func (obj *dataflowFlowProfileL4ProtocolTcp) Ecn() bool {

	return *obj.obj.Ecn

}

// Ecn returns a bool
// early congestion notification
func (obj *dataflowFlowProfileL4ProtocolTcp) HasEcn() bool {
	return obj.obj.Ecn != nil
}

// SetEcn sets the bool value in the DataflowFlowProfileL4ProtocolTcp object
// early congestion notification
func (obj *dataflowFlowProfileL4ProtocolTcp) SetEcn(value bool) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.Ecn = &value
	return obj
}

// EnableTimestamp returns a bool
// enable tcp timestamping
func (obj *dataflowFlowProfileL4ProtocolTcp) EnableTimestamp() bool {

	return *obj.obj.EnableTimestamp

}

// EnableTimestamp returns a bool
// enable tcp timestamping
func (obj *dataflowFlowProfileL4ProtocolTcp) HasEnableTimestamp() bool {
	return obj.obj.EnableTimestamp != nil
}

// SetEnableTimestamp sets the bool value in the DataflowFlowProfileL4ProtocolTcp object
// enable tcp timestamping
func (obj *dataflowFlowProfileL4ProtocolTcp) SetEnableTimestamp(value bool) DataflowFlowProfileL4ProtocolTcp {

	obj.obj.EnableTimestamp = &value
	return obj
}

// DestinationPort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) DestinationPort() L4PortRange {
	if obj.obj.DestinationPort == nil {
		obj.obj.DestinationPort = NewL4PortRange().Msg()
	}
	if obj.destinationPortHolder == nil {
		obj.destinationPortHolder = &l4PortRange{obj: obj.obj.DestinationPort}
	}
	return obj.destinationPortHolder
}

// DestinationPort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) HasDestinationPort() bool {
	return obj.obj.DestinationPort != nil
}

// SetDestinationPort sets the L4PortRange value in the DataflowFlowProfileL4ProtocolTcp object
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) SetDestinationPort(value L4PortRange) DataflowFlowProfileL4ProtocolTcp {

	obj.destinationPortHolder = nil
	obj.obj.DestinationPort = value.Msg()

	return obj
}

// SourcePort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) SourcePort() L4PortRange {
	if obj.obj.SourcePort == nil {
		obj.obj.SourcePort = NewL4PortRange().Msg()
	}
	if obj.sourcePortHolder == nil {
		obj.sourcePortHolder = &l4PortRange{obj: obj.obj.SourcePort}
	}
	return obj.sourcePortHolder
}

// SourcePort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) HasSourcePort() bool {
	return obj.obj.SourcePort != nil
}

// SetSourcePort sets the L4PortRange value in the DataflowFlowProfileL4ProtocolTcp object
// description is TBD
func (obj *dataflowFlowProfileL4ProtocolTcp) SetSourcePort(value L4PortRange) DataflowFlowProfileL4ProtocolTcp {

	obj.sourcePortHolder = nil
	obj.obj.SourcePort = value.Msg()

	return obj
}

func (obj *dataflowFlowProfileL4ProtocolTcp) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.DestinationPort != nil {
		obj.DestinationPort().validateObj(set_default)
	}

	if obj.obj.SourcePort != nil {
		obj.SourcePort().validateObj(set_default)
	}

}

func (obj *dataflowFlowProfileL4ProtocolTcp) setDefault() {
	if obj.obj.CongestionAlgorithm == nil {
		obj.SetCongestionAlgorithm(DataflowFlowProfileL4ProtocolTcpCongestionAlgorithm.CUBIC)

	}
	if obj.obj.Mss == nil {
		obj.SetMss(1500)
	}

}

// ***** DataflowFlowProfileL4ProtocolUdp *****
type dataflowFlowProfileL4ProtocolUdp struct {
	obj *onexdataflowapi.DataflowFlowProfileL4ProtocolUdp
}

func NewDataflowFlowProfileL4ProtocolUdp() DataflowFlowProfileL4ProtocolUdp {
	obj := dataflowFlowProfileL4ProtocolUdp{obj: &onexdataflowapi.DataflowFlowProfileL4ProtocolUdp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileL4ProtocolUdp) Msg() *onexdataflowapi.DataflowFlowProfileL4ProtocolUdp {
	return obj.obj
}

func (obj *dataflowFlowProfileL4ProtocolUdp) SetMsg(msg *onexdataflowapi.DataflowFlowProfileL4ProtocolUdp) DataflowFlowProfileL4ProtocolUdp {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileL4ProtocolUdp) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *dataflowFlowProfileL4ProtocolUdp) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileL4ProtocolUdp) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileL4ProtocolUdp) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *dataflowFlowProfileL4ProtocolUdp) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *dataflowFlowProfileL4ProtocolUdp) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *dataflowFlowProfileL4ProtocolUdp) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileL4ProtocolUdp) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileL4ProtocolUdp) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowFlowProfileL4ProtocolUdp is description is TBD
type DataflowFlowProfileL4ProtocolUdp interface {
	Msg() *onexdataflowapi.DataflowFlowProfileL4ProtocolUdp
	SetMsg(*onexdataflowapi.DataflowFlowProfileL4ProtocolUdp) DataflowFlowProfileL4ProtocolUdp
	// ToPbText marshals DataflowFlowProfileL4ProtocolUdp to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileL4ProtocolUdp to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileL4ProtocolUdp to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileL4ProtocolUdp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileL4ProtocolUdp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileL4ProtocolUdp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileL4ProtocolUdp
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *dataflowFlowProfileL4ProtocolUdp) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowFlowProfileL4ProtocolUdp) setDefault() {

}

// ***** L4PortRange *****
type l4PortRange struct {
	obj               *onexdataflowapi.L4PortRange
	singleValueHolder L4PortRangeSingleValue
	rangeHolder       L4PortRangeRange
}

func NewL4PortRange() L4PortRange {
	obj := l4PortRange{obj: &onexdataflowapi.L4PortRange{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRange) Msg() *onexdataflowapi.L4PortRange {
	return obj.obj
}

func (obj *l4PortRange) SetMsg(msg *onexdataflowapi.L4PortRange) L4PortRange {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRange) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *l4PortRange) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *l4PortRange) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRange) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *l4PortRange) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRange) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}
	obj.setNil()
	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *l4PortRange) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *l4PortRange) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *l4PortRange) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *l4PortRange) setNil() {
	obj.singleValueHolder = nil
	obj.rangeHolder = nil
}

// L4PortRange is layer4 protocol source or destination port values
type L4PortRange interface {
	Msg() *onexdataflowapi.L4PortRange
	SetMsg(*onexdataflowapi.L4PortRange) L4PortRange
	// ToPbText marshals L4PortRange to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals L4PortRange to YAML text
	ToYaml() (string, error)
	// ToJson marshals L4PortRange to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals L4PortRange from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRange from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRange from JSON text
	FromJson(value string) error
	// Validate validates L4PortRange
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Choice returns L4PortRangeChoiceEnum, set in L4PortRange
	Choice() L4PortRangeChoiceEnum
	// SetChoice assigns L4PortRangeChoiceEnum provided by user to L4PortRange
	SetChoice(value L4PortRangeChoiceEnum) L4PortRange
	// HasChoice checks if Choice has been set in L4PortRange
	HasChoice() bool
	// SingleValue returns L4PortRangeSingleValue, set in L4PortRange.
	// L4PortRangeSingleValue is description is TBD
	SingleValue() L4PortRangeSingleValue
	// SetSingleValue assigns L4PortRangeSingleValue provided by user to L4PortRange.
	// L4PortRangeSingleValue is description is TBD
	SetSingleValue(value L4PortRangeSingleValue) L4PortRange
	// HasSingleValue checks if SingleValue has been set in L4PortRange
	HasSingleValue() bool
	// Range returns L4PortRangeRange, set in L4PortRange.
	// L4PortRangeRange is description is TBD
	Range() L4PortRangeRange
	// SetRange assigns L4PortRangeRange provided by user to L4PortRange.
	// L4PortRangeRange is description is TBD
	SetRange(value L4PortRangeRange) L4PortRange
	// HasRange checks if Range has been set in L4PortRange
	HasRange() bool
	setNil()
}

type L4PortRangeChoiceEnum string

//  Enum of Choice on L4PortRange
var L4PortRangeChoice = struct {
	SINGLE_VALUE L4PortRangeChoiceEnum
	RANGE        L4PortRangeChoiceEnum
}{
	SINGLE_VALUE: L4PortRangeChoiceEnum("single_value"),
	RANGE:        L4PortRangeChoiceEnum("range"),
}

func (obj *l4PortRange) Choice() L4PortRangeChoiceEnum {
	return L4PortRangeChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// None
func (obj *l4PortRange) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *l4PortRange) SetChoice(value L4PortRangeChoiceEnum) L4PortRange {
	intValue, ok := onexdataflowapi.L4PortRange_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on L4PortRangeChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.L4PortRange_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Range = nil
	obj.rangeHolder = nil
	obj.obj.SingleValue = nil
	obj.singleValueHolder = nil

	if value == L4PortRangeChoice.SINGLE_VALUE {
		obj.obj.SingleValue = NewL4PortRangeSingleValue().Msg()
	}

	if value == L4PortRangeChoice.RANGE {
		obj.obj.Range = NewL4PortRangeRange().Msg()
	}

	return obj
}

// SingleValue returns a L4PortRangeSingleValue
// description is TBD
func (obj *l4PortRange) SingleValue() L4PortRangeSingleValue {
	if obj.obj.SingleValue == nil {
		obj.SetChoice(L4PortRangeChoice.SINGLE_VALUE)
	}
	if obj.singleValueHolder == nil {
		obj.singleValueHolder = &l4PortRangeSingleValue{obj: obj.obj.SingleValue}
	}
	return obj.singleValueHolder
}

// SingleValue returns a L4PortRangeSingleValue
// description is TBD
func (obj *l4PortRange) HasSingleValue() bool {
	return obj.obj.SingleValue != nil
}

// SetSingleValue sets the L4PortRangeSingleValue value in the L4PortRange object
// description is TBD
func (obj *l4PortRange) SetSingleValue(value L4PortRangeSingleValue) L4PortRange {
	obj.SetChoice(L4PortRangeChoice.SINGLE_VALUE)
	obj.singleValueHolder = nil
	obj.obj.SingleValue = value.Msg()

	return obj
}

// Range returns a L4PortRangeRange
// description is TBD
func (obj *l4PortRange) Range() L4PortRangeRange {
	if obj.obj.Range == nil {
		obj.SetChoice(L4PortRangeChoice.RANGE)
	}
	if obj.rangeHolder == nil {
		obj.rangeHolder = &l4PortRangeRange{obj: obj.obj.Range}
	}
	return obj.rangeHolder
}

// Range returns a L4PortRangeRange
// description is TBD
func (obj *l4PortRange) HasRange() bool {
	return obj.obj.Range != nil
}

// SetRange sets the L4PortRangeRange value in the L4PortRange object
// description is TBD
func (obj *l4PortRange) SetRange(value L4PortRangeRange) L4PortRange {
	obj.SetChoice(L4PortRangeChoice.RANGE)
	obj.rangeHolder = nil
	obj.obj.Range = value.Msg()

	return obj
}

func (obj *l4PortRange) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.SingleValue != nil {
		obj.SingleValue().validateObj(set_default)
	}

	if obj.obj.Range != nil {
		obj.Range().validateObj(set_default)
	}

}

func (obj *l4PortRange) setDefault() {

}

// ***** L4PortRangeSingleValue *****
type l4PortRangeSingleValue struct {
	obj *onexdataflowapi.L4PortRangeSingleValue
}

func NewL4PortRangeSingleValue() L4PortRangeSingleValue {
	obj := l4PortRangeSingleValue{obj: &onexdataflowapi.L4PortRangeSingleValue{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRangeSingleValue) Msg() *onexdataflowapi.L4PortRangeSingleValue {
	return obj.obj
}

func (obj *l4PortRangeSingleValue) SetMsg(msg *onexdataflowapi.L4PortRangeSingleValue) L4PortRangeSingleValue {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRangeSingleValue) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *l4PortRangeSingleValue) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *l4PortRangeSingleValue) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRangeSingleValue) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *l4PortRangeSingleValue) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRangeSingleValue) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *l4PortRangeSingleValue) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *l4PortRangeSingleValue) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *l4PortRangeSingleValue) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// L4PortRangeSingleValue is description is TBD
type L4PortRangeSingleValue interface {
	Msg() *onexdataflowapi.L4PortRangeSingleValue
	SetMsg(*onexdataflowapi.L4PortRangeSingleValue) L4PortRangeSingleValue
	// ToPbText marshals L4PortRangeSingleValue to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals L4PortRangeSingleValue to YAML text
	ToYaml() (string, error)
	// ToJson marshals L4PortRangeSingleValue to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals L4PortRangeSingleValue from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRangeSingleValue from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRangeSingleValue from JSON text
	FromJson(value string) error
	// Validate validates L4PortRangeSingleValue
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Value returns int32, set in L4PortRangeSingleValue.
	Value() int32
	// SetValue assigns int32 provided by user to L4PortRangeSingleValue
	SetValue(value int32) L4PortRangeSingleValue
	// HasValue checks if Value has been set in L4PortRangeSingleValue
	HasValue() bool
}

// Value returns a int32
// description is TBD
func (obj *l4PortRangeSingleValue) Value() int32 {

	return *obj.obj.Value

}

// Value returns a int32
// description is TBD
func (obj *l4PortRangeSingleValue) HasValue() bool {
	return obj.obj.Value != nil
}

// SetValue sets the int32 value in the L4PortRangeSingleValue object
// description is TBD
func (obj *l4PortRangeSingleValue) SetValue(value int32) L4PortRangeSingleValue {

	obj.obj.Value = &value
	return obj
}

func (obj *l4PortRangeSingleValue) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *l4PortRangeSingleValue) setDefault() {
	if obj.obj.Value == nil {
		obj.SetValue(1)
	}

}

// ***** L4PortRangeRange *****
type l4PortRangeRange struct {
	obj *onexdataflowapi.L4PortRangeRange
}

func NewL4PortRangeRange() L4PortRangeRange {
	obj := l4PortRangeRange{obj: &onexdataflowapi.L4PortRangeRange{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRangeRange) Msg() *onexdataflowapi.L4PortRangeRange {
	return obj.obj
}

func (obj *l4PortRangeRange) SetMsg(msg *onexdataflowapi.L4PortRangeRange) L4PortRangeRange {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRangeRange) ToPbText() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	protoMarshal, err := proto.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(protoMarshal), nil
}

func (obj *l4PortRangeRange) FromPbText(value string) error {
	retObj := proto.Unmarshal([]byte(value), obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *l4PortRangeRange) ToYaml() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRangeRange) FromYaml(value string) error {
	if value == "" {
		value = "{}"
	}
	data, err := yaml.YAMLToJSON([]byte(value))
	if err != nil {
		return err
	}
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	uError := opts.Unmarshal([]byte(data), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return nil
}

func (obj *l4PortRangeRange) ToJson() (string, error) {
	vErr := obj.Validate()
	if vErr != nil {
		return "", vErr
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (obj *l4PortRangeRange) FromJson(value string) error {
	opts := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}
	if value == "" {
		value = "{}"
	}
	uError := opts.Unmarshal([]byte(value), obj.Msg())
	if uError != nil {
		return fmt.Errorf("unmarshal error %s", strings.Replace(
			uError.Error(), "\u00a0", " ", -1)[7:])
	}

	err := obj.validateFromText()
	if err != nil {
		return err
	}
	return nil
}

func (obj *l4PortRangeRange) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *l4PortRangeRange) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *l4PortRangeRange) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// L4PortRangeRange is description is TBD
type L4PortRangeRange interface {
	Msg() *onexdataflowapi.L4PortRangeRange
	SetMsg(*onexdataflowapi.L4PortRangeRange) L4PortRangeRange
	// ToPbText marshals L4PortRangeRange to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals L4PortRangeRange to YAML text
	ToYaml() (string, error)
	// ToJson marshals L4PortRangeRange to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals L4PortRangeRange from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRangeRange from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRangeRange from JSON text
	FromJson(value string) error
	// Validate validates L4PortRangeRange
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StartValue returns int32, set in L4PortRangeRange.
	StartValue() int32
	// SetStartValue assigns int32 provided by user to L4PortRangeRange
	SetStartValue(value int32) L4PortRangeRange
	// HasStartValue checks if StartValue has been set in L4PortRangeRange
	HasStartValue() bool
	// Increment returns int32, set in L4PortRangeRange.
	Increment() int32
	// SetIncrement assigns int32 provided by user to L4PortRangeRange
	SetIncrement(value int32) L4PortRangeRange
	// HasIncrement checks if Increment has been set in L4PortRangeRange
	HasIncrement() bool
}

// StartValue returns a int32
// description is TBD
func (obj *l4PortRangeRange) StartValue() int32 {

	return *obj.obj.StartValue

}

// StartValue returns a int32
// description is TBD
func (obj *l4PortRangeRange) HasStartValue() bool {
	return obj.obj.StartValue != nil
}

// SetStartValue sets the int32 value in the L4PortRangeRange object
// description is TBD
func (obj *l4PortRangeRange) SetStartValue(value int32) L4PortRangeRange {

	obj.obj.StartValue = &value
	return obj
}

// Increment returns a int32
// description is TBD
func (obj *l4PortRangeRange) Increment() int32 {

	return *obj.obj.Increment

}

// Increment returns a int32
// description is TBD
func (obj *l4PortRangeRange) HasIncrement() bool {
	return obj.obj.Increment != nil
}

// SetIncrement sets the int32 value in the L4PortRangeRange object
// description is TBD
func (obj *l4PortRangeRange) SetIncrement(value int32) L4PortRangeRange {

	obj.obj.Increment = &value
	return obj
}

func (obj *l4PortRangeRange) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Increment != nil {
		if *obj.obj.Increment < 1 || *obj.obj.Increment > 2147483647 {
			validation = append(
				validation,
				fmt.Sprintf("1 <= L4PortRangeRange.Increment <= 2147483647 but Got %d", *obj.obj.Increment))
		}

	}

}

func (obj *l4PortRangeRange) setDefault() {
	if obj.obj.StartValue == nil {
		obj.SetStartValue(1)
	}
	if obj.obj.Increment == nil {
		obj.SetIncrement(1)
	}

}
