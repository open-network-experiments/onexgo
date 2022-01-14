// Distributed System Emulator Dataflow API and Data Models 0.0.1
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

// OnexgodataflowapiApi the Dataflow API and Data Models
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
	// NewGetMetricsResponse returns a new instance of GetMetricsResponse.
	// GetMetricsResponse is description is TBD
	NewGetMetricsResponse() GetMetricsResponse
	// SetConfig sets the ONEx dataflow config
	SetConfig(config Config) (Config, error)
	// GetConfig gets the ONEx dataflow config from the server, as currently configured
	GetConfig(getConfigDetails GetConfigDetails) (Config, error)
	// RunExperiment runs the currently configured experiment
	RunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error)
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
	resp, err := api.httpSendRecv("config", configJson, "POST")

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
	resp, err := api.httpSendRecv("config", getConfigDetailsJson, "GET")

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
	resp, err := api.httpSendRecv("control/experiment", experimentRequestJson, "POST")

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

func (api *onexgodataflowapiApi) httpGetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error) {
	metricsRequestJson, err := metricsRequest.ToJson()
	if err != nil {
		return nil, err
	}
	resp, err := api.httpSendRecv("results/metrics", metricsRequestJson, "POST")

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
	obj *onexdataflowapi.ErrorDetails
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
	// Errors returns []string, set in ErrorDetails.
	Errors() []string
	// SetErrors assigns []string provided by user to ErrorDetails
	SetErrors(value []string) ErrorDetails
}

// Errors returns a []string
// description is TBD
func (obj *errorDetails) Errors() []string {
	if obj.obj.Errors == nil {
		obj.obj.Errors = make([]string, 0)
	}
	return obj.obj.Errors
}

// SetErrors sets the []string value in the ErrorDetails object
// description is TBD
func (obj *errorDetails) SetErrors(value []string) ErrorDetails {

	if obj.obj.Errors == nil {
		obj.obj.Errors = make([]string, 0)
	}
	obj.obj.Errors = value

	return obj
}

func (obj *errorDetails) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *errorDetails) setDefault() {

}

// ***** WarningDetails *****
type warningDetails struct {
	obj *onexdataflowapi.WarningDetails
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
	// Warnings returns []string, set in WarningDetails.
	Warnings() []string
	// SetWarnings assigns []string provided by user to WarningDetails
	SetWarnings(value []string) WarningDetails
}

// Warnings returns a []string
// description is TBD
func (obj *warningDetails) Warnings() []string {
	if obj.obj.Warnings == nil {
		obj.obj.Warnings = make([]string, 0)
	}
	return obj.obj.Warnings
}

// SetWarnings sets the []string value in the WarningDetails object
// description is TBD
func (obj *warningDetails) SetWarnings(value []string) WarningDetails {

	if obj.obj.Warnings == nil {
		obj.obj.Warnings = make([]string, 0)
	}
	obj.obj.Warnings = value

	return obj
}

func (obj *warningDetails) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *warningDetails) setDefault() {

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
	// Jct returns float32, set in MetricsResponse.
	Jct() float32
	// SetJct assigns float32 provided by user to MetricsResponse
	SetJct(value float32) MetricsResponse
	// HasJct checks if Jct has been set in MetricsResponse
	HasJct() bool
	// FlowResults returns MetricsResponseMetricsResponseFlowResultIter, set in MetricsResponse
	FlowResults() MetricsResponseMetricsResponseFlowResultIter
	setNil()
}

// Jct returns a float32
// job completion time in micro seconds
func (obj *metricsResponse) Jct() float32 {

	return *obj.obj.Jct

}

// Jct returns a float32
// job completion time in micro seconds
func (obj *metricsResponse) HasJct() bool {
	return obj.obj.Jct != nil
}

// SetJct sets the float32 value in the MetricsResponse object
// job completion time in micro seconds
func (obj *metricsResponse) SetJct(value float32) MetricsResponse {

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
	// ManagementAddress returns string, set in DataflowHostManagement.
	ManagementAddress() string
	// SetManagementAddress assigns string provided by user to DataflowHostManagement
	SetManagementAddress(value string) DataflowHostManagement
	// NicName returns string, set in DataflowHostManagement.
	NicName() string
	// SetNicName assigns string provided by user to DataflowHostManagement
	SetNicName(value string) DataflowHostManagement
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

// ManagementAddress returns a string
// Hostname or address of management interface of a server running dataflow traffic
func (obj *dataflowHostManagement) ManagementAddress() string {

	return obj.obj.ManagementAddress
}

// SetManagementAddress sets the string value in the DataflowHostManagement object
// Hostname or address of management interface of a server running dataflow traffic
func (obj *dataflowHostManagement) SetManagementAddress(value string) DataflowHostManagement {

	obj.obj.ManagementAddress = value
	return obj
}

// NicName returns a string
// unique idenfier for the network interface card (nic), e.g. "eth1"
func (obj *dataflowHostManagement) NicName() string {

	return obj.obj.NicName
}

// SetNicName sets the string value in the DataflowHostManagement object
// unique idenfier for the network interface card (nic), e.g. "eth1"
func (obj *dataflowHostManagement) SetNicName(value string) DataflowHostManagement {

	obj.obj.NicName = value
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

	// ManagementAddress is required
	if obj.obj.ManagementAddress == "" {
		validation = append(validation, "ManagementAddress is required field on interface DataflowHostManagement")
	}

	// NicName is required
	if obj.obj.NicName == "" {
		validation = append(validation, "NicName is required field on interface DataflowHostManagement")
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
}{
	SCATTER:    DataflowWorkloadItemChoiceEnum("scatter"),
	GATHER:     DataflowWorkloadItemChoiceEnum("gather"),
	ALL_REDUCE: DataflowWorkloadItemChoiceEnum("all_reduce"),
	LOOP:       DataflowWorkloadItemChoiceEnum("loop"),
	COMPUTE:    DataflowWorkloadItemChoiceEnum("compute"),
	BROADCAST:  DataflowWorkloadItemChoiceEnum("broadcast"),
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

}

func (obj *dataflowWorkloadItem) setDefault() {

}

// ***** DataflowFlowProfile *****
type dataflowFlowProfile struct {
	obj            *onexdataflowapi.DataflowFlowProfile
	ethernetHolder DataflowFlowProfileEthernet
	tcpHolder      DataflowFlowProfileTcp
	udpHolder      DataflowFlowProfileUdp
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
	obj.ethernetHolder = nil
	obj.tcpHolder = nil
	obj.udpHolder = nil
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
	// L2ProtocolChoice returns DataflowFlowProfileL2ProtocolChoiceEnum, set in DataflowFlowProfile
	L2ProtocolChoice() DataflowFlowProfileL2ProtocolChoiceEnum
	// SetL2ProtocolChoice assigns DataflowFlowProfileL2ProtocolChoiceEnum provided by user to DataflowFlowProfile
	SetL2ProtocolChoice(value DataflowFlowProfileL2ProtocolChoiceEnum) DataflowFlowProfile
	// HasL2ProtocolChoice checks if L2ProtocolChoice has been set in DataflowFlowProfile
	HasL2ProtocolChoice() bool
	// Ethernet returns DataflowFlowProfileEthernet, set in DataflowFlowProfile.
	// DataflowFlowProfileEthernet is description is TBD
	Ethernet() DataflowFlowProfileEthernet
	// SetEthernet assigns DataflowFlowProfileEthernet provided by user to DataflowFlowProfile.
	// DataflowFlowProfileEthernet is description is TBD
	SetEthernet(value DataflowFlowProfileEthernet) DataflowFlowProfile
	// HasEthernet checks if Ethernet has been set in DataflowFlowProfile
	HasEthernet() bool
	// L4ProtocolChoice returns DataflowFlowProfileL4ProtocolChoiceEnum, set in DataflowFlowProfile
	L4ProtocolChoice() DataflowFlowProfileL4ProtocolChoiceEnum
	// SetL4ProtocolChoice assigns DataflowFlowProfileL4ProtocolChoiceEnum provided by user to DataflowFlowProfile
	SetL4ProtocolChoice(value DataflowFlowProfileL4ProtocolChoiceEnum) DataflowFlowProfile
	// HasL4ProtocolChoice checks if L4ProtocolChoice has been set in DataflowFlowProfile
	HasL4ProtocolChoice() bool
	// Tcp returns DataflowFlowProfileTcp, set in DataflowFlowProfile.
	// DataflowFlowProfileTcp is description is TBD
	Tcp() DataflowFlowProfileTcp
	// SetTcp assigns DataflowFlowProfileTcp provided by user to DataflowFlowProfile.
	// DataflowFlowProfileTcp is description is TBD
	SetTcp(value DataflowFlowProfileTcp) DataflowFlowProfile
	// HasTcp checks if Tcp has been set in DataflowFlowProfile
	HasTcp() bool
	// Udp returns DataflowFlowProfileUdp, set in DataflowFlowProfile.
	// DataflowFlowProfileUdp is description is TBD
	Udp() DataflowFlowProfileUdp
	// SetUdp assigns DataflowFlowProfileUdp provided by user to DataflowFlowProfile.
	// DataflowFlowProfileUdp is description is TBD
	SetUdp(value DataflowFlowProfileUdp) DataflowFlowProfile
	// HasUdp checks if Udp has been set in DataflowFlowProfile
	HasUdp() bool
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

type DataflowFlowProfileL2ProtocolChoiceEnum string

//  Enum of L2ProtocolChoice on DataflowFlowProfile
var DataflowFlowProfileL2ProtocolChoice = struct {
	ETHERNET DataflowFlowProfileL2ProtocolChoiceEnum
}{
	ETHERNET: DataflowFlowProfileL2ProtocolChoiceEnum("ethernet"),
}

func (obj *dataflowFlowProfile) L2ProtocolChoice() DataflowFlowProfileL2ProtocolChoiceEnum {
	return DataflowFlowProfileL2ProtocolChoiceEnum(obj.obj.L2ProtocolChoice.Enum().String())
}

// L2ProtocolChoice returns a string
// layer2 protocol selection
func (obj *dataflowFlowProfile) HasL2ProtocolChoice() bool {
	return obj.obj.L2ProtocolChoice != nil
}

func (obj *dataflowFlowProfile) SetL2ProtocolChoice(value DataflowFlowProfileL2ProtocolChoiceEnum) DataflowFlowProfile {
	intValue, ok := onexdataflowapi.DataflowFlowProfile_L2ProtocolChoice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileL2ProtocolChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfile_L2ProtocolChoice_Enum(intValue)
	obj.obj.L2ProtocolChoice = &enumValue

	return obj
}

// Ethernet returns a DataflowFlowProfileEthernet
// description is TBD
func (obj *dataflowFlowProfile) Ethernet() DataflowFlowProfileEthernet {
	if obj.obj.Ethernet == nil {
		obj.obj.Ethernet = NewDataflowFlowProfileEthernet().Msg()
	}
	if obj.ethernetHolder == nil {
		obj.ethernetHolder = &dataflowFlowProfileEthernet{obj: obj.obj.Ethernet}
	}
	return obj.ethernetHolder
}

// Ethernet returns a DataflowFlowProfileEthernet
// description is TBD
func (obj *dataflowFlowProfile) HasEthernet() bool {
	return obj.obj.Ethernet != nil
}

// SetEthernet sets the DataflowFlowProfileEthernet value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetEthernet(value DataflowFlowProfileEthernet) DataflowFlowProfile {

	obj.ethernetHolder = nil
	obj.obj.Ethernet = value.Msg()

	return obj
}

type DataflowFlowProfileL4ProtocolChoiceEnum string

//  Enum of L4ProtocolChoice on DataflowFlowProfile
var DataflowFlowProfileL4ProtocolChoice = struct {
	TCP DataflowFlowProfileL4ProtocolChoiceEnum
	UDP DataflowFlowProfileL4ProtocolChoiceEnum
}{
	TCP: DataflowFlowProfileL4ProtocolChoiceEnum("tcp"),
	UDP: DataflowFlowProfileL4ProtocolChoiceEnum("udp"),
}

func (obj *dataflowFlowProfile) L4ProtocolChoice() DataflowFlowProfileL4ProtocolChoiceEnum {
	return DataflowFlowProfileL4ProtocolChoiceEnum(obj.obj.L4ProtocolChoice.Enum().String())
}

// L4ProtocolChoice returns a string
// layer4 protocol selection
func (obj *dataflowFlowProfile) HasL4ProtocolChoice() bool {
	return obj.obj.L4ProtocolChoice != nil
}

func (obj *dataflowFlowProfile) SetL4ProtocolChoice(value DataflowFlowProfileL4ProtocolChoiceEnum) DataflowFlowProfile {
	intValue, ok := onexdataflowapi.DataflowFlowProfile_L4ProtocolChoice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileL4ProtocolChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfile_L4ProtocolChoice_Enum(intValue)
	obj.obj.L4ProtocolChoice = &enumValue

	return obj
}

// Tcp returns a DataflowFlowProfileTcp
// description is TBD
func (obj *dataflowFlowProfile) Tcp() DataflowFlowProfileTcp {
	if obj.obj.Tcp == nil {
		obj.obj.Tcp = NewDataflowFlowProfileTcp().Msg()
	}
	if obj.tcpHolder == nil {
		obj.tcpHolder = &dataflowFlowProfileTcp{obj: obj.obj.Tcp}
	}
	return obj.tcpHolder
}

// Tcp returns a DataflowFlowProfileTcp
// description is TBD
func (obj *dataflowFlowProfile) HasTcp() bool {
	return obj.obj.Tcp != nil
}

// SetTcp sets the DataflowFlowProfileTcp value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetTcp(value DataflowFlowProfileTcp) DataflowFlowProfile {

	obj.tcpHolder = nil
	obj.obj.Tcp = value.Msg()

	return obj
}

// Udp returns a DataflowFlowProfileUdp
// description is TBD
func (obj *dataflowFlowProfile) Udp() DataflowFlowProfileUdp {
	if obj.obj.Udp == nil {
		obj.obj.Udp = NewDataflowFlowProfileUdp().Msg()
	}
	if obj.udpHolder == nil {
		obj.udpHolder = &dataflowFlowProfileUdp{obj: obj.obj.Udp}
	}
	return obj.udpHolder
}

// Udp returns a DataflowFlowProfileUdp
// description is TBD
func (obj *dataflowFlowProfile) HasUdp() bool {
	return obj.obj.Udp != nil
}

// SetUdp sets the DataflowFlowProfileUdp value in the DataflowFlowProfile object
// description is TBD
func (obj *dataflowFlowProfile) SetUdp(value DataflowFlowProfileUdp) DataflowFlowProfile {

	obj.udpHolder = nil
	obj.obj.Udp = value.Msg()

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

	if obj.obj.Ethernet != nil {
		obj.Ethernet().validateObj(set_default)
	}

	if obj.obj.Tcp != nil {
		obj.Tcp().validateObj(set_default)
	}

	if obj.obj.Udp != nil {
		obj.Udp().validateObj(set_default)
	}

}

func (obj *dataflowFlowProfile) setDefault() {

}

// ***** MetricsResponseFlowResult *****
type metricsResponseFlowResult struct {
	obj           *onexdataflowapi.MetricsResponseFlowResult
	tcpInfoHolder MetricsResponseFlowResultTcpInfo
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
	obj.tcpInfoHolder = nil
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
	// FlowNumber returns int32, set in MetricsResponseFlowResult.
	FlowNumber() int32
	// SetFlowNumber assigns int32 provided by user to MetricsResponseFlowResult
	SetFlowNumber(value int32) MetricsResponseFlowResult
	// HasFlowNumber checks if FlowNumber has been set in MetricsResponseFlowResult
	HasFlowNumber() bool
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
	// Fct returns float32, set in MetricsResponseFlowResult.
	Fct() float32
	// SetFct assigns float32 provided by user to MetricsResponseFlowResult
	SetFct(value float32) MetricsResponseFlowResult
	// HasFct checks if Fct has been set in MetricsResponseFlowResult
	HasFct() bool
	// FirstTimestamp returns int32, set in MetricsResponseFlowResult.
	FirstTimestamp() int32
	// SetFirstTimestamp assigns int32 provided by user to MetricsResponseFlowResult
	SetFirstTimestamp(value int32) MetricsResponseFlowResult
	// HasFirstTimestamp checks if FirstTimestamp has been set in MetricsResponseFlowResult
	HasFirstTimestamp() bool
	// LastTimestamp returns int32, set in MetricsResponseFlowResult.
	LastTimestamp() int32
	// SetLastTimestamp assigns int32 provided by user to MetricsResponseFlowResult
	SetLastTimestamp(value int32) MetricsResponseFlowResult
	// HasLastTimestamp checks if LastTimestamp has been set in MetricsResponseFlowResult
	HasLastTimestamp() bool
	// BytesTx returns int32, set in MetricsResponseFlowResult.
	BytesTx() int32
	// SetBytesTx assigns int32 provided by user to MetricsResponseFlowResult
	SetBytesTx(value int32) MetricsResponseFlowResult
	// HasBytesTx checks if BytesTx has been set in MetricsResponseFlowResult
	HasBytesTx() bool
	// BytesRx returns int32, set in MetricsResponseFlowResult.
	BytesRx() int32
	// SetBytesRx assigns int32 provided by user to MetricsResponseFlowResult
	SetBytesRx(value int32) MetricsResponseFlowResult
	// HasBytesRx checks if BytesRx has been set in MetricsResponseFlowResult
	HasBytesRx() bool
	// TcpInfo returns MetricsResponseFlowResultTcpInfo, set in MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	TcpInfo() MetricsResponseFlowResultTcpInfo
	// SetTcpInfo assigns MetricsResponseFlowResultTcpInfo provided by user to MetricsResponseFlowResult.
	// MetricsResponseFlowResultTcpInfo is tCP information for this flow
	SetTcpInfo(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult
	// HasTcpInfo checks if TcpInfo has been set in MetricsResponseFlowResult
	HasTcpInfo() bool
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

// FlowNumber returns a int32
// description is TBD
func (obj *metricsResponseFlowResult) FlowNumber() int32 {

	return *obj.obj.FlowNumber

}

// FlowNumber returns a int32
// description is TBD
func (obj *metricsResponseFlowResult) HasFlowNumber() bool {
	return obj.obj.FlowNumber != nil
}

// SetFlowNumber sets the int32 value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetFlowNumber(value int32) MetricsResponseFlowResult {

	obj.obj.FlowNumber = &value
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

// Fct returns a float32
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) Fct() float32 {

	return *obj.obj.Fct

}

// Fct returns a float32
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) HasFct() bool {
	return obj.obj.Fct != nil
}

// SetFct sets the float32 value in the MetricsResponseFlowResult object
// flow completion time in micro seconds
func (obj *metricsResponseFlowResult) SetFct(value float32) MetricsResponseFlowResult {

	obj.obj.Fct = &value
	return obj
}

// FirstTimestamp returns a int32
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) FirstTimestamp() int32 {

	return *obj.obj.FirstTimestamp

}

// FirstTimestamp returns a int32
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) HasFirstTimestamp() bool {
	return obj.obj.FirstTimestamp != nil
}

// SetFirstTimestamp sets the int32 value in the MetricsResponseFlowResult object
// first timestamp in micro seconds
func (obj *metricsResponseFlowResult) SetFirstTimestamp(value int32) MetricsResponseFlowResult {

	obj.obj.FirstTimestamp = &value
	return obj
}

// LastTimestamp returns a int32
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) LastTimestamp() int32 {

	return *obj.obj.LastTimestamp

}

// LastTimestamp returns a int32
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) HasLastTimestamp() bool {
	return obj.obj.LastTimestamp != nil
}

// SetLastTimestamp sets the int32 value in the MetricsResponseFlowResult object
// last timestamp in micro seconds
func (obj *metricsResponseFlowResult) SetLastTimestamp(value int32) MetricsResponseFlowResult {

	obj.obj.LastTimestamp = &value
	return obj
}

// BytesTx returns a int32
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) BytesTx() int32 {

	return *obj.obj.BytesTx

}

// BytesTx returns a int32
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) HasBytesTx() bool {
	return obj.obj.BytesTx != nil
}

// SetBytesTx sets the int32 value in the MetricsResponseFlowResult object
// bytes transmitted from src to dst
func (obj *metricsResponseFlowResult) SetBytesTx(value int32) MetricsResponseFlowResult {

	obj.obj.BytesTx = &value
	return obj
}

// BytesRx returns a int32
// bytes received by src from dst
func (obj *metricsResponseFlowResult) BytesRx() int32 {

	return *obj.obj.BytesRx

}

// BytesRx returns a int32
// bytes received by src from dst
func (obj *metricsResponseFlowResult) HasBytesRx() bool {
	return obj.obj.BytesRx != nil
}

// SetBytesRx sets the int32 value in the MetricsResponseFlowResult object
// bytes received by src from dst
func (obj *metricsResponseFlowResult) SetBytesRx(value int32) MetricsResponseFlowResult {

	obj.obj.BytesRx = &value
	return obj
}

// TcpInfo returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) TcpInfo() MetricsResponseFlowResultTcpInfo {
	if obj.obj.TcpInfo == nil {
		obj.obj.TcpInfo = NewMetricsResponseFlowResultTcpInfo().Msg()
	}
	if obj.tcpInfoHolder == nil {
		obj.tcpInfoHolder = &metricsResponseFlowResultTcpInfo{obj: obj.obj.TcpInfo}
	}
	return obj.tcpInfoHolder
}

// TcpInfo returns a MetricsResponseFlowResultTcpInfo
// description is TBD
func (obj *metricsResponseFlowResult) HasTcpInfo() bool {
	return obj.obj.TcpInfo != nil
}

// SetTcpInfo sets the MetricsResponseFlowResultTcpInfo value in the MetricsResponseFlowResult object
// description is TBD
func (obj *metricsResponseFlowResult) SetTcpInfo(value MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResult {

	obj.tcpInfoHolder = nil
	obj.obj.TcpInfo = value.Msg()

	return obj
}

func (obj *metricsResponseFlowResult) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.TcpInfo != nil {
		obj.TcpInfo().validateObj(set_default)
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

// ***** DataflowFlowProfileEthernet *****
type dataflowFlowProfileEthernet struct {
	obj *onexdataflowapi.DataflowFlowProfileEthernet
}

func NewDataflowFlowProfileEthernet() DataflowFlowProfileEthernet {
	obj := dataflowFlowProfileEthernet{obj: &onexdataflowapi.DataflowFlowProfileEthernet{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileEthernet) Msg() *onexdataflowapi.DataflowFlowProfileEthernet {
	return obj.obj
}

func (obj *dataflowFlowProfileEthernet) SetMsg(msg *onexdataflowapi.DataflowFlowProfileEthernet) DataflowFlowProfileEthernet {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileEthernet) ToPbText() (string, error) {
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

func (obj *dataflowFlowProfileEthernet) FromPbText(value string) error {
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

func (obj *dataflowFlowProfileEthernet) ToYaml() (string, error) {
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

func (obj *dataflowFlowProfileEthernet) FromYaml(value string) error {
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

func (obj *dataflowFlowProfileEthernet) ToJson() (string, error) {
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

func (obj *dataflowFlowProfileEthernet) FromJson(value string) error {
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

func (obj *dataflowFlowProfileEthernet) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileEthernet) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileEthernet) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowFlowProfileEthernet is description is TBD
type DataflowFlowProfileEthernet interface {
	Msg() *onexdataflowapi.DataflowFlowProfileEthernet
	SetMsg(*onexdataflowapi.DataflowFlowProfileEthernet) DataflowFlowProfileEthernet
	// ToPbText marshals DataflowFlowProfileEthernet to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileEthernet to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileEthernet to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileEthernet from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileEthernet from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileEthernet from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileEthernet
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Mtu returns int32, set in DataflowFlowProfileEthernet.
	Mtu() int32
	// SetMtu assigns int32 provided by user to DataflowFlowProfileEthernet
	SetMtu(value int32) DataflowFlowProfileEthernet
	// HasMtu checks if Mtu has been set in DataflowFlowProfileEthernet
	HasMtu() bool
}

// Mtu returns a int32
// Maximum Transmission Unit
func (obj *dataflowFlowProfileEthernet) Mtu() int32 {

	return *obj.obj.Mtu

}

// Mtu returns a int32
// Maximum Transmission Unit
func (obj *dataflowFlowProfileEthernet) HasMtu() bool {
	return obj.obj.Mtu != nil
}

// SetMtu sets the int32 value in the DataflowFlowProfileEthernet object
// Maximum Transmission Unit
func (obj *dataflowFlowProfileEthernet) SetMtu(value int32) DataflowFlowProfileEthernet {

	obj.obj.Mtu = &value
	return obj
}

func (obj *dataflowFlowProfileEthernet) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowFlowProfileEthernet) setDefault() {
	if obj.obj.Mtu == nil {
		obj.SetMtu(1500)
	}

}

// ***** DataflowFlowProfileTcp *****
type dataflowFlowProfileTcp struct {
	obj                   *onexdataflowapi.DataflowFlowProfileTcp
	destinationPortHolder L4PortRange
	sourcePortHolder      L4PortRange
}

func NewDataflowFlowProfileTcp() DataflowFlowProfileTcp {
	obj := dataflowFlowProfileTcp{obj: &onexdataflowapi.DataflowFlowProfileTcp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileTcp) Msg() *onexdataflowapi.DataflowFlowProfileTcp {
	return obj.obj
}

func (obj *dataflowFlowProfileTcp) SetMsg(msg *onexdataflowapi.DataflowFlowProfileTcp) DataflowFlowProfileTcp {
	obj.setNil()
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileTcp) ToPbText() (string, error) {
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

func (obj *dataflowFlowProfileTcp) FromPbText(value string) error {
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

func (obj *dataflowFlowProfileTcp) ToYaml() (string, error) {
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

func (obj *dataflowFlowProfileTcp) FromYaml(value string) error {
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

func (obj *dataflowFlowProfileTcp) ToJson() (string, error) {
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

func (obj *dataflowFlowProfileTcp) FromJson(value string) error {
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

func (obj *dataflowFlowProfileTcp) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileTcp) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileTcp) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

func (obj *dataflowFlowProfileTcp) setNil() {
	obj.destinationPortHolder = nil
	obj.sourcePortHolder = nil
}

// DataflowFlowProfileTcp is description is TBD
type DataflowFlowProfileTcp interface {
	Msg() *onexdataflowapi.DataflowFlowProfileTcp
	SetMsg(*onexdataflowapi.DataflowFlowProfileTcp) DataflowFlowProfileTcp
	// ToPbText marshals DataflowFlowProfileTcp to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileTcp to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileTcp to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileTcp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileTcp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileTcp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileTcp
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// CongestionAlgorithm returns DataflowFlowProfileTcpCongestionAlgorithmEnum, set in DataflowFlowProfileTcp
	CongestionAlgorithm() DataflowFlowProfileTcpCongestionAlgorithmEnum
	// SetCongestionAlgorithm assigns DataflowFlowProfileTcpCongestionAlgorithmEnum provided by user to DataflowFlowProfileTcp
	SetCongestionAlgorithm(value DataflowFlowProfileTcpCongestionAlgorithmEnum) DataflowFlowProfileTcp
	// HasCongestionAlgorithm checks if CongestionAlgorithm has been set in DataflowFlowProfileTcp
	HasCongestionAlgorithm() bool
	// Initcwnd returns int32, set in DataflowFlowProfileTcp.
	Initcwnd() int32
	// SetInitcwnd assigns int32 provided by user to DataflowFlowProfileTcp
	SetInitcwnd(value int32) DataflowFlowProfileTcp
	// HasInitcwnd checks if Initcwnd has been set in DataflowFlowProfileTcp
	HasInitcwnd() bool
	// SendBuf returns int32, set in DataflowFlowProfileTcp.
	SendBuf() int32
	// SetSendBuf assigns int32 provided by user to DataflowFlowProfileTcp
	SetSendBuf(value int32) DataflowFlowProfileTcp
	// HasSendBuf checks if SendBuf has been set in DataflowFlowProfileTcp
	HasSendBuf() bool
	// ReceiveBuf returns int32, set in DataflowFlowProfileTcp.
	ReceiveBuf() int32
	// SetReceiveBuf assigns int32 provided by user to DataflowFlowProfileTcp
	SetReceiveBuf(value int32) DataflowFlowProfileTcp
	// HasReceiveBuf checks if ReceiveBuf has been set in DataflowFlowProfileTcp
	HasReceiveBuf() bool
	// DestinationPort returns L4PortRange, set in DataflowFlowProfileTcp.
	// L4PortRange is layer4 protocol source or destination port values
	DestinationPort() L4PortRange
	// SetDestinationPort assigns L4PortRange provided by user to DataflowFlowProfileTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SetDestinationPort(value L4PortRange) DataflowFlowProfileTcp
	// HasDestinationPort checks if DestinationPort has been set in DataflowFlowProfileTcp
	HasDestinationPort() bool
	// SourcePort returns L4PortRange, set in DataflowFlowProfileTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SourcePort() L4PortRange
	// SetSourcePort assigns L4PortRange provided by user to DataflowFlowProfileTcp.
	// L4PortRange is layer4 protocol source or destination port values
	SetSourcePort(value L4PortRange) DataflowFlowProfileTcp
	// HasSourcePort checks if SourcePort has been set in DataflowFlowProfileTcp
	HasSourcePort() bool
	setNil()
}

type DataflowFlowProfileTcpCongestionAlgorithmEnum string

//  Enum of CongestionAlgorithm on DataflowFlowProfileTcp
var DataflowFlowProfileTcpCongestionAlgorithm = struct {
	BBR   DataflowFlowProfileTcpCongestionAlgorithmEnum
	DCTCP DataflowFlowProfileTcpCongestionAlgorithmEnum
	CUBIC DataflowFlowProfileTcpCongestionAlgorithmEnum
	RENO  DataflowFlowProfileTcpCongestionAlgorithmEnum
}{
	BBR:   DataflowFlowProfileTcpCongestionAlgorithmEnum("bbr"),
	DCTCP: DataflowFlowProfileTcpCongestionAlgorithmEnum("dctcp"),
	CUBIC: DataflowFlowProfileTcpCongestionAlgorithmEnum("cubic"),
	RENO:  DataflowFlowProfileTcpCongestionAlgorithmEnum("reno"),
}

func (obj *dataflowFlowProfileTcp) CongestionAlgorithm() DataflowFlowProfileTcpCongestionAlgorithmEnum {
	return DataflowFlowProfileTcpCongestionAlgorithmEnum(obj.obj.CongestionAlgorithm.Enum().String())
}

// CongestionAlgorithm returns a string
// The TCP congestion algorithm:
// bbr - Bottleneck Bandwidth and Round-trip propagation time
// dctcp - Data center TCP
// cubic - cubic window increase function
// reno - TCP New Reno
func (obj *dataflowFlowProfileTcp) HasCongestionAlgorithm() bool {
	return obj.obj.CongestionAlgorithm != nil
}

func (obj *dataflowFlowProfileTcp) SetCongestionAlgorithm(value DataflowFlowProfileTcpCongestionAlgorithmEnum) DataflowFlowProfileTcp {
	intValue, ok := onexdataflowapi.DataflowFlowProfileTcp_CongestionAlgorithm_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileTcpCongestionAlgorithmEnum", string(value)))
		return obj
	}
	enumValue := onexdataflowapi.DataflowFlowProfileTcp_CongestionAlgorithm_Enum(intValue)
	obj.obj.CongestionAlgorithm = &enumValue

	return obj
}

// Initcwnd returns a int32
// initial congestion window
func (obj *dataflowFlowProfileTcp) Initcwnd() int32 {

	return *obj.obj.Initcwnd

}

// Initcwnd returns a int32
// initial congestion window
func (obj *dataflowFlowProfileTcp) HasInitcwnd() bool {
	return obj.obj.Initcwnd != nil
}

// SetInitcwnd sets the int32 value in the DataflowFlowProfileTcp object
// initial congestion window
func (obj *dataflowFlowProfileTcp) SetInitcwnd(value int32) DataflowFlowProfileTcp {

	obj.obj.Initcwnd = &value
	return obj
}

// SendBuf returns a int32
// send buffer size
func (obj *dataflowFlowProfileTcp) SendBuf() int32 {

	return *obj.obj.SendBuf

}

// SendBuf returns a int32
// send buffer size
func (obj *dataflowFlowProfileTcp) HasSendBuf() bool {
	return obj.obj.SendBuf != nil
}

// SetSendBuf sets the int32 value in the DataflowFlowProfileTcp object
// send buffer size
func (obj *dataflowFlowProfileTcp) SetSendBuf(value int32) DataflowFlowProfileTcp {

	obj.obj.SendBuf = &value
	return obj
}

// ReceiveBuf returns a int32
// receive buffer size
func (obj *dataflowFlowProfileTcp) ReceiveBuf() int32 {

	return *obj.obj.ReceiveBuf

}

// ReceiveBuf returns a int32
// receive buffer size
func (obj *dataflowFlowProfileTcp) HasReceiveBuf() bool {
	return obj.obj.ReceiveBuf != nil
}

// SetReceiveBuf sets the int32 value in the DataflowFlowProfileTcp object
// receive buffer size
func (obj *dataflowFlowProfileTcp) SetReceiveBuf(value int32) DataflowFlowProfileTcp {

	obj.obj.ReceiveBuf = &value
	return obj
}

// DestinationPort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileTcp) DestinationPort() L4PortRange {
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
func (obj *dataflowFlowProfileTcp) HasDestinationPort() bool {
	return obj.obj.DestinationPort != nil
}

// SetDestinationPort sets the L4PortRange value in the DataflowFlowProfileTcp object
// description is TBD
func (obj *dataflowFlowProfileTcp) SetDestinationPort(value L4PortRange) DataflowFlowProfileTcp {

	obj.destinationPortHolder = nil
	obj.obj.DestinationPort = value.Msg()

	return obj
}

// SourcePort returns a L4PortRange
// description is TBD
func (obj *dataflowFlowProfileTcp) SourcePort() L4PortRange {
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
func (obj *dataflowFlowProfileTcp) HasSourcePort() bool {
	return obj.obj.SourcePort != nil
}

// SetSourcePort sets the L4PortRange value in the DataflowFlowProfileTcp object
// description is TBD
func (obj *dataflowFlowProfileTcp) SetSourcePort(value L4PortRange) DataflowFlowProfileTcp {

	obj.sourcePortHolder = nil
	obj.obj.SourcePort = value.Msg()

	return obj
}

func (obj *dataflowFlowProfileTcp) validateObj(set_default bool) {
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

func (obj *dataflowFlowProfileTcp) setDefault() {
	if obj.obj.CongestionAlgorithm == nil {
		obj.SetCongestionAlgorithm(DataflowFlowProfileTcpCongestionAlgorithm.CUBIC)

	}

}

// ***** DataflowFlowProfileUdp *****
type dataflowFlowProfileUdp struct {
	obj *onexdataflowapi.DataflowFlowProfileUdp
}

func NewDataflowFlowProfileUdp() DataflowFlowProfileUdp {
	obj := dataflowFlowProfileUdp{obj: &onexdataflowapi.DataflowFlowProfileUdp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileUdp) Msg() *onexdataflowapi.DataflowFlowProfileUdp {
	return obj.obj
}

func (obj *dataflowFlowProfileUdp) SetMsg(msg *onexdataflowapi.DataflowFlowProfileUdp) DataflowFlowProfileUdp {

	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileUdp) ToPbText() (string, error) {
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

func (obj *dataflowFlowProfileUdp) FromPbText(value string) error {
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

func (obj *dataflowFlowProfileUdp) ToYaml() (string, error) {
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

func (obj *dataflowFlowProfileUdp) FromYaml(value string) error {
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

func (obj *dataflowFlowProfileUdp) ToJson() (string, error) {
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

func (obj *dataflowFlowProfileUdp) FromJson(value string) error {
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

func (obj *dataflowFlowProfileUdp) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *dataflowFlowProfileUdp) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *dataflowFlowProfileUdp) String() string {
	str, err := obj.ToYaml()
	if err != nil {
		return err.Error()
	}
	return str
}

// DataflowFlowProfileUdp is description is TBD
type DataflowFlowProfileUdp interface {
	Msg() *onexdataflowapi.DataflowFlowProfileUdp
	SetMsg(*onexdataflowapi.DataflowFlowProfileUdp) DataflowFlowProfileUdp
	// ToPbText marshals DataflowFlowProfileUdp to protobuf text
	ToPbText() (string, error)
	// ToYaml marshals DataflowFlowProfileUdp to YAML text
	ToYaml() (string, error)
	// ToJson marshals DataflowFlowProfileUdp to JSON text
	ToJson() (string, error)
	// FromPbText unmarshals DataflowFlowProfileUdp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileUdp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileUdp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileUdp
	Validate() error
	// A stringer function
	String() string
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *dataflowFlowProfileUdp) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *dataflowFlowProfileUdp) setDefault() {

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
	// Rtt returns float32, set in MetricsResponseFlowResultTcpInfo.
	Rtt() float32
	// SetRtt assigns float32 provided by user to MetricsResponseFlowResultTcpInfo
	SetRtt(value float32) MetricsResponseFlowResultTcpInfo
	// HasRtt checks if Rtt has been set in MetricsResponseFlowResultTcpInfo
	HasRtt() bool
	// RttVariance returns float32, set in MetricsResponseFlowResultTcpInfo.
	RttVariance() float32
	// SetRttVariance assigns float32 provided by user to MetricsResponseFlowResultTcpInfo
	SetRttVariance(value float32) MetricsResponseFlowResultTcpInfo
	// HasRttVariance checks if RttVariance has been set in MetricsResponseFlowResultTcpInfo
	HasRttVariance() bool
	// Retransmissions returns float32, set in MetricsResponseFlowResultTcpInfo.
	Retransmissions() float32
	// SetRetransmissions assigns float32 provided by user to MetricsResponseFlowResultTcpInfo
	SetRetransmissions(value float32) MetricsResponseFlowResultTcpInfo
	// HasRetransmissions checks if Retransmissions has been set in MetricsResponseFlowResultTcpInfo
	HasRetransmissions() bool
}

// Rtt returns a float32
// round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) Rtt() float32 {

	return *obj.obj.Rtt

}

// Rtt returns a float32
// round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) HasRtt() bool {
	return obj.obj.Rtt != nil
}

// SetRtt sets the float32 value in the MetricsResponseFlowResultTcpInfo object
// round trip time in microseconds
func (obj *metricsResponseFlowResultTcpInfo) SetRtt(value float32) MetricsResponseFlowResultTcpInfo {

	obj.obj.Rtt = &value
	return obj
}

// RttVariance returns a float32
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) RttVariance() float32 {

	return *obj.obj.RttVariance

}

// RttVariance returns a float32
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) HasRttVariance() bool {
	return obj.obj.RttVariance != nil
}

// SetRttVariance sets the float32 value in the MetricsResponseFlowResultTcpInfo object
// round trip time variance in microseconds, larger values indicate less stable performance
func (obj *metricsResponseFlowResultTcpInfo) SetRttVariance(value float32) MetricsResponseFlowResultTcpInfo {

	obj.obj.RttVariance = &value
	return obj
}

// Retransmissions returns a float32
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) Retransmissions() float32 {

	return *obj.obj.Retransmissions

}

// Retransmissions returns a float32
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) HasRetransmissions() bool {
	return obj.obj.Retransmissions != nil
}

// SetRetransmissions sets the float32 value in the MetricsResponseFlowResultTcpInfo object
// total number of TCP retransmissions
func (obj *metricsResponseFlowResultTcpInfo) SetRetransmissions(value float32) MetricsResponseFlowResultTcpInfo {

	obj.obj.Retransmissions = &value
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
