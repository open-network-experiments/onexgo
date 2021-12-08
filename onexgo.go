// ONEx Data Models 0.0.1
// License: MIT

package onexgo

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
	"github.com/golang/protobuf/proto"
	onexdatamodel "github.com/open-network-experiments/onexgo/onexdatamodel"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type onexgoApi struct {
	api
	grpcClient onexdatamodel.OpenapiClient
	httpClient httpClient
}

// grpcConnect builds up a grpc connection
func (api *onexgoApi) grpcConnect() error {
	if api.grpcClient == nil {
		ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.dialTimeout)
		defer cancelFunc()
		conn, err := grpc.DialContext(ctx, api.grpc.location, grpc.WithInsecure())
		if err != nil {
			return err
		}
		api.grpcClient = onexdatamodel.NewOpenapiClient(conn)
		api.grpc.clientConnection = conn
	}
	return nil
}

func (api *onexgoApi) grpcClose() error {
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

func (api *onexgoApi) Close() error {
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
func NewApi() OnexgoApi {
	api := onexgoApi{}
	return &api
}

// httpConnect builds up a http connection
func (api *onexgoApi) httpConnect() error {
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

func (api *onexgoApi) httpSendRecv(urlPath string, jsonBody string, method string) (*http.Response, error) {
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

// OnexgoApi the data models for the entire ONEx ecosystem
type OnexgoApi interface {
	Api
	// NewConfig returns a new instance of Config.
	// Config is the resources that define a fabric configuration.
	NewConfig() Config
	// NewConfigRequest returns a new instance of ConfigRequest.
	// ConfigRequest is getConfig request details
	NewConfigRequest() ConfigRequest
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
	// SetConfig sets the unified ONEx data model
	SetConfig(config Config) (WarningDetails, error)
	// GetConfig gets the unified ONEx data model from the server, as currently configured
	GetConfig(configRequest ConfigRequest) (ConfigResponse, error)
	// RunExperiment runs the currently configured experiment
	RunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error)
	// GetMetrics description is TBD
	GetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error)
}

func (api *onexgoApi) NewConfig() Config {
	return NewConfig()
}

func (api *onexgoApi) NewConfigRequest() ConfigRequest {
	return NewConfigRequest()
}

func (api *onexgoApi) NewExperimentRequest() ExperimentRequest {
	return NewExperimentRequest()
}

func (api *onexgoApi) NewMetricsRequest() MetricsRequest {
	return NewMetricsRequest()
}

func (api *onexgoApi) NewSetConfigResponse() SetConfigResponse {
	return NewSetConfigResponse()
}

func (api *onexgoApi) NewGetConfigResponse() GetConfigResponse {
	return NewGetConfigResponse()
}

func (api *onexgoApi) NewRunExperimentResponse() RunExperimentResponse {
	return NewRunExperimentResponse()
}

func (api *onexgoApi) NewGetMetricsResponse() GetMetricsResponse {
	return NewGetMetricsResponse()
}

func (api *onexgoApi) SetConfig(config Config) (WarningDetails, error) {

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
	request := onexdatamodel.SetConfigRequest{Config: config.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.SetConfig(ctx, &request)
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

func (api *onexgoApi) GetConfig(configRequest ConfigRequest) (ConfigResponse, error) {

	err := configRequest.Validate()
	if err != nil {
		return nil, err
	}

	if api.hasHttpTransport() {
		return api.httpGetConfig(configRequest)
	}

	if err := api.grpcConnect(); err != nil {
		return nil, err
	}
	request := onexdatamodel.GetConfigRequest{ConfigRequest: configRequest.Msg()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), api.grpc.requestTimeout)
	defer cancelFunc()
	resp, err := api.grpcClient.GetConfig(ctx, &request)
	if err != nil {
		return nil, err
	}
	if resp.GetStatusCode_200() != nil {
		return NewConfigResponse().SetMsg(resp.GetStatusCode_200()), nil
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

func (api *onexgoApi) RunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error) {

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
	request := onexdatamodel.RunExperimentRequest{ExperimentRequest: experimentRequest.Msg()}
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

func (api *onexgoApi) GetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error) {

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
	request := onexdatamodel.GetMetricsRequest{MetricsRequest: metricsRequest.Msg()}
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

func (api *onexgoApi) httpSetConfig(config Config) (WarningDetails, error) {
	resp, err := api.httpSendRecv("config", config.ToJson(), "POST")
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

func (api *onexgoApi) httpGetConfig(configRequest ConfigRequest) (ConfigResponse, error) {
	resp, err := api.httpSendRecv("config", configRequest.ToJson(), "GET")
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

func (api *onexgoApi) httpRunExperiment(experimentRequest ExperimentRequest) (WarningDetails, error) {
	resp, err := api.httpSendRecv("control/experiment", experimentRequest.ToJson(), "POST")
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

func (api *onexgoApi) httpGetMetrics(metricsRequest MetricsRequest) (MetricsResponse, error) {
	resp, err := api.httpSendRecv("results/metrics", metricsRequest.ToJson(), "POST")
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

type config struct {
	obj            *onexdatamodel.Config
	hostsHolder    ConfigHostIter
	fabricHolder   Fabric
	dataflowHolder Dataflow
	chaosHolder    Chaos
}

func NewConfig() Config {
	obj := config{obj: &onexdatamodel.Config{}}
	obj.setDefault()
	return &obj
}

func (obj *config) Msg() *onexdatamodel.Config {
	return obj.obj
}

func (obj *config) SetMsg(msg *onexdatamodel.Config) Config {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *config) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *config) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *config) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *config) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *config) setNil() {
	obj.hostsHolder = nil
	obj.fabricHolder = nil
	obj.dataflowHolder = nil
	obj.chaosHolder = nil
}

// Config is the resources that define a fabric configuration.
type Config interface {
	Msg() *onexdatamodel.Config
	SetMsg(*onexdatamodel.Config) Config
	// ToPbText marshals Config to protobuf text
	ToPbText() string
	// ToYaml marshals Config to YAML text
	ToYaml() string
	// ToJson marshals Config to JSON text
	ToJson() string
	// FromPbText unmarshals Config from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Config from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Config from JSON text
	FromJson(value string) error
	// Validate validates Config
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Hosts returns ConfigHostIter, set in Config
	Hosts() ConfigHostIter
	// Fabric returns Fabric, set in Config.
	// Fabric is description is TBD
	Fabric() Fabric
	// SetFabric assigns Fabric provided by user to Config.
	// Fabric is description is TBD
	SetFabric(value Fabric) Config
	// HasFabric checks if Fabric has been set in Config
	HasFabric() bool
	// Dataflow returns Dataflow, set in Config.
	// Dataflow is description is TBD
	Dataflow() Dataflow
	// SetDataflow assigns Dataflow provided by user to Config.
	// Dataflow is description is TBD
	SetDataflow(value Dataflow) Config
	// HasDataflow checks if Dataflow has been set in Config
	HasDataflow() bool
	// Chaos returns Chaos, set in Config.
	// Chaos is description is TBD
	Chaos() Chaos
	// SetChaos assigns Chaos provided by user to Config.
	// Chaos is description is TBD
	SetChaos(value Chaos) Config
	// HasChaos checks if Chaos has been set in Config
	HasChaos() bool
	setNil()
}

// Hosts returns a []Host
// description is TBD
func (obj *config) Hosts() ConfigHostIter {
	if len(obj.obj.Hosts) == 0 {
		obj.obj.Hosts = []*onexdatamodel.Host{}
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
	newObj := &onexdatamodel.Host{}
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
		obj.obj.obj.Hosts = []*onexdatamodel.Host{}
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

// Fabric returns a Fabric
// description is TBD
func (obj *config) Fabric() Fabric {
	if obj.obj.Fabric == nil {
		obj.obj.Fabric = NewFabric().Msg()
	}
	if obj.fabricHolder == nil {
		obj.fabricHolder = &fabric{obj: obj.obj.Fabric}
	}
	return obj.fabricHolder
}

// Fabric returns a Fabric
// description is TBD
func (obj *config) HasFabric() bool {
	return obj.obj.Fabric != nil
}

// SetFabric sets the Fabric value in the Config object
// description is TBD
func (obj *config) SetFabric(value Fabric) Config {

	obj.Fabric().SetMsg(value.Msg())
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

	obj.Dataflow().SetMsg(value.Msg())
	return obj
}

// Chaos returns a Chaos
// description is TBD
func (obj *config) Chaos() Chaos {
	if obj.obj.Chaos == nil {
		obj.obj.Chaos = NewChaos().Msg()
	}
	if obj.chaosHolder == nil {
		obj.chaosHolder = &chaos{obj: obj.obj.Chaos}
	}
	return obj.chaosHolder
}

// Chaos returns a Chaos
// description is TBD
func (obj *config) HasChaos() bool {
	return obj.obj.Chaos != nil
}

// SetChaos sets the Chaos value in the Config object
// description is TBD
func (obj *config) SetChaos(value Chaos) Config {

	obj.Chaos().SetMsg(value.Msg())
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

	if obj.obj.Fabric != nil {
		obj.Fabric().validateObj(set_default)
	}

	if obj.obj.Dataflow != nil {
		obj.Dataflow().validateObj(set_default)
	}

	if obj.obj.Chaos != nil {
		obj.Chaos().validateObj(set_default)
	}

}

func (obj *config) setDefault() {

}

type configRequest struct {
	obj *onexdatamodel.ConfigRequest
}

func NewConfigRequest() ConfigRequest {
	obj := configRequest{obj: &onexdatamodel.ConfigRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *configRequest) Msg() *onexdatamodel.ConfigRequest {
	return obj.obj
}

func (obj *configRequest) SetMsg(msg *onexdatamodel.ConfigRequest) ConfigRequest {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *configRequest) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *configRequest) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *configRequest) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *configRequest) FromYaml(value string) error {
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

func (obj *configRequest) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *configRequest) FromJson(value string) error {
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

func (obj *configRequest) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *configRequest) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// ConfigRequest is getConfig request details
type ConfigRequest interface {
	Msg() *onexdatamodel.ConfigRequest
	SetMsg(*onexdatamodel.ConfigRequest) ConfigRequest
	// ToPbText marshals ConfigRequest to protobuf text
	ToPbText() string
	// ToYaml marshals ConfigRequest to YAML text
	ToYaml() string
	// ToJson marshals ConfigRequest to JSON text
	ToJson() string
	// FromPbText unmarshals ConfigRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ConfigRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ConfigRequest from JSON text
	FromJson(value string) error
	// Validate validates ConfigRequest
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *configRequest) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *configRequest) setDefault() {

}

type experimentRequest struct {
	obj *onexdatamodel.ExperimentRequest
}

func NewExperimentRequest() ExperimentRequest {
	obj := experimentRequest{obj: &onexdatamodel.ExperimentRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *experimentRequest) Msg() *onexdatamodel.ExperimentRequest {
	return obj.obj
}

func (obj *experimentRequest) SetMsg(msg *onexdatamodel.ExperimentRequest) ExperimentRequest {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *experimentRequest) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *experimentRequest) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *experimentRequest) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *experimentRequest) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// ExperimentRequest is experiment request details
type ExperimentRequest interface {
	Msg() *onexdatamodel.ExperimentRequest
	SetMsg(*onexdatamodel.ExperimentRequest) ExperimentRequest
	// ToPbText marshals ExperimentRequest to protobuf text
	ToPbText() string
	// ToYaml marshals ExperimentRequest to YAML text
	ToYaml() string
	// ToJson marshals ExperimentRequest to JSON text
	ToJson() string
	// FromPbText unmarshals ExperimentRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ExperimentRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ExperimentRequest from JSON text
	FromJson(value string) error
	// Validate validates ExperimentRequest
	Validate() error
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

type metricsRequest struct {
	obj *onexdatamodel.MetricsRequest
}

func NewMetricsRequest() MetricsRequest {
	obj := metricsRequest{obj: &onexdatamodel.MetricsRequest{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsRequest) Msg() *onexdatamodel.MetricsRequest {
	return obj.obj
}

func (obj *metricsRequest) SetMsg(msg *onexdatamodel.MetricsRequest) MetricsRequest {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsRequest) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *metricsRequest) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsRequest) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsRequest) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// MetricsRequest is metrics request details
type MetricsRequest interface {
	Msg() *onexdatamodel.MetricsRequest
	SetMsg(*onexdatamodel.MetricsRequest) MetricsRequest
	// ToPbText marshals MetricsRequest to protobuf text
	ToPbText() string
	// ToYaml marshals MetricsRequest to YAML text
	ToYaml() string
	// ToJson marshals MetricsRequest to JSON text
	ToJson() string
	// FromPbText unmarshals MetricsRequest from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsRequest from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsRequest from JSON text
	FromJson(value string) error
	// Validate validates MetricsRequest
	Validate() error
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

type setConfigResponse struct {
	obj                  *onexdatamodel.SetConfigResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
	statusCode_200Holder WarningDetails
}

func NewSetConfigResponse() SetConfigResponse {
	obj := setConfigResponse{obj: &onexdatamodel.SetConfigResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *setConfigResponse) Msg() *onexdatamodel.SetConfigResponse {
	return obj.obj
}

func (obj *setConfigResponse) SetMsg(msg *onexdatamodel.SetConfigResponse) SetConfigResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *setConfigResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *setConfigResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *setConfigResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *setConfigResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *setConfigResponse) setNil() {
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
	obj.statusCode_200Holder = nil
}

// SetConfigResponse is description is TBD
type SetConfigResponse interface {
	Msg() *onexdatamodel.SetConfigResponse
	SetMsg(*onexdatamodel.SetConfigResponse) SetConfigResponse
	// ToPbText marshals SetConfigResponse to protobuf text
	ToPbText() string
	// ToYaml marshals SetConfigResponse to YAML text
	ToYaml() string
	// ToJson marshals SetConfigResponse to JSON text
	ToJson() string
	// FromPbText unmarshals SetConfigResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals SetConfigResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals SetConfigResponse from JSON text
	FromJson(value string) error
	// Validate validates SetConfigResponse
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
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
	// StatusCode200 returns WarningDetails, set in SetConfigResponse.
	// WarningDetails is description is TBD
	StatusCode200() WarningDetails
	// SetStatusCode200 assigns WarningDetails provided by user to SetConfigResponse.
	// WarningDetails is description is TBD
	SetStatusCode200(value WarningDetails) SetConfigResponse
	// HasStatusCode200 checks if StatusCode200 has been set in SetConfigResponse
	HasStatusCode200() bool
	setNil()
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

	obj.StatusCode400().SetMsg(value.Msg())
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

	obj.StatusCode500().SetMsg(value.Msg())
	return obj
}

// StatusCode200 returns a WarningDetails
// description is TBD
func (obj *setConfigResponse) StatusCode200() WarningDetails {
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
func (obj *setConfigResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the WarningDetails value in the SetConfigResponse object
// description is TBD
func (obj *setConfigResponse) SetStatusCode200(value WarningDetails) SetConfigResponse {

	obj.StatusCode200().SetMsg(value.Msg())
	return obj
}

func (obj *setConfigResponse) validateObj(set_default bool) {
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

func (obj *setConfigResponse) setDefault() {

}

type getConfigResponse struct {
	obj                  *onexdatamodel.GetConfigResponse
	statusCode_200Holder ConfigResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewGetConfigResponse() GetConfigResponse {
	obj := getConfigResponse{obj: &onexdatamodel.GetConfigResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *getConfigResponse) Msg() *onexdatamodel.GetConfigResponse {
	return obj.obj
}

func (obj *getConfigResponse) SetMsg(msg *onexdatamodel.GetConfigResponse) GetConfigResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getConfigResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *getConfigResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *getConfigResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *getConfigResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *getConfigResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// GetConfigResponse is description is TBD
type GetConfigResponse interface {
	Msg() *onexdatamodel.GetConfigResponse
	SetMsg(*onexdatamodel.GetConfigResponse) GetConfigResponse
	// ToPbText marshals GetConfigResponse to protobuf text
	ToPbText() string
	// ToYaml marshals GetConfigResponse to YAML text
	ToYaml() string
	// ToJson marshals GetConfigResponse to JSON text
	ToJson() string
	// FromPbText unmarshals GetConfigResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetConfigResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetConfigResponse from JSON text
	FromJson(value string) error
	// Validate validates GetConfigResponse
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// StatusCode200 returns ConfigResponse, set in GetConfigResponse.
	// ConfigResponse is getConfig response details
	StatusCode200() ConfigResponse
	// SetStatusCode200 assigns ConfigResponse provided by user to GetConfigResponse.
	// ConfigResponse is getConfig response details
	SetStatusCode200(value ConfigResponse) GetConfigResponse
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

// StatusCode200 returns a ConfigResponse
// description is TBD
func (obj *getConfigResponse) StatusCode200() ConfigResponse {
	if obj.obj.StatusCode_200 == nil {
		obj.obj.StatusCode_200 = NewConfigResponse().Msg()
	}
	if obj.statusCode_200Holder == nil {
		obj.statusCode_200Holder = &configResponse{obj: obj.obj.StatusCode_200}
	}
	return obj.statusCode_200Holder
}

// StatusCode200 returns a ConfigResponse
// description is TBD
func (obj *getConfigResponse) HasStatusCode200() bool {
	return obj.obj.StatusCode_200 != nil
}

// SetStatusCode200 sets the ConfigResponse value in the GetConfigResponse object
// description is TBD
func (obj *getConfigResponse) SetStatusCode200(value ConfigResponse) GetConfigResponse {

	obj.StatusCode200().SetMsg(value.Msg())
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

	obj.StatusCode400().SetMsg(value.Msg())
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

	obj.StatusCode500().SetMsg(value.Msg())
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

type runExperimentResponse struct {
	obj                  *onexdatamodel.RunExperimentResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
	statusCode_200Holder WarningDetails
}

func NewRunExperimentResponse() RunExperimentResponse {
	obj := runExperimentResponse{obj: &onexdatamodel.RunExperimentResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *runExperimentResponse) Msg() *onexdatamodel.RunExperimentResponse {
	return obj.obj
}

func (obj *runExperimentResponse) SetMsg(msg *onexdatamodel.RunExperimentResponse) RunExperimentResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *runExperimentResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *runExperimentResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *runExperimentResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *runExperimentResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *runExperimentResponse) setNil() {
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
	obj.statusCode_200Holder = nil
}

// RunExperimentResponse is description is TBD
type RunExperimentResponse interface {
	Msg() *onexdatamodel.RunExperimentResponse
	SetMsg(*onexdatamodel.RunExperimentResponse) RunExperimentResponse
	// ToPbText marshals RunExperimentResponse to protobuf text
	ToPbText() string
	// ToYaml marshals RunExperimentResponse to YAML text
	ToYaml() string
	// ToJson marshals RunExperimentResponse to JSON text
	ToJson() string
	// FromPbText unmarshals RunExperimentResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals RunExperimentResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals RunExperimentResponse from JSON text
	FromJson(value string) error
	// Validate validates RunExperimentResponse
	Validate() error
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

	obj.StatusCode400().SetMsg(value.Msg())
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

	obj.StatusCode500().SetMsg(value.Msg())
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

	obj.StatusCode200().SetMsg(value.Msg())
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

type getMetricsResponse struct {
	obj                  *onexdatamodel.GetMetricsResponse
	statusCode_200Holder MetricsResponse
	statusCode_400Holder ErrorDetails
	statusCode_500Holder ErrorDetails
}

func NewGetMetricsResponse() GetMetricsResponse {
	obj := getMetricsResponse{obj: &onexdatamodel.GetMetricsResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *getMetricsResponse) Msg() *onexdatamodel.GetMetricsResponse {
	return obj.obj
}

func (obj *getMetricsResponse) SetMsg(msg *onexdatamodel.GetMetricsResponse) GetMetricsResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *getMetricsResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *getMetricsResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *getMetricsResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *getMetricsResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *getMetricsResponse) setNil() {
	obj.statusCode_200Holder = nil
	obj.statusCode_400Holder = nil
	obj.statusCode_500Holder = nil
}

// GetMetricsResponse is description is TBD
type GetMetricsResponse interface {
	Msg() *onexdatamodel.GetMetricsResponse
	SetMsg(*onexdatamodel.GetMetricsResponse) GetMetricsResponse
	// ToPbText marshals GetMetricsResponse to protobuf text
	ToPbText() string
	// ToYaml marshals GetMetricsResponse to YAML text
	ToYaml() string
	// ToJson marshals GetMetricsResponse to JSON text
	ToJson() string
	// FromPbText unmarshals GetMetricsResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals GetMetricsResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals GetMetricsResponse from JSON text
	FromJson(value string) error
	// Validate validates GetMetricsResponse
	Validate() error
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

	obj.StatusCode200().SetMsg(value.Msg())
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

	obj.StatusCode400().SetMsg(value.Msg())
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

	obj.StatusCode500().SetMsg(value.Msg())
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

type host struct {
	obj *onexdatamodel.Host
}

func NewHost() Host {
	obj := host{obj: &onexdatamodel.Host{}}
	obj.setDefault()
	return &obj
}

func (obj *host) Msg() *onexdatamodel.Host {
	return obj.obj
}

func (obj *host) SetMsg(msg *onexdatamodel.Host) Host {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *host) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *host) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *host) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *host) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// Host is description is TBD
type Host interface {
	Msg() *onexdatamodel.Host
	SetMsg(*onexdatamodel.Host) Host
	// ToPbText marshals Host to protobuf text
	ToPbText() string
	// ToYaml marshals Host to YAML text
	ToYaml() string
	// ToJson marshals Host to JSON text
	ToJson() string
	// FromPbText unmarshals Host from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Host from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Host from JSON text
	FromJson(value string) error
	// Validate validates Host
	Validate() error
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

type fabric struct {
	obj                *onexdatamodel.Fabric
	spinePodRackHolder FabricSpinePodRack
	qosProfilesHolder  FabricFabricQosProfileIter
}

func NewFabric() Fabric {
	obj := fabric{obj: &onexdatamodel.Fabric{}}
	obj.setDefault()
	return &obj
}

func (obj *fabric) Msg() *onexdatamodel.Fabric {
	return obj.obj
}

func (obj *fabric) SetMsg(msg *onexdatamodel.Fabric) Fabric {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabric) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabric) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *fabric) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabric) FromYaml(value string) error {
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

func (obj *fabric) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabric) FromJson(value string) error {
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

func (obj *fabric) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabric) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *fabric) setNil() {
	obj.spinePodRackHolder = nil
	obj.qosProfilesHolder = nil
}

// Fabric is description is TBD
type Fabric interface {
	Msg() *onexdatamodel.Fabric
	SetMsg(*onexdatamodel.Fabric) Fabric
	// ToPbText marshals Fabric to protobuf text
	ToPbText() string
	// ToYaml marshals Fabric to YAML text
	ToYaml() string
	// ToJson marshals Fabric to JSON text
	ToJson() string
	// FromPbText unmarshals Fabric from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Fabric from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Fabric from JSON text
	FromJson(value string) error
	// Validate validates Fabric
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Choice returns FabricChoiceEnum, set in Fabric
	Choice() FabricChoiceEnum
	// SetChoice assigns FabricChoiceEnum provided by user to Fabric
	SetChoice(value FabricChoiceEnum) Fabric
	// HasChoice checks if Choice has been set in Fabric
	HasChoice() bool
	// SpinePodRack returns FabricSpinePodRack, set in Fabric.
	// FabricSpinePodRack is an emulation of a spine/pod/rack topology
	SpinePodRack() FabricSpinePodRack
	// SetSpinePodRack assigns FabricSpinePodRack provided by user to Fabric.
	// FabricSpinePodRack is an emulation of a spine/pod/rack topology
	SetSpinePodRack(value FabricSpinePodRack) Fabric
	// HasSpinePodRack checks if SpinePodRack has been set in Fabric
	HasSpinePodRack() bool
	// QosProfiles returns FabricFabricQosProfileIter, set in Fabric
	QosProfiles() FabricFabricQosProfileIter
	setNil()
}

type FabricChoiceEnum string

//  Enum of Choice on Fabric
var FabricChoice = struct {
	SPINE_POD_RACK FabricChoiceEnum
}{
	SPINE_POD_RACK: FabricChoiceEnum("spine_pod_rack"),
}

func (obj *fabric) Choice() FabricChoiceEnum {
	return FabricChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *fabric) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *fabric) SetChoice(value FabricChoiceEnum) Fabric {
	intValue, ok := onexdatamodel.Fabric_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.Fabric_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.SpinePodRack = nil
	obj.spinePodRackHolder = nil

	if value == FabricChoice.SPINE_POD_RACK {
		obj.obj.SpinePodRack = NewFabricSpinePodRack().Msg()
	}

	return obj
}

// SpinePodRack returns a FabricSpinePodRack
// description is TBD
func (obj *fabric) SpinePodRack() FabricSpinePodRack {
	if obj.obj.SpinePodRack == nil {
		obj.SetChoice(FabricChoice.SPINE_POD_RACK)
	}
	if obj.spinePodRackHolder == nil {
		obj.spinePodRackHolder = &fabricSpinePodRack{obj: obj.obj.SpinePodRack}
	}
	return obj.spinePodRackHolder
}

// SpinePodRack returns a FabricSpinePodRack
// description is TBD
func (obj *fabric) HasSpinePodRack() bool {
	return obj.obj.SpinePodRack != nil
}

// SetSpinePodRack sets the FabricSpinePodRack value in the Fabric object
// description is TBD
func (obj *fabric) SetSpinePodRack(value FabricSpinePodRack) Fabric {
	obj.SetChoice(FabricChoice.SPINE_POD_RACK)
	obj.SpinePodRack().SetMsg(value.Msg())
	return obj
}

// QosProfiles returns a []FabricQosProfile
// A list of Quality of Service (QoS) profiles
func (obj *fabric) QosProfiles() FabricFabricQosProfileIter {
	if len(obj.obj.QosProfiles) == 0 {
		obj.obj.QosProfiles = []*onexdatamodel.FabricQosProfile{}
	}
	if obj.qosProfilesHolder == nil {
		obj.qosProfilesHolder = newFabricFabricQosProfileIter().setMsg(obj)
	}
	return obj.qosProfilesHolder
}

type fabricFabricQosProfileIter struct {
	obj                   *fabric
	fabricQosProfileSlice []FabricQosProfile
}

func newFabricFabricQosProfileIter() FabricFabricQosProfileIter {
	return &fabricFabricQosProfileIter{}
}

type FabricFabricQosProfileIter interface {
	setMsg(*fabric) FabricFabricQosProfileIter
	Items() []FabricQosProfile
	Add() FabricQosProfile
	Append(items ...FabricQosProfile) FabricFabricQosProfileIter
	Set(index int, newObj FabricQosProfile) FabricFabricQosProfileIter
	Clear() FabricFabricQosProfileIter
	clearHolderSlice() FabricFabricQosProfileIter
	appendHolderSlice(item FabricQosProfile) FabricFabricQosProfileIter
}

func (obj *fabricFabricQosProfileIter) setMsg(msg *fabric) FabricFabricQosProfileIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.QosProfiles {
		obj.appendHolderSlice(&fabricQosProfile{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricFabricQosProfileIter) Items() []FabricQosProfile {
	return obj.fabricQosProfileSlice
}

func (obj *fabricFabricQosProfileIter) Add() FabricQosProfile {
	newObj := &onexdatamodel.FabricQosProfile{}
	obj.obj.obj.QosProfiles = append(obj.obj.obj.QosProfiles, newObj)
	newLibObj := &fabricQosProfile{obj: newObj}
	newLibObj.setDefault()
	obj.fabricQosProfileSlice = append(obj.fabricQosProfileSlice, newLibObj)
	return newLibObj
}

func (obj *fabricFabricQosProfileIter) Append(items ...FabricQosProfile) FabricFabricQosProfileIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.QosProfiles = append(obj.obj.obj.QosProfiles, newObj)
		obj.fabricQosProfileSlice = append(obj.fabricQosProfileSlice, item)
	}
	return obj
}

func (obj *fabricFabricQosProfileIter) Set(index int, newObj FabricQosProfile) FabricFabricQosProfileIter {
	obj.obj.obj.QosProfiles[index] = newObj.Msg()
	obj.fabricQosProfileSlice[index] = newObj
	return obj
}
func (obj *fabricFabricQosProfileIter) Clear() FabricFabricQosProfileIter {
	if len(obj.obj.obj.QosProfiles) > 0 {
		obj.obj.obj.QosProfiles = []*onexdatamodel.FabricQosProfile{}
		obj.fabricQosProfileSlice = []FabricQosProfile{}
	}
	return obj
}
func (obj *fabricFabricQosProfileIter) clearHolderSlice() FabricFabricQosProfileIter {
	if len(obj.fabricQosProfileSlice) > 0 {
		obj.fabricQosProfileSlice = []FabricQosProfile{}
	}
	return obj
}
func (obj *fabricFabricQosProfileIter) appendHolderSlice(item FabricQosProfile) FabricFabricQosProfileIter {
	obj.fabricQosProfileSlice = append(obj.fabricQosProfileSlice, item)
	return obj
}

func (obj *fabric) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.SpinePodRack != nil {
		obj.SpinePodRack().validateObj(set_default)
	}

	if obj.obj.QosProfiles != nil {

		if set_default {
			obj.QosProfiles().clearHolderSlice()
			for _, item := range obj.obj.QosProfiles {
				obj.QosProfiles().appendHolderSlice(&fabricQosProfile{obj: item})
			}
		}
		for _, item := range obj.QosProfiles().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *fabric) setDefault() {

}

type dataflow struct {
	obj                  *onexdatamodel.Dataflow
	hostManagementHolder DataflowDataflowHostManagementIter
	workloadHolder       DataflowDataflowWorkloadItemIter
	flowProfilesHolder   DataflowDataflowFlowProfileIter
}

func NewDataflow() Dataflow {
	obj := dataflow{obj: &onexdatamodel.Dataflow{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflow) Msg() *onexdatamodel.Dataflow {
	return obj.obj
}

func (obj *dataflow) SetMsg(msg *onexdatamodel.Dataflow) Dataflow {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflow) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflow) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflow) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflow) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflow) setNil() {
	obj.hostManagementHolder = nil
	obj.workloadHolder = nil
	obj.flowProfilesHolder = nil
}

// Dataflow is description is TBD
type Dataflow interface {
	Msg() *onexdatamodel.Dataflow
	SetMsg(*onexdatamodel.Dataflow) Dataflow
	// ToPbText marshals Dataflow to protobuf text
	ToPbText() string
	// ToYaml marshals Dataflow to YAML text
	ToYaml() string
	// ToJson marshals Dataflow to JSON text
	ToJson() string
	// FromPbText unmarshals Dataflow from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Dataflow from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Dataflow from JSON text
	FromJson(value string) error
	// Validate validates Dataflow
	Validate() error
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
		obj.obj.HostManagement = []*onexdatamodel.DataflowHostManagement{}
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
	newObj := &onexdatamodel.DataflowHostManagement{}
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
		obj.obj.obj.HostManagement = []*onexdatamodel.DataflowHostManagement{}
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
		obj.obj.Workload = []*onexdatamodel.DataflowWorkloadItem{}
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
	newObj := &onexdatamodel.DataflowWorkloadItem{}
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
		obj.obj.obj.Workload = []*onexdatamodel.DataflowWorkloadItem{}
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
		obj.obj.FlowProfiles = []*onexdatamodel.DataflowFlowProfile{}
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
	newObj := &onexdatamodel.DataflowFlowProfile{}
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
		obj.obj.obj.FlowProfiles = []*onexdatamodel.DataflowFlowProfile{}
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

type chaos struct {
	obj                     *onexdatamodel.Chaos
	backgroundTrafficHolder ChaosBackgroundTraffic
}

func NewChaos() Chaos {
	obj := chaos{obj: &onexdatamodel.Chaos{}}
	obj.setDefault()
	return &obj
}

func (obj *chaos) Msg() *onexdatamodel.Chaos {
	return obj.obj
}

func (obj *chaos) SetMsg(msg *onexdatamodel.Chaos) Chaos {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaos) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaos) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaos) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaos) FromYaml(value string) error {
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

func (obj *chaos) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaos) FromJson(value string) error {
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

func (obj *chaos) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaos) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaos) setNil() {
	obj.backgroundTrafficHolder = nil
}

// Chaos is description is TBD
type Chaos interface {
	Msg() *onexdatamodel.Chaos
	SetMsg(*onexdatamodel.Chaos) Chaos
	// ToPbText marshals Chaos to protobuf text
	ToPbText() string
	// ToYaml marshals Chaos to YAML text
	ToYaml() string
	// ToJson marshals Chaos to JSON text
	ToJson() string
	// FromPbText unmarshals Chaos from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals Chaos from YAML text
	FromYaml(value string) error
	// FromJson unmarshals Chaos from JSON text
	FromJson(value string) error
	// Validate validates Chaos
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// BackgroundTraffic returns ChaosBackgroundTraffic, set in Chaos.
	// ChaosBackgroundTraffic is description is TBD
	BackgroundTraffic() ChaosBackgroundTraffic
	// SetBackgroundTraffic assigns ChaosBackgroundTraffic provided by user to Chaos.
	// ChaosBackgroundTraffic is description is TBD
	SetBackgroundTraffic(value ChaosBackgroundTraffic) Chaos
	// HasBackgroundTraffic checks if BackgroundTraffic has been set in Chaos
	HasBackgroundTraffic() bool
	setNil()
}

// BackgroundTraffic returns a ChaosBackgroundTraffic
// description is TBD
func (obj *chaos) BackgroundTraffic() ChaosBackgroundTraffic {
	if obj.obj.BackgroundTraffic == nil {
		obj.obj.BackgroundTraffic = NewChaosBackgroundTraffic().Msg()
	}
	if obj.backgroundTrafficHolder == nil {
		obj.backgroundTrafficHolder = &chaosBackgroundTraffic{obj: obj.obj.BackgroundTraffic}
	}
	return obj.backgroundTrafficHolder
}

// BackgroundTraffic returns a ChaosBackgroundTraffic
// description is TBD
func (obj *chaos) HasBackgroundTraffic() bool {
	return obj.obj.BackgroundTraffic != nil
}

// SetBackgroundTraffic sets the ChaosBackgroundTraffic value in the Chaos object
// description is TBD
func (obj *chaos) SetBackgroundTraffic(value ChaosBackgroundTraffic) Chaos {

	obj.BackgroundTraffic().SetMsg(value.Msg())
	return obj
}

func (obj *chaos) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.BackgroundTraffic != nil {
		obj.BackgroundTraffic().validateObj(set_default)
	}

}

func (obj *chaos) setDefault() {

}

type errorDetails struct {
	obj *onexdatamodel.ErrorDetails
}

func NewErrorDetails() ErrorDetails {
	obj := errorDetails{obj: &onexdatamodel.ErrorDetails{}}
	obj.setDefault()
	return &obj
}

func (obj *errorDetails) Msg() *onexdatamodel.ErrorDetails {
	return obj.obj
}

func (obj *errorDetails) SetMsg(msg *onexdatamodel.ErrorDetails) ErrorDetails {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *errorDetails) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *errorDetails) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *errorDetails) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *errorDetails) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// ErrorDetails is description is TBD
type ErrorDetails interface {
	Msg() *onexdatamodel.ErrorDetails
	SetMsg(*onexdatamodel.ErrorDetails) ErrorDetails
	// ToPbText marshals ErrorDetails to protobuf text
	ToPbText() string
	// ToYaml marshals ErrorDetails to YAML text
	ToYaml() string
	// ToJson marshals ErrorDetails to JSON text
	ToJson() string
	// FromPbText unmarshals ErrorDetails from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ErrorDetails from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ErrorDetails from JSON text
	FromJson(value string) error
	// Validate validates ErrorDetails
	Validate() error
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

type warningDetails struct {
	obj *onexdatamodel.WarningDetails
}

func NewWarningDetails() WarningDetails {
	obj := warningDetails{obj: &onexdatamodel.WarningDetails{}}
	obj.setDefault()
	return &obj
}

func (obj *warningDetails) Msg() *onexdatamodel.WarningDetails {
	return obj.obj
}

func (obj *warningDetails) SetMsg(msg *onexdatamodel.WarningDetails) WarningDetails {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *warningDetails) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *warningDetails) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *warningDetails) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *warningDetails) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// WarningDetails is description is TBD
type WarningDetails interface {
	Msg() *onexdatamodel.WarningDetails
	SetMsg(*onexdatamodel.WarningDetails) WarningDetails
	// ToPbText marshals WarningDetails to protobuf text
	ToPbText() string
	// ToYaml marshals WarningDetails to YAML text
	ToYaml() string
	// ToJson marshals WarningDetails to JSON text
	ToJson() string
	// FromPbText unmarshals WarningDetails from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals WarningDetails from YAML text
	FromYaml(value string) error
	// FromJson unmarshals WarningDetails from JSON text
	FromJson(value string) error
	// Validate validates WarningDetails
	Validate() error
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

type configResponse struct {
	obj          *onexdatamodel.ConfigResponse
	configHolder Config
}

func NewConfigResponse() ConfigResponse {
	obj := configResponse{obj: &onexdatamodel.ConfigResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *configResponse) Msg() *onexdatamodel.ConfigResponse {
	return obj.obj
}

func (obj *configResponse) SetMsg(msg *onexdatamodel.ConfigResponse) ConfigResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *configResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *configResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *configResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *configResponse) FromYaml(value string) error {
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

func (obj *configResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *configResponse) FromJson(value string) error {
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

func (obj *configResponse) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *configResponse) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *configResponse) setNil() {
	obj.configHolder = nil
}

// ConfigResponse is getConfig response details
type ConfigResponse interface {
	Msg() *onexdatamodel.ConfigResponse
	SetMsg(*onexdatamodel.ConfigResponse) ConfigResponse
	// ToPbText marshals ConfigResponse to protobuf text
	ToPbText() string
	// ToYaml marshals ConfigResponse to YAML text
	ToYaml() string
	// ToJson marshals ConfigResponse to JSON text
	ToJson() string
	// FromPbText unmarshals ConfigResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ConfigResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ConfigResponse from JSON text
	FromJson(value string) error
	// Validate validates ConfigResponse
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Config returns Config, set in ConfigResponse.
	// Config is the resources that define a fabric configuration.
	Config() Config
	// SetConfig assigns Config provided by user to ConfigResponse.
	// Config is the resources that define a fabric configuration.
	SetConfig(value Config) ConfigResponse
	// HasConfig checks if Config has been set in ConfigResponse
	HasConfig() bool
	setNil()
}

// Config returns a Config
// description is TBD
func (obj *configResponse) Config() Config {
	if obj.obj.Config == nil {
		obj.obj.Config = NewConfig().Msg()
	}
	if obj.configHolder == nil {
		obj.configHolder = &config{obj: obj.obj.Config}
	}
	return obj.configHolder
}

// Config returns a Config
// description is TBD
func (obj *configResponse) HasConfig() bool {
	return obj.obj.Config != nil
}

// SetConfig sets the Config value in the ConfigResponse object
// description is TBD
func (obj *configResponse) SetConfig(value Config) ConfigResponse {

	obj.Config().SetMsg(value.Msg())
	return obj
}

func (obj *configResponse) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Config != nil {
		obj.Config().validateObj(set_default)
	}

}

func (obj *configResponse) setDefault() {

}

type metricsResponse struct {
	obj               *onexdatamodel.MetricsResponse
	flowResultsHolder MetricsResponseMetricsResponseFlowResultIter
}

func NewMetricsResponse() MetricsResponse {
	obj := metricsResponse{obj: &onexdatamodel.MetricsResponse{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponse) Msg() *onexdatamodel.MetricsResponse {
	return obj.obj
}

func (obj *metricsResponse) SetMsg(msg *onexdatamodel.MetricsResponse) MetricsResponse {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponse) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *metricsResponse) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *metricsResponse) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsResponse) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsResponse) setNil() {
	obj.flowResultsHolder = nil
}

// MetricsResponse is metrics response details
type MetricsResponse interface {
	Msg() *onexdatamodel.MetricsResponse
	SetMsg(*onexdatamodel.MetricsResponse) MetricsResponse
	// ToPbText marshals MetricsResponse to protobuf text
	ToPbText() string
	// ToYaml marshals MetricsResponse to YAML text
	ToYaml() string
	// ToJson marshals MetricsResponse to JSON text
	ToJson() string
	// FromPbText unmarshals MetricsResponse from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponse from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponse from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponse
	Validate() error
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
// job completion time
func (obj *metricsResponse) Jct() float32 {

	return *obj.obj.Jct

}

// Jct returns a float32
// job completion time
func (obj *metricsResponse) HasJct() bool {
	return obj.obj.Jct != nil
}

// SetJct sets the float32 value in the MetricsResponse object
// job completion time
func (obj *metricsResponse) SetJct(value float32) MetricsResponse {

	obj.obj.Jct = &value
	return obj
}

// FlowResults returns a []MetricsResponseFlowResult
// description is TBD
func (obj *metricsResponse) FlowResults() MetricsResponseMetricsResponseFlowResultIter {
	if len(obj.obj.FlowResults) == 0 {
		obj.obj.FlowResults = []*onexdatamodel.MetricsResponseFlowResult{}
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
	newObj := &onexdatamodel.MetricsResponseFlowResult{}
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
		obj.obj.obj.FlowResults = []*onexdatamodel.MetricsResponseFlowResult{}
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

type fabricSpinePodRack struct {
	obj                *onexdatamodel.FabricSpinePodRack
	spinesHolder       FabricSpinePodRackFabricSpineIter
	podsHolder         FabricSpinePodRackFabricPodIter
	hostLinksHolder    FabricSpinePodRackSwitchHostLinkIter
	podProfilesHolder  FabricSpinePodRackFabricPodProfileIter
	rackProfilesHolder FabricSpinePodRackFabricRackProfileIter
}

func NewFabricSpinePodRack() FabricSpinePodRack {
	obj := fabricSpinePodRack{obj: &onexdatamodel.FabricSpinePodRack{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricSpinePodRack) Msg() *onexdatamodel.FabricSpinePodRack {
	return obj.obj
}

func (obj *fabricSpinePodRack) SetMsg(msg *onexdatamodel.FabricSpinePodRack) FabricSpinePodRack {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricSpinePodRack) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricSpinePodRack) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *fabricSpinePodRack) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricSpinePodRack) FromYaml(value string) error {
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

func (obj *fabricSpinePodRack) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricSpinePodRack) FromJson(value string) error {
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

func (obj *fabricSpinePodRack) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricSpinePodRack) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *fabricSpinePodRack) setNil() {
	obj.spinesHolder = nil
	obj.podsHolder = nil
	obj.hostLinksHolder = nil
	obj.podProfilesHolder = nil
	obj.rackProfilesHolder = nil
}

// FabricSpinePodRack is an emulation of a spine/pod/rack topology
type FabricSpinePodRack interface {
	Msg() *onexdatamodel.FabricSpinePodRack
	SetMsg(*onexdatamodel.FabricSpinePodRack) FabricSpinePodRack
	// ToPbText marshals FabricSpinePodRack to protobuf text
	ToPbText() string
	// ToYaml marshals FabricSpinePodRack to YAML text
	ToYaml() string
	// ToJson marshals FabricSpinePodRack to JSON text
	ToJson() string
	// FromPbText unmarshals FabricSpinePodRack from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricSpinePodRack from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricSpinePodRack from JSON text
	FromJson(value string) error
	// Validate validates FabricSpinePodRack
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Spines returns FabricSpinePodRackFabricSpineIter, set in FabricSpinePodRack
	Spines() FabricSpinePodRackFabricSpineIter
	// Pods returns FabricSpinePodRackFabricPodIter, set in FabricSpinePodRack
	Pods() FabricSpinePodRackFabricPodIter
	// HostLinks returns FabricSpinePodRackSwitchHostLinkIter, set in FabricSpinePodRack
	HostLinks() FabricSpinePodRackSwitchHostLinkIter
	// PodProfiles returns FabricSpinePodRackFabricPodProfileIter, set in FabricSpinePodRack
	PodProfiles() FabricSpinePodRackFabricPodProfileIter
	// RackProfiles returns FabricSpinePodRackFabricRackProfileIter, set in FabricSpinePodRack
	RackProfiles() FabricSpinePodRackFabricRackProfileIter
	setNil()
}

// Spines returns a []FabricSpine
// The spines in the topology.
func (obj *fabricSpinePodRack) Spines() FabricSpinePodRackFabricSpineIter {
	if len(obj.obj.Spines) == 0 {
		obj.obj.Spines = []*onexdatamodel.FabricSpine{}
	}
	if obj.spinesHolder == nil {
		obj.spinesHolder = newFabricSpinePodRackFabricSpineIter().setMsg(obj)
	}
	return obj.spinesHolder
}

type fabricSpinePodRackFabricSpineIter struct {
	obj              *fabricSpinePodRack
	fabricSpineSlice []FabricSpine
}

func newFabricSpinePodRackFabricSpineIter() FabricSpinePodRackFabricSpineIter {
	return &fabricSpinePodRackFabricSpineIter{}
}

type FabricSpinePodRackFabricSpineIter interface {
	setMsg(*fabricSpinePodRack) FabricSpinePodRackFabricSpineIter
	Items() []FabricSpine
	Add() FabricSpine
	Append(items ...FabricSpine) FabricSpinePodRackFabricSpineIter
	Set(index int, newObj FabricSpine) FabricSpinePodRackFabricSpineIter
	Clear() FabricSpinePodRackFabricSpineIter
	clearHolderSlice() FabricSpinePodRackFabricSpineIter
	appendHolderSlice(item FabricSpine) FabricSpinePodRackFabricSpineIter
}

func (obj *fabricSpinePodRackFabricSpineIter) setMsg(msg *fabricSpinePodRack) FabricSpinePodRackFabricSpineIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Spines {
		obj.appendHolderSlice(&fabricSpine{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricSpinePodRackFabricSpineIter) Items() []FabricSpine {
	return obj.fabricSpineSlice
}

func (obj *fabricSpinePodRackFabricSpineIter) Add() FabricSpine {
	newObj := &onexdatamodel.FabricSpine{}
	obj.obj.obj.Spines = append(obj.obj.obj.Spines, newObj)
	newLibObj := &fabricSpine{obj: newObj}
	newLibObj.setDefault()
	obj.fabricSpineSlice = append(obj.fabricSpineSlice, newLibObj)
	return newLibObj
}

func (obj *fabricSpinePodRackFabricSpineIter) Append(items ...FabricSpine) FabricSpinePodRackFabricSpineIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Spines = append(obj.obj.obj.Spines, newObj)
		obj.fabricSpineSlice = append(obj.fabricSpineSlice, item)
	}
	return obj
}

func (obj *fabricSpinePodRackFabricSpineIter) Set(index int, newObj FabricSpine) FabricSpinePodRackFabricSpineIter {
	obj.obj.obj.Spines[index] = newObj.Msg()
	obj.fabricSpineSlice[index] = newObj
	return obj
}
func (obj *fabricSpinePodRackFabricSpineIter) Clear() FabricSpinePodRackFabricSpineIter {
	if len(obj.obj.obj.Spines) > 0 {
		obj.obj.obj.Spines = []*onexdatamodel.FabricSpine{}
		obj.fabricSpineSlice = []FabricSpine{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricSpineIter) clearHolderSlice() FabricSpinePodRackFabricSpineIter {
	if len(obj.fabricSpineSlice) > 0 {
		obj.fabricSpineSlice = []FabricSpine{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricSpineIter) appendHolderSlice(item FabricSpine) FabricSpinePodRackFabricSpineIter {
	obj.fabricSpineSlice = append(obj.fabricSpineSlice, item)
	return obj
}

// Pods returns a []FabricPod
// The pods in the topology.
func (obj *fabricSpinePodRack) Pods() FabricSpinePodRackFabricPodIter {
	if len(obj.obj.Pods) == 0 {
		obj.obj.Pods = []*onexdatamodel.FabricPod{}
	}
	if obj.podsHolder == nil {
		obj.podsHolder = newFabricSpinePodRackFabricPodIter().setMsg(obj)
	}
	return obj.podsHolder
}

type fabricSpinePodRackFabricPodIter struct {
	obj            *fabricSpinePodRack
	fabricPodSlice []FabricPod
}

func newFabricSpinePodRackFabricPodIter() FabricSpinePodRackFabricPodIter {
	return &fabricSpinePodRackFabricPodIter{}
}

type FabricSpinePodRackFabricPodIter interface {
	setMsg(*fabricSpinePodRack) FabricSpinePodRackFabricPodIter
	Items() []FabricPod
	Add() FabricPod
	Append(items ...FabricPod) FabricSpinePodRackFabricPodIter
	Set(index int, newObj FabricPod) FabricSpinePodRackFabricPodIter
	Clear() FabricSpinePodRackFabricPodIter
	clearHolderSlice() FabricSpinePodRackFabricPodIter
	appendHolderSlice(item FabricPod) FabricSpinePodRackFabricPodIter
}

func (obj *fabricSpinePodRackFabricPodIter) setMsg(msg *fabricSpinePodRack) FabricSpinePodRackFabricPodIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Pods {
		obj.appendHolderSlice(&fabricPod{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricSpinePodRackFabricPodIter) Items() []FabricPod {
	return obj.fabricPodSlice
}

func (obj *fabricSpinePodRackFabricPodIter) Add() FabricPod {
	newObj := &onexdatamodel.FabricPod{}
	obj.obj.obj.Pods = append(obj.obj.obj.Pods, newObj)
	newLibObj := &fabricPod{obj: newObj}
	newLibObj.setDefault()
	obj.fabricPodSlice = append(obj.fabricPodSlice, newLibObj)
	return newLibObj
}

func (obj *fabricSpinePodRackFabricPodIter) Append(items ...FabricPod) FabricSpinePodRackFabricPodIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Pods = append(obj.obj.obj.Pods, newObj)
		obj.fabricPodSlice = append(obj.fabricPodSlice, item)
	}
	return obj
}

func (obj *fabricSpinePodRackFabricPodIter) Set(index int, newObj FabricPod) FabricSpinePodRackFabricPodIter {
	obj.obj.obj.Pods[index] = newObj.Msg()
	obj.fabricPodSlice[index] = newObj
	return obj
}
func (obj *fabricSpinePodRackFabricPodIter) Clear() FabricSpinePodRackFabricPodIter {
	if len(obj.obj.obj.Pods) > 0 {
		obj.obj.obj.Pods = []*onexdatamodel.FabricPod{}
		obj.fabricPodSlice = []FabricPod{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricPodIter) clearHolderSlice() FabricSpinePodRackFabricPodIter {
	if len(obj.fabricPodSlice) > 0 {
		obj.fabricPodSlice = []FabricPod{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricPodIter) appendHolderSlice(item FabricPod) FabricSpinePodRackFabricPodIter {
	obj.fabricPodSlice = append(obj.fabricPodSlice, item)
	return obj
}

// HostLinks returns a []SwitchHostLink
// description is TBD
func (obj *fabricSpinePodRack) HostLinks() FabricSpinePodRackSwitchHostLinkIter {
	if len(obj.obj.HostLinks) == 0 {
		obj.obj.HostLinks = []*onexdatamodel.SwitchHostLink{}
	}
	if obj.hostLinksHolder == nil {
		obj.hostLinksHolder = newFabricSpinePodRackSwitchHostLinkIter().setMsg(obj)
	}
	return obj.hostLinksHolder
}

type fabricSpinePodRackSwitchHostLinkIter struct {
	obj                 *fabricSpinePodRack
	switchHostLinkSlice []SwitchHostLink
}

func newFabricSpinePodRackSwitchHostLinkIter() FabricSpinePodRackSwitchHostLinkIter {
	return &fabricSpinePodRackSwitchHostLinkIter{}
}

type FabricSpinePodRackSwitchHostLinkIter interface {
	setMsg(*fabricSpinePodRack) FabricSpinePodRackSwitchHostLinkIter
	Items() []SwitchHostLink
	Add() SwitchHostLink
	Append(items ...SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter
	Set(index int, newObj SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter
	Clear() FabricSpinePodRackSwitchHostLinkIter
	clearHolderSlice() FabricSpinePodRackSwitchHostLinkIter
	appendHolderSlice(item SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter
}

func (obj *fabricSpinePodRackSwitchHostLinkIter) setMsg(msg *fabricSpinePodRack) FabricSpinePodRackSwitchHostLinkIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.HostLinks {
		obj.appendHolderSlice(&switchHostLink{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricSpinePodRackSwitchHostLinkIter) Items() []SwitchHostLink {
	return obj.switchHostLinkSlice
}

func (obj *fabricSpinePodRackSwitchHostLinkIter) Add() SwitchHostLink {
	newObj := &onexdatamodel.SwitchHostLink{}
	obj.obj.obj.HostLinks = append(obj.obj.obj.HostLinks, newObj)
	newLibObj := &switchHostLink{obj: newObj}
	newLibObj.setDefault()
	obj.switchHostLinkSlice = append(obj.switchHostLinkSlice, newLibObj)
	return newLibObj
}

func (obj *fabricSpinePodRackSwitchHostLinkIter) Append(items ...SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.HostLinks = append(obj.obj.obj.HostLinks, newObj)
		obj.switchHostLinkSlice = append(obj.switchHostLinkSlice, item)
	}
	return obj
}

func (obj *fabricSpinePodRackSwitchHostLinkIter) Set(index int, newObj SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter {
	obj.obj.obj.HostLinks[index] = newObj.Msg()
	obj.switchHostLinkSlice[index] = newObj
	return obj
}
func (obj *fabricSpinePodRackSwitchHostLinkIter) Clear() FabricSpinePodRackSwitchHostLinkIter {
	if len(obj.obj.obj.HostLinks) > 0 {
		obj.obj.obj.HostLinks = []*onexdatamodel.SwitchHostLink{}
		obj.switchHostLinkSlice = []SwitchHostLink{}
	}
	return obj
}
func (obj *fabricSpinePodRackSwitchHostLinkIter) clearHolderSlice() FabricSpinePodRackSwitchHostLinkIter {
	if len(obj.switchHostLinkSlice) > 0 {
		obj.switchHostLinkSlice = []SwitchHostLink{}
	}
	return obj
}
func (obj *fabricSpinePodRackSwitchHostLinkIter) appendHolderSlice(item SwitchHostLink) FabricSpinePodRackSwitchHostLinkIter {
	obj.switchHostLinkSlice = append(obj.switchHostLinkSlice, item)
	return obj
}

// PodProfiles returns a []FabricPodProfile
// A list of pod profiles
func (obj *fabricSpinePodRack) PodProfiles() FabricSpinePodRackFabricPodProfileIter {
	if len(obj.obj.PodProfiles) == 0 {
		obj.obj.PodProfiles = []*onexdatamodel.FabricPodProfile{}
	}
	if obj.podProfilesHolder == nil {
		obj.podProfilesHolder = newFabricSpinePodRackFabricPodProfileIter().setMsg(obj)
	}
	return obj.podProfilesHolder
}

type fabricSpinePodRackFabricPodProfileIter struct {
	obj                   *fabricSpinePodRack
	fabricPodProfileSlice []FabricPodProfile
}

func newFabricSpinePodRackFabricPodProfileIter() FabricSpinePodRackFabricPodProfileIter {
	return &fabricSpinePodRackFabricPodProfileIter{}
}

type FabricSpinePodRackFabricPodProfileIter interface {
	setMsg(*fabricSpinePodRack) FabricSpinePodRackFabricPodProfileIter
	Items() []FabricPodProfile
	Add() FabricPodProfile
	Append(items ...FabricPodProfile) FabricSpinePodRackFabricPodProfileIter
	Set(index int, newObj FabricPodProfile) FabricSpinePodRackFabricPodProfileIter
	Clear() FabricSpinePodRackFabricPodProfileIter
	clearHolderSlice() FabricSpinePodRackFabricPodProfileIter
	appendHolderSlice(item FabricPodProfile) FabricSpinePodRackFabricPodProfileIter
}

func (obj *fabricSpinePodRackFabricPodProfileIter) setMsg(msg *fabricSpinePodRack) FabricSpinePodRackFabricPodProfileIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.PodProfiles {
		obj.appendHolderSlice(&fabricPodProfile{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricSpinePodRackFabricPodProfileIter) Items() []FabricPodProfile {
	return obj.fabricPodProfileSlice
}

func (obj *fabricSpinePodRackFabricPodProfileIter) Add() FabricPodProfile {
	newObj := &onexdatamodel.FabricPodProfile{}
	obj.obj.obj.PodProfiles = append(obj.obj.obj.PodProfiles, newObj)
	newLibObj := &fabricPodProfile{obj: newObj}
	newLibObj.setDefault()
	obj.fabricPodProfileSlice = append(obj.fabricPodProfileSlice, newLibObj)
	return newLibObj
}

func (obj *fabricSpinePodRackFabricPodProfileIter) Append(items ...FabricPodProfile) FabricSpinePodRackFabricPodProfileIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.PodProfiles = append(obj.obj.obj.PodProfiles, newObj)
		obj.fabricPodProfileSlice = append(obj.fabricPodProfileSlice, item)
	}
	return obj
}

func (obj *fabricSpinePodRackFabricPodProfileIter) Set(index int, newObj FabricPodProfile) FabricSpinePodRackFabricPodProfileIter {
	obj.obj.obj.PodProfiles[index] = newObj.Msg()
	obj.fabricPodProfileSlice[index] = newObj
	return obj
}
func (obj *fabricSpinePodRackFabricPodProfileIter) Clear() FabricSpinePodRackFabricPodProfileIter {
	if len(obj.obj.obj.PodProfiles) > 0 {
		obj.obj.obj.PodProfiles = []*onexdatamodel.FabricPodProfile{}
		obj.fabricPodProfileSlice = []FabricPodProfile{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricPodProfileIter) clearHolderSlice() FabricSpinePodRackFabricPodProfileIter {
	if len(obj.fabricPodProfileSlice) > 0 {
		obj.fabricPodProfileSlice = []FabricPodProfile{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricPodProfileIter) appendHolderSlice(item FabricPodProfile) FabricSpinePodRackFabricPodProfileIter {
	obj.fabricPodProfileSlice = append(obj.fabricPodProfileSlice, item)
	return obj
}

// RackProfiles returns a []FabricRackProfile
// A list of rack profiles
func (obj *fabricSpinePodRack) RackProfiles() FabricSpinePodRackFabricRackProfileIter {
	if len(obj.obj.RackProfiles) == 0 {
		obj.obj.RackProfiles = []*onexdatamodel.FabricRackProfile{}
	}
	if obj.rackProfilesHolder == nil {
		obj.rackProfilesHolder = newFabricSpinePodRackFabricRackProfileIter().setMsg(obj)
	}
	return obj.rackProfilesHolder
}

type fabricSpinePodRackFabricRackProfileIter struct {
	obj                    *fabricSpinePodRack
	fabricRackProfileSlice []FabricRackProfile
}

func newFabricSpinePodRackFabricRackProfileIter() FabricSpinePodRackFabricRackProfileIter {
	return &fabricSpinePodRackFabricRackProfileIter{}
}

type FabricSpinePodRackFabricRackProfileIter interface {
	setMsg(*fabricSpinePodRack) FabricSpinePodRackFabricRackProfileIter
	Items() []FabricRackProfile
	Add() FabricRackProfile
	Append(items ...FabricRackProfile) FabricSpinePodRackFabricRackProfileIter
	Set(index int, newObj FabricRackProfile) FabricSpinePodRackFabricRackProfileIter
	Clear() FabricSpinePodRackFabricRackProfileIter
	clearHolderSlice() FabricSpinePodRackFabricRackProfileIter
	appendHolderSlice(item FabricRackProfile) FabricSpinePodRackFabricRackProfileIter
}

func (obj *fabricSpinePodRackFabricRackProfileIter) setMsg(msg *fabricSpinePodRack) FabricSpinePodRackFabricRackProfileIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.RackProfiles {
		obj.appendHolderSlice(&fabricRackProfile{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *fabricSpinePodRackFabricRackProfileIter) Items() []FabricRackProfile {
	return obj.fabricRackProfileSlice
}

func (obj *fabricSpinePodRackFabricRackProfileIter) Add() FabricRackProfile {
	newObj := &onexdatamodel.FabricRackProfile{}
	obj.obj.obj.RackProfiles = append(obj.obj.obj.RackProfiles, newObj)
	newLibObj := &fabricRackProfile{obj: newObj}
	newLibObj.setDefault()
	obj.fabricRackProfileSlice = append(obj.fabricRackProfileSlice, newLibObj)
	return newLibObj
}

func (obj *fabricSpinePodRackFabricRackProfileIter) Append(items ...FabricRackProfile) FabricSpinePodRackFabricRackProfileIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.RackProfiles = append(obj.obj.obj.RackProfiles, newObj)
		obj.fabricRackProfileSlice = append(obj.fabricRackProfileSlice, item)
	}
	return obj
}

func (obj *fabricSpinePodRackFabricRackProfileIter) Set(index int, newObj FabricRackProfile) FabricSpinePodRackFabricRackProfileIter {
	obj.obj.obj.RackProfiles[index] = newObj.Msg()
	obj.fabricRackProfileSlice[index] = newObj
	return obj
}
func (obj *fabricSpinePodRackFabricRackProfileIter) Clear() FabricSpinePodRackFabricRackProfileIter {
	if len(obj.obj.obj.RackProfiles) > 0 {
		obj.obj.obj.RackProfiles = []*onexdatamodel.FabricRackProfile{}
		obj.fabricRackProfileSlice = []FabricRackProfile{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricRackProfileIter) clearHolderSlice() FabricSpinePodRackFabricRackProfileIter {
	if len(obj.fabricRackProfileSlice) > 0 {
		obj.fabricRackProfileSlice = []FabricRackProfile{}
	}
	return obj
}
func (obj *fabricSpinePodRackFabricRackProfileIter) appendHolderSlice(item FabricRackProfile) FabricSpinePodRackFabricRackProfileIter {
	obj.fabricRackProfileSlice = append(obj.fabricRackProfileSlice, item)
	return obj
}

func (obj *fabricSpinePodRack) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Spines != nil {

		if set_default {
			obj.Spines().clearHolderSlice()
			for _, item := range obj.obj.Spines {
				obj.Spines().appendHolderSlice(&fabricSpine{obj: item})
			}
		}
		for _, item := range obj.Spines().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.Pods != nil {

		if set_default {
			obj.Pods().clearHolderSlice()
			for _, item := range obj.obj.Pods {
				obj.Pods().appendHolderSlice(&fabricPod{obj: item})
			}
		}
		for _, item := range obj.Pods().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.HostLinks != nil {

		if set_default {
			obj.HostLinks().clearHolderSlice()
			for _, item := range obj.obj.HostLinks {
				obj.HostLinks().appendHolderSlice(&switchHostLink{obj: item})
			}
		}
		for _, item := range obj.HostLinks().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.PodProfiles != nil {

		if set_default {
			obj.PodProfiles().clearHolderSlice()
			for _, item := range obj.obj.PodProfiles {
				obj.PodProfiles().appendHolderSlice(&fabricPodProfile{obj: item})
			}
		}
		for _, item := range obj.PodProfiles().Items() {
			item.validateObj(set_default)
		}

	}

	if obj.obj.RackProfiles != nil {

		if set_default {
			obj.RackProfiles().clearHolderSlice()
			for _, item := range obj.obj.RackProfiles {
				obj.RackProfiles().appendHolderSlice(&fabricRackProfile{obj: item})
			}
		}
		for _, item := range obj.RackProfiles().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *fabricSpinePodRack) setDefault() {

}

type fabricQosProfile struct {
	obj                        *onexdatamodel.FabricQosProfile
	ingressAdmissionHolder     FabricQosProfileIngressAdmission
	schedulerHolder            FabricQosProfileScheduler
	packetClassificationHolder FabricQosProfilePacketClassification
}

func NewFabricQosProfile() FabricQosProfile {
	obj := fabricQosProfile{obj: &onexdatamodel.FabricQosProfile{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricQosProfile) Msg() *onexdatamodel.FabricQosProfile {
	return obj.obj
}

func (obj *fabricQosProfile) SetMsg(msg *onexdatamodel.FabricQosProfile) FabricQosProfile {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricQosProfile) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricQosProfile) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *fabricQosProfile) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfile) FromYaml(value string) error {
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

func (obj *fabricQosProfile) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfile) FromJson(value string) error {
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

func (obj *fabricQosProfile) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricQosProfile) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *fabricQosProfile) setNil() {
	obj.ingressAdmissionHolder = nil
	obj.schedulerHolder = nil
	obj.packetClassificationHolder = nil
}

// FabricQosProfile is description is TBD
type FabricQosProfile interface {
	Msg() *onexdatamodel.FabricQosProfile
	SetMsg(*onexdatamodel.FabricQosProfile) FabricQosProfile
	// ToPbText marshals FabricQosProfile to protobuf text
	ToPbText() string
	// ToYaml marshals FabricQosProfile to YAML text
	ToYaml() string
	// ToJson marshals FabricQosProfile to JSON text
	ToJson() string
	// FromPbText unmarshals FabricQosProfile from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricQosProfile from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricQosProfile from JSON text
	FromJson(value string) error
	// Validate validates FabricQosProfile
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in FabricQosProfile.
	Name() string
	// SetName assigns string provided by user to FabricQosProfile
	SetName(value string) FabricQosProfile
	// IngressAdmission returns FabricQosProfileIngressAdmission, set in FabricQosProfile.
	// FabricQosProfileIngressAdmission is description is TBD
	IngressAdmission() FabricQosProfileIngressAdmission
	// SetIngressAdmission assigns FabricQosProfileIngressAdmission provided by user to FabricQosProfile.
	// FabricQosProfileIngressAdmission is description is TBD
	SetIngressAdmission(value FabricQosProfileIngressAdmission) FabricQosProfile
	// HasIngressAdmission checks if IngressAdmission has been set in FabricQosProfile
	HasIngressAdmission() bool
	// Scheduler returns FabricQosProfileScheduler, set in FabricQosProfile.
	// FabricQosProfileScheduler is description is TBD
	Scheduler() FabricQosProfileScheduler
	// SetScheduler assigns FabricQosProfileScheduler provided by user to FabricQosProfile.
	// FabricQosProfileScheduler is description is TBD
	SetScheduler(value FabricQosProfileScheduler) FabricQosProfile
	// HasScheduler checks if Scheduler has been set in FabricQosProfile
	HasScheduler() bool
	// PacketClassification returns FabricQosProfilePacketClassification, set in FabricQosProfile.
	// FabricQosProfilePacketClassification is description is TBD
	PacketClassification() FabricQosProfilePacketClassification
	// SetPacketClassification assigns FabricQosProfilePacketClassification provided by user to FabricQosProfile.
	// FabricQosProfilePacketClassification is description is TBD
	SetPacketClassification(value FabricQosProfilePacketClassification) FabricQosProfile
	// HasPacketClassification checks if PacketClassification has been set in FabricQosProfile
	HasPacketClassification() bool
	setNil()
}

// Name returns a string
// description is TBD
func (obj *fabricQosProfile) Name() string {

	return obj.obj.Name
}

// SetName sets the string value in the FabricQosProfile object
// description is TBD
func (obj *fabricQosProfile) SetName(value string) FabricQosProfile {

	obj.obj.Name = value
	return obj
}

// IngressAdmission returns a FabricQosProfileIngressAdmission
// description is TBD
func (obj *fabricQosProfile) IngressAdmission() FabricQosProfileIngressAdmission {
	if obj.obj.IngressAdmission == nil {
		obj.obj.IngressAdmission = NewFabricQosProfileIngressAdmission().Msg()
	}
	if obj.ingressAdmissionHolder == nil {
		obj.ingressAdmissionHolder = &fabricQosProfileIngressAdmission{obj: obj.obj.IngressAdmission}
	}
	return obj.ingressAdmissionHolder
}

// IngressAdmission returns a FabricQosProfileIngressAdmission
// description is TBD
func (obj *fabricQosProfile) HasIngressAdmission() bool {
	return obj.obj.IngressAdmission != nil
}

// SetIngressAdmission sets the FabricQosProfileIngressAdmission value in the FabricQosProfile object
// description is TBD
func (obj *fabricQosProfile) SetIngressAdmission(value FabricQosProfileIngressAdmission) FabricQosProfile {

	obj.IngressAdmission().SetMsg(value.Msg())
	return obj
}

// Scheduler returns a FabricQosProfileScheduler
// description is TBD
func (obj *fabricQosProfile) Scheduler() FabricQosProfileScheduler {
	if obj.obj.Scheduler == nil {
		obj.obj.Scheduler = NewFabricQosProfileScheduler().Msg()
	}
	if obj.schedulerHolder == nil {
		obj.schedulerHolder = &fabricQosProfileScheduler{obj: obj.obj.Scheduler}
	}
	return obj.schedulerHolder
}

// Scheduler returns a FabricQosProfileScheduler
// description is TBD
func (obj *fabricQosProfile) HasScheduler() bool {
	return obj.obj.Scheduler != nil
}

// SetScheduler sets the FabricQosProfileScheduler value in the FabricQosProfile object
// description is TBD
func (obj *fabricQosProfile) SetScheduler(value FabricQosProfileScheduler) FabricQosProfile {

	obj.Scheduler().SetMsg(value.Msg())
	return obj
}

// PacketClassification returns a FabricQosProfilePacketClassification
// description is TBD
func (obj *fabricQosProfile) PacketClassification() FabricQosProfilePacketClassification {
	if obj.obj.PacketClassification == nil {
		obj.obj.PacketClassification = NewFabricQosProfilePacketClassification().Msg()
	}
	if obj.packetClassificationHolder == nil {
		obj.packetClassificationHolder = &fabricQosProfilePacketClassification{obj: obj.obj.PacketClassification}
	}
	return obj.packetClassificationHolder
}

// PacketClassification returns a FabricQosProfilePacketClassification
// description is TBD
func (obj *fabricQosProfile) HasPacketClassification() bool {
	return obj.obj.PacketClassification != nil
}

// SetPacketClassification sets the FabricQosProfilePacketClassification value in the FabricQosProfile object
// description is TBD
func (obj *fabricQosProfile) SetPacketClassification(value FabricQosProfilePacketClassification) FabricQosProfile {

	obj.PacketClassification().SetMsg(value.Msg())
	return obj
}

func (obj *fabricQosProfile) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// Name is required
	if obj.obj.Name == "" {
		validation = append(validation, "Name is required field on interface FabricQosProfile")
	}

	if obj.obj.IngressAdmission != nil {
		obj.IngressAdmission().validateObj(set_default)
	}

	if obj.obj.Scheduler != nil {
		obj.Scheduler().validateObj(set_default)
	}

	if obj.obj.PacketClassification != nil {
		obj.PacketClassification().validateObj(set_default)
	}

}

func (obj *fabricQosProfile) setDefault() {

}

type dataflowHostManagement struct {
	obj *onexdatamodel.DataflowHostManagement
}

func NewDataflowHostManagement() DataflowHostManagement {
	obj := dataflowHostManagement{obj: &onexdatamodel.DataflowHostManagement{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowHostManagement) Msg() *onexdatamodel.DataflowHostManagement {
	return obj.obj
}

func (obj *dataflowHostManagement) SetMsg(msg *onexdatamodel.DataflowHostManagement) DataflowHostManagement {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowHostManagement) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowHostManagement) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowHostManagement) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowHostManagement) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowHostManagement is auxillary host information needed to run dataflow experiments
type DataflowHostManagement interface {
	Msg() *onexdatamodel.DataflowHostManagement
	SetMsg(*onexdatamodel.DataflowHostManagement) DataflowHostManagement
	// ToPbText marshals DataflowHostManagement to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowHostManagement to YAML text
	ToYaml() string
	// ToJson marshals DataflowHostManagement to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowHostManagement from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowHostManagement from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowHostManagement from JSON text
	FromJson(value string) error
	// Validate validates DataflowHostManagement
	Validate() error
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

type dataflowWorkloadItem struct {
	obj             *onexdatamodel.DataflowWorkloadItem
	scatterHolder   DataflowScatterWorkload
	gatherHolder    DataflowGatherWorkload
	loopHolder      DataflowLoopWorkload
	computeHolder   DataflowComputeWorkload
	allReduceHolder DataflowAllReduceWorkload
	broadcastHolder DataflowBroadcastWorkload
}

func NewDataflowWorkloadItem() DataflowWorkloadItem {
	obj := dataflowWorkloadItem{obj: &onexdatamodel.DataflowWorkloadItem{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowWorkloadItem) Msg() *onexdatamodel.DataflowWorkloadItem {
	return obj.obj
}

func (obj *dataflowWorkloadItem) SetMsg(msg *onexdatamodel.DataflowWorkloadItem) DataflowWorkloadItem {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowWorkloadItem) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowWorkloadItem) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflowWorkloadItem) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowWorkloadItem) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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
	Msg() *onexdatamodel.DataflowWorkloadItem
	SetMsg(*onexdatamodel.DataflowWorkloadItem) DataflowWorkloadItem
	// ToPbText marshals DataflowWorkloadItem to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowWorkloadItem to YAML text
	ToYaml() string
	// ToJson marshals DataflowWorkloadItem to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowWorkloadItem from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowWorkloadItem from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowWorkloadItem from JSON text
	FromJson(value string) error
	// Validate validates DataflowWorkloadItem
	Validate() error
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
	intValue, ok := onexdatamodel.DataflowWorkloadItem_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowWorkloadItemChoiceEnum", string(value)))
		return obj
	}
	obj.obj.Choice = onexdatamodel.DataflowWorkloadItem_Choice_Enum(intValue)
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
	obj.Scatter().SetMsg(value.Msg())
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
	obj.Gather().SetMsg(value.Msg())
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
	obj.Loop().SetMsg(value.Msg())
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
	obj.Compute().SetMsg(value.Msg())
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
	obj.AllReduce().SetMsg(value.Msg())
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
	obj.Broadcast().SetMsg(value.Msg())
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

type dataflowFlowProfile struct {
	obj            *onexdatamodel.DataflowFlowProfile
	ethernetHolder DataflowFlowProfileEthernet
	tcpHolder      DataflowFlowProfileTcp
	udpHolder      DataflowFlowProfileUdp
}

func NewDataflowFlowProfile() DataflowFlowProfile {
	obj := dataflowFlowProfile{obj: &onexdatamodel.DataflowFlowProfile{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfile) Msg() *onexdatamodel.DataflowFlowProfile {
	return obj.obj
}

func (obj *dataflowFlowProfile) SetMsg(msg *onexdatamodel.DataflowFlowProfile) DataflowFlowProfile {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfile) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowFlowProfile) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflowFlowProfile) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfile) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfile) setNil() {
	obj.ethernetHolder = nil
	obj.tcpHolder = nil
	obj.udpHolder = nil
}

// DataflowFlowProfile is description is TBD
type DataflowFlowProfile interface {
	Msg() *onexdatamodel.DataflowFlowProfile
	SetMsg(*onexdatamodel.DataflowFlowProfile) DataflowFlowProfile
	// ToPbText marshals DataflowFlowProfile to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowFlowProfile to YAML text
	ToYaml() string
	// ToJson marshals DataflowFlowProfile to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowFlowProfile from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfile from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfile from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfile
	Validate() error
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
	intValue, ok := onexdatamodel.DataflowFlowProfile_L2ProtocolChoice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileL2ProtocolChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.DataflowFlowProfile_L2ProtocolChoice_Enum(intValue)
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

	obj.Ethernet().SetMsg(value.Msg())
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
	intValue, ok := onexdatamodel.DataflowFlowProfile_L4ProtocolChoice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileL4ProtocolChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.DataflowFlowProfile_L4ProtocolChoice_Enum(intValue)
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

	obj.Tcp().SetMsg(value.Msg())
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

	obj.Udp().SetMsg(value.Msg())
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

type chaosBackgroundTraffic struct {
	obj         *onexdatamodel.ChaosBackgroundTraffic
	flowsHolder ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
}

func NewChaosBackgroundTraffic() ChaosBackgroundTraffic {
	obj := chaosBackgroundTraffic{obj: &onexdatamodel.ChaosBackgroundTraffic{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTraffic) Msg() *onexdatamodel.ChaosBackgroundTraffic {
	return obj.obj
}

func (obj *chaosBackgroundTraffic) SetMsg(msg *onexdatamodel.ChaosBackgroundTraffic) ChaosBackgroundTraffic {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTraffic) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTraffic) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaosBackgroundTraffic) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTraffic) FromYaml(value string) error {
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

func (obj *chaosBackgroundTraffic) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTraffic) FromJson(value string) error {
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

func (obj *chaosBackgroundTraffic) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTraffic) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaosBackgroundTraffic) setNil() {
	obj.flowsHolder = nil
}

// ChaosBackgroundTraffic is description is TBD
type ChaosBackgroundTraffic interface {
	Msg() *onexdatamodel.ChaosBackgroundTraffic
	SetMsg(*onexdatamodel.ChaosBackgroundTraffic) ChaosBackgroundTraffic
	// ToPbText marshals ChaosBackgroundTraffic to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTraffic to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTraffic to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTraffic from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTraffic from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTraffic from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTraffic
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Flows returns ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter, set in ChaosBackgroundTraffic
	Flows() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	setNil()
}

// Flows returns a []ChaosBackgroundTrafficFlow
// description is TBD
func (obj *chaosBackgroundTraffic) Flows() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	if len(obj.obj.Flows) == 0 {
		obj.obj.Flows = []*onexdatamodel.ChaosBackgroundTrafficFlow{}
	}
	if obj.flowsHolder == nil {
		obj.flowsHolder = newChaosBackgroundTrafficChaosBackgroundTrafficFlowIter().setMsg(obj)
	}
	return obj.flowsHolder
}

type chaosBackgroundTrafficChaosBackgroundTrafficFlowIter struct {
	obj                             *chaosBackgroundTraffic
	chaosBackgroundTrafficFlowSlice []ChaosBackgroundTrafficFlow
}

func newChaosBackgroundTrafficChaosBackgroundTrafficFlowIter() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	return &chaosBackgroundTrafficChaosBackgroundTrafficFlowIter{}
}

type ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter interface {
	setMsg(*chaosBackgroundTraffic) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	Items() []ChaosBackgroundTrafficFlow
	Add() ChaosBackgroundTrafficFlow
	Append(items ...ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	Set(index int, newObj ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	Clear() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	clearHolderSlice() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
	appendHolderSlice(item ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter
}

func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) setMsg(msg *chaosBackgroundTraffic) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Flows {
		obj.appendHolderSlice(&chaosBackgroundTrafficFlow{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) Items() []ChaosBackgroundTrafficFlow {
	return obj.chaosBackgroundTrafficFlowSlice
}

func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) Add() ChaosBackgroundTrafficFlow {
	newObj := &onexdatamodel.ChaosBackgroundTrafficFlow{}
	obj.obj.obj.Flows = append(obj.obj.obj.Flows, newObj)
	newLibObj := &chaosBackgroundTrafficFlow{obj: newObj}
	newLibObj.setDefault()
	obj.chaosBackgroundTrafficFlowSlice = append(obj.chaosBackgroundTrafficFlowSlice, newLibObj)
	return newLibObj
}

func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) Append(items ...ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Flows = append(obj.obj.obj.Flows, newObj)
		obj.chaosBackgroundTrafficFlowSlice = append(obj.chaosBackgroundTrafficFlowSlice, item)
	}
	return obj
}

func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) Set(index int, newObj ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	obj.obj.obj.Flows[index] = newObj.Msg()
	obj.chaosBackgroundTrafficFlowSlice[index] = newObj
	return obj
}
func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) Clear() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	if len(obj.obj.obj.Flows) > 0 {
		obj.obj.obj.Flows = []*onexdatamodel.ChaosBackgroundTrafficFlow{}
		obj.chaosBackgroundTrafficFlowSlice = []ChaosBackgroundTrafficFlow{}
	}
	return obj
}
func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) clearHolderSlice() ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	if len(obj.chaosBackgroundTrafficFlowSlice) > 0 {
		obj.chaosBackgroundTrafficFlowSlice = []ChaosBackgroundTrafficFlow{}
	}
	return obj
}
func (obj *chaosBackgroundTrafficChaosBackgroundTrafficFlowIter) appendHolderSlice(item ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficChaosBackgroundTrafficFlowIter {
	obj.chaosBackgroundTrafficFlowSlice = append(obj.chaosBackgroundTrafficFlowSlice, item)
	return obj
}

func (obj *chaosBackgroundTraffic) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Flows != nil {

		if set_default {
			obj.Flows().clearHolderSlice()
			for _, item := range obj.obj.Flows {
				obj.Flows().appendHolderSlice(&chaosBackgroundTrafficFlow{obj: item})
			}
		}
		for _, item := range obj.Flows().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *chaosBackgroundTraffic) setDefault() {

}

type metricsResponseFlowResult struct {
	obj           *onexdatamodel.MetricsResponseFlowResult
	tcpInfoHolder MetricsResponseFlowResultTcpInfo
}

func NewMetricsResponseFlowResult() MetricsResponseFlowResult {
	obj := metricsResponseFlowResult{obj: &onexdatamodel.MetricsResponseFlowResult{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponseFlowResult) Msg() *onexdatamodel.MetricsResponseFlowResult {
	return obj.obj
}

func (obj *metricsResponseFlowResult) SetMsg(msg *onexdatamodel.MetricsResponseFlowResult) MetricsResponseFlowResult {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponseFlowResult) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *metricsResponseFlowResult) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *metricsResponseFlowResult) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsResponseFlowResult) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsResponseFlowResult) setNil() {
	obj.tcpInfoHolder = nil
}

// MetricsResponseFlowResult is result for a single data flow
type MetricsResponseFlowResult interface {
	Msg() *onexdatamodel.MetricsResponseFlowResult
	SetMsg(*onexdatamodel.MetricsResponseFlowResult) MetricsResponseFlowResult
	// ToPbText marshals MetricsResponseFlowResult to protobuf text
	ToPbText() string
	// ToYaml marshals MetricsResponseFlowResult to YAML text
	ToYaml() string
	// ToJson marshals MetricsResponseFlowResult to JSON text
	ToJson() string
	// FromPbText unmarshals MetricsResponseFlowResult from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponseFlowResult from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponseFlowResult from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponseFlowResult
	Validate() error
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
// flow completion time
func (obj *metricsResponseFlowResult) Fct() float32 {

	return *obj.obj.Fct

}

// Fct returns a float32
// flow completion time
func (obj *metricsResponseFlowResult) HasFct() bool {
	return obj.obj.Fct != nil
}

// SetFct sets the float32 value in the MetricsResponseFlowResult object
// flow completion time
func (obj *metricsResponseFlowResult) SetFct(value float32) MetricsResponseFlowResult {

	obj.obj.Fct = &value
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

	obj.TcpInfo().SetMsg(value.Msg())
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

type fabricSpine struct {
	obj *onexdatamodel.FabricSpine
}

func NewFabricSpine() FabricSpine {
	obj := fabricSpine{obj: &onexdatamodel.FabricSpine{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricSpine) Msg() *onexdatamodel.FabricSpine {
	return obj.obj
}

func (obj *fabricSpine) SetMsg(msg *onexdatamodel.FabricSpine) FabricSpine {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricSpine) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricSpine) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricSpine) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricSpine) FromYaml(value string) error {
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

func (obj *fabricSpine) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricSpine) FromJson(value string) error {
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

func (obj *fabricSpine) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricSpine) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricSpine is description is TBD
type FabricSpine interface {
	Msg() *onexdatamodel.FabricSpine
	SetMsg(*onexdatamodel.FabricSpine) FabricSpine
	// ToPbText marshals FabricSpine to protobuf text
	ToPbText() string
	// ToYaml marshals FabricSpine to YAML text
	ToYaml() string
	// ToJson marshals FabricSpine to JSON text
	ToJson() string
	// FromPbText unmarshals FabricSpine from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricSpine from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricSpine from JSON text
	FromJson(value string) error
	// Validate validates FabricSpine
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Count returns int32, set in FabricSpine.
	Count() int32
	// SetCount assigns int32 provided by user to FabricSpine
	SetCount(value int32) FabricSpine
	// HasCount checks if Count has been set in FabricSpine
	HasCount() bool
	// DownlinkEcmpMode returns FabricSpineDownlinkEcmpModeEnum, set in FabricSpine
	DownlinkEcmpMode() FabricSpineDownlinkEcmpModeEnum
	// SetDownlinkEcmpMode assigns FabricSpineDownlinkEcmpModeEnum provided by user to FabricSpine
	SetDownlinkEcmpMode(value FabricSpineDownlinkEcmpModeEnum) FabricSpine
	// HasDownlinkEcmpMode checks if DownlinkEcmpMode has been set in FabricSpine
	HasDownlinkEcmpMode() bool
	// QosProfileName returns string, set in FabricSpine.
	QosProfileName() string
	// SetQosProfileName assigns string provided by user to FabricSpine
	SetQosProfileName(value string) FabricSpine
	// HasQosProfileName checks if QosProfileName has been set in FabricSpine
	HasQosProfileName() bool
}

// Count returns a int32
// The number of spines to be created with each spine sharing the same
// downlink_ecmp_mode and qos_profile_name properties.
func (obj *fabricSpine) Count() int32 {

	return *obj.obj.Count

}

// Count returns a int32
// The number of spines to be created with each spine sharing the same
// downlink_ecmp_mode and qos_profile_name properties.
func (obj *fabricSpine) HasCount() bool {
	return obj.obj.Count != nil
}

// SetCount sets the int32 value in the FabricSpine object
// The number of spines to be created with each spine sharing the same
// downlink_ecmp_mode and qos_profile_name properties.
func (obj *fabricSpine) SetCount(value int32) FabricSpine {

	obj.obj.Count = &value
	return obj
}

type FabricSpineDownlinkEcmpModeEnum string

//  Enum of DownlinkEcmpMode on FabricSpine
var FabricSpineDownlinkEcmpMode = struct {
	RANDOM_SPRAY FabricSpineDownlinkEcmpModeEnum
	HASH_3_TUPLE FabricSpineDownlinkEcmpModeEnum
	HASH_5_TUPLE FabricSpineDownlinkEcmpModeEnum
}{
	RANDOM_SPRAY: FabricSpineDownlinkEcmpModeEnum("random_spray"),
	HASH_3_TUPLE: FabricSpineDownlinkEcmpModeEnum("hash_3_tuple"),
	HASH_5_TUPLE: FabricSpineDownlinkEcmpModeEnum("hash_5_tuple"),
}

func (obj *fabricSpine) DownlinkEcmpMode() FabricSpineDownlinkEcmpModeEnum {
	return FabricSpineDownlinkEcmpModeEnum(obj.obj.DownlinkEcmpMode.Enum().String())
}

// DownlinkEcmpMode returns a string
// The algorithm for packet distribution over ECMP links.
// - random_spray randomly puts each packet on an ECMP member links
// - hash_3_tuple is a 3 tuple hash of ipv4 src, dst, protocol
// - hash_5_tuple is static_hash_ipv4_l4 but a different resulting RTAG7 hash mode
func (obj *fabricSpine) HasDownlinkEcmpMode() bool {
	return obj.obj.DownlinkEcmpMode != nil
}

func (obj *fabricSpine) SetDownlinkEcmpMode(value FabricSpineDownlinkEcmpModeEnum) FabricSpine {
	intValue, ok := onexdatamodel.FabricSpine_DownlinkEcmpMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricSpineDownlinkEcmpModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricSpine_DownlinkEcmpMode_Enum(intValue)
	obj.obj.DownlinkEcmpMode = &enumValue

	return obj
}

// QosProfileName returns a string
// The name of a qos profile shared by the spines.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricSpine) QosProfileName() string {

	return *obj.obj.QosProfileName

}

// QosProfileName returns a string
// The name of a qos profile shared by the spines.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricSpine) HasQosProfileName() bool {
	return obj.obj.QosProfileName != nil
}

// SetQosProfileName sets the string value in the FabricSpine object
// The name of a qos profile shared by the spines.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricSpine) SetQosProfileName(value string) FabricSpine {

	obj.obj.QosProfileName = &value
	return obj
}

func (obj *fabricSpine) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricSpine) setDefault() {
	if obj.obj.Count == nil {
		obj.SetCount(1)
	}

}

type fabricPod struct {
	obj *onexdatamodel.FabricPod
}

func NewFabricPod() FabricPod {
	obj := fabricPod{obj: &onexdatamodel.FabricPod{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricPod) Msg() *onexdatamodel.FabricPod {
	return obj.obj
}

func (obj *fabricPod) SetMsg(msg *onexdatamodel.FabricPod) FabricPod {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricPod) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricPod) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricPod) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPod) FromYaml(value string) error {
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

func (obj *fabricPod) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPod) FromJson(value string) error {
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

func (obj *fabricPod) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricPod) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricPod is description is TBD
type FabricPod interface {
	Msg() *onexdatamodel.FabricPod
	SetMsg(*onexdatamodel.FabricPod) FabricPod
	// ToPbText marshals FabricPod to protobuf text
	ToPbText() string
	// ToYaml marshals FabricPod to YAML text
	ToYaml() string
	// ToJson marshals FabricPod to JSON text
	ToJson() string
	// FromPbText unmarshals FabricPod from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricPod from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricPod from JSON text
	FromJson(value string) error
	// Validate validates FabricPod
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Count returns int32, set in FabricPod.
	Count() int32
	// SetCount assigns int32 provided by user to FabricPod
	SetCount(value int32) FabricPod
	// HasCount checks if Count has been set in FabricPod
	HasCount() bool
	// PodProfileName returns []string, set in FabricPod.
	PodProfileName() []string
	// SetPodProfileName assigns []string provided by user to FabricPod
	SetPodProfileName(value []string) FabricPod
}

// Count returns a int32
// The number of pods that will share the same profile
func (obj *fabricPod) Count() int32 {

	return *obj.obj.Count

}

// Count returns a int32
// The number of pods that will share the same profile
func (obj *fabricPod) HasCount() bool {
	return obj.obj.Count != nil
}

// SetCount sets the int32 value in the FabricPod object
// The number of pods that will share the same profile
func (obj *fabricPod) SetCount(value int32) FabricPod {

	obj.obj.Count = &value
	return obj
}

// PodProfileName returns a []string
// The pod profile associated with this pod.
func (obj *fabricPod) PodProfileName() []string {
	if obj.obj.PodProfileName == nil {
		obj.obj.PodProfileName = make([]string, 0)
	}
	return obj.obj.PodProfileName
}

// SetPodProfileName sets the []string value in the FabricPod object
// The pod profile associated with this pod.
func (obj *fabricPod) SetPodProfileName(value []string) FabricPod {

	if obj.obj.PodProfileName == nil {
		obj.obj.PodProfileName = make([]string, 0)
	}
	obj.obj.PodProfileName = value

	return obj
}

func (obj *fabricPod) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricPod) setDefault() {
	if obj.obj.Count == nil {
		obj.SetCount(1)
	}

}

type switchHostLink struct {
	obj        *onexdatamodel.SwitchHostLink
	podHolder  SwitchHostLinkSwitchRef
	rackHolder SwitchHostLinkSwitchRef
}

func NewSwitchHostLink() SwitchHostLink {
	obj := switchHostLink{obj: &onexdatamodel.SwitchHostLink{}}
	obj.setDefault()
	return &obj
}

func (obj *switchHostLink) Msg() *onexdatamodel.SwitchHostLink {
	return obj.obj
}

func (obj *switchHostLink) SetMsg(msg *onexdatamodel.SwitchHostLink) SwitchHostLink {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *switchHostLink) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *switchHostLink) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *switchHostLink) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *switchHostLink) FromYaml(value string) error {
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

func (obj *switchHostLink) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *switchHostLink) FromJson(value string) error {
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

func (obj *switchHostLink) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *switchHostLink) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *switchHostLink) setNil() {
	obj.podHolder = nil
	obj.rackHolder = nil
}

// SwitchHostLink is the ingress point of a host which is the index of a spine switch,
// a pod/rack switch.
type SwitchHostLink interface {
	Msg() *onexdatamodel.SwitchHostLink
	SetMsg(*onexdatamodel.SwitchHostLink) SwitchHostLink
	// ToPbText marshals SwitchHostLink to protobuf text
	ToPbText() string
	// ToYaml marshals SwitchHostLink to YAML text
	ToYaml() string
	// ToJson marshals SwitchHostLink to JSON text
	ToJson() string
	// FromPbText unmarshals SwitchHostLink from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals SwitchHostLink from YAML text
	FromYaml(value string) error
	// FromJson unmarshals SwitchHostLink from JSON text
	FromJson(value string) error
	// Validate validates SwitchHostLink
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// HostName returns string, set in SwitchHostLink.
	HostName() string
	// SetHostName assigns string provided by user to SwitchHostLink
	SetHostName(value string) SwitchHostLink
	// FrontPanelPort returns int32, set in SwitchHostLink.
	FrontPanelPort() int32
	// SetFrontPanelPort assigns int32 provided by user to SwitchHostLink
	SetFrontPanelPort(value int32) SwitchHostLink
	// HasFrontPanelPort checks if FrontPanelPort has been set in SwitchHostLink
	HasFrontPanelPort() bool
	// Choice returns SwitchHostLinkChoiceEnum, set in SwitchHostLink
	Choice() SwitchHostLinkChoiceEnum
	// SetChoice assigns SwitchHostLinkChoiceEnum provided by user to SwitchHostLink
	SetChoice(value SwitchHostLinkChoiceEnum) SwitchHostLink
	// HasChoice checks if Choice has been set in SwitchHostLink
	HasChoice() bool
	// Spine returns int32, set in SwitchHostLink.
	Spine() int32
	// SetSpine assigns int32 provided by user to SwitchHostLink
	SetSpine(value int32) SwitchHostLink
	// HasSpine checks if Spine has been set in SwitchHostLink
	HasSpine() bool
	// Pod returns SwitchHostLinkSwitchRef, set in SwitchHostLink.
	// SwitchHostLinkSwitchRef is location of the switch based on pod and switch index
	Pod() SwitchHostLinkSwitchRef
	// SetPod assigns SwitchHostLinkSwitchRef provided by user to SwitchHostLink.
	// SwitchHostLinkSwitchRef is location of the switch based on pod and switch index
	SetPod(value SwitchHostLinkSwitchRef) SwitchHostLink
	// HasPod checks if Pod has been set in SwitchHostLink
	HasPod() bool
	// Rack returns SwitchHostLinkSwitchRef, set in SwitchHostLink.
	// SwitchHostLinkSwitchRef is location of the switch based on pod and switch index
	Rack() SwitchHostLinkSwitchRef
	// SetRack assigns SwitchHostLinkSwitchRef provided by user to SwitchHostLink.
	// SwitchHostLinkSwitchRef is location of the switch based on pod and switch index
	SetRack(value SwitchHostLinkSwitchRef) SwitchHostLink
	// HasRack checks if Rack has been set in SwitchHostLink
	HasRack() bool
	setNil()
}

// HostName returns a string
// TBD
//
// x-constraint:
// - #components/schemas/Host/properties/name
//
func (obj *switchHostLink) HostName() string {

	return obj.obj.HostName
}

// SetHostName sets the string value in the SwitchHostLink object
// TBD
//
// x-constraint:
// - #components/schemas/Host/properties/name
//
func (obj *switchHostLink) SetHostName(value string) SwitchHostLink {

	obj.obj.HostName = value
	return obj
}

// FrontPanelPort returns a int32
// Optional front panel port number, if fabric is rendered on physical box
func (obj *switchHostLink) FrontPanelPort() int32 {

	return *obj.obj.FrontPanelPort

}

// FrontPanelPort returns a int32
// Optional front panel port number, if fabric is rendered on physical box
func (obj *switchHostLink) HasFrontPanelPort() bool {
	return obj.obj.FrontPanelPort != nil
}

// SetFrontPanelPort sets the int32 value in the SwitchHostLink object
// Optional front panel port number, if fabric is rendered on physical box
func (obj *switchHostLink) SetFrontPanelPort(value int32) SwitchHostLink {

	obj.obj.FrontPanelPort = &value
	return obj
}

type SwitchHostLinkChoiceEnum string

//  Enum of Choice on SwitchHostLink
var SwitchHostLinkChoice = struct {
	SPINE SwitchHostLinkChoiceEnum
	POD   SwitchHostLinkChoiceEnum
	RACK  SwitchHostLinkChoiceEnum
}{
	SPINE: SwitchHostLinkChoiceEnum("spine"),
	POD:   SwitchHostLinkChoiceEnum("pod"),
	RACK:  SwitchHostLinkChoiceEnum("rack"),
}

func (obj *switchHostLink) Choice() SwitchHostLinkChoiceEnum {
	return SwitchHostLinkChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *switchHostLink) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *switchHostLink) SetChoice(value SwitchHostLinkChoiceEnum) SwitchHostLink {
	intValue, ok := onexdatamodel.SwitchHostLink_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on SwitchHostLinkChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.SwitchHostLink_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Rack = nil
	obj.rackHolder = nil
	obj.obj.Pod = nil
	obj.podHolder = nil
	obj.obj.Spine = nil

	if value == SwitchHostLinkChoice.POD {
		obj.obj.Pod = NewSwitchHostLinkSwitchRef().Msg()
	}

	if value == SwitchHostLinkChoice.RACK {
		obj.obj.Rack = NewSwitchHostLinkSwitchRef().Msg()
	}

	return obj
}

// Spine returns a int32
// One based index of the spine switch based on the number of spines
// configured in the spine_pod_rack topology.
func (obj *switchHostLink) Spine() int32 {

	if obj.obj.Spine == nil {
		obj.SetChoice(SwitchHostLinkChoice.SPINE)
	}

	return *obj.obj.Spine

}

// Spine returns a int32
// One based index of the spine switch based on the number of spines
// configured in the spine_pod_rack topology.
func (obj *switchHostLink) HasSpine() bool {
	return obj.obj.Spine != nil
}

// SetSpine sets the int32 value in the SwitchHostLink object
// One based index of the spine switch based on the number of spines
// configured in the spine_pod_rack topology.
func (obj *switchHostLink) SetSpine(value int32) SwitchHostLink {
	obj.SetChoice(SwitchHostLinkChoice.SPINE)
	obj.obj.Spine = &value
	return obj
}

// Pod returns a SwitchHostLinkSwitchRef
// description is TBD
func (obj *switchHostLink) Pod() SwitchHostLinkSwitchRef {
	if obj.obj.Pod == nil {
		obj.SetChoice(SwitchHostLinkChoice.POD)
	}
	if obj.podHolder == nil {
		obj.podHolder = &switchHostLinkSwitchRef{obj: obj.obj.Pod}
	}
	return obj.podHolder
}

// Pod returns a SwitchHostLinkSwitchRef
// description is TBD
func (obj *switchHostLink) HasPod() bool {
	return obj.obj.Pod != nil
}

// SetPod sets the SwitchHostLinkSwitchRef value in the SwitchHostLink object
// description is TBD
func (obj *switchHostLink) SetPod(value SwitchHostLinkSwitchRef) SwitchHostLink {
	obj.SetChoice(SwitchHostLinkChoice.POD)
	obj.Pod().SetMsg(value.Msg())
	return obj
}

// Rack returns a SwitchHostLinkSwitchRef
// description is TBD
func (obj *switchHostLink) Rack() SwitchHostLinkSwitchRef {
	if obj.obj.Rack == nil {
		obj.SetChoice(SwitchHostLinkChoice.RACK)
	}
	if obj.rackHolder == nil {
		obj.rackHolder = &switchHostLinkSwitchRef{obj: obj.obj.Rack}
	}
	return obj.rackHolder
}

// Rack returns a SwitchHostLinkSwitchRef
// description is TBD
func (obj *switchHostLink) HasRack() bool {
	return obj.obj.Rack != nil
}

// SetRack sets the SwitchHostLinkSwitchRef value in the SwitchHostLink object
// description is TBD
func (obj *switchHostLink) SetRack(value SwitchHostLinkSwitchRef) SwitchHostLink {
	obj.SetChoice(SwitchHostLinkChoice.RACK)
	obj.Rack().SetMsg(value.Msg())
	return obj
}

func (obj *switchHostLink) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	// HostName is required
	if obj.obj.HostName == "" {
		validation = append(validation, "HostName is required field on interface SwitchHostLink")
	}

	if obj.obj.Spine != nil {
		if *obj.obj.Spine < 1 || *obj.obj.Spine > 2147483647 {
			validation = append(
				validation,
				fmt.Sprintf("1 <= SwitchHostLink.Spine <= 2147483647 but Got %d", *obj.obj.Spine))
		}

	}

	if obj.obj.Pod != nil {
		obj.Pod().validateObj(set_default)
	}

	if obj.obj.Rack != nil {
		obj.Rack().validateObj(set_default)
	}

}

func (obj *switchHostLink) setDefault() {
	if obj.obj.Choice == nil {
		obj.SetChoice(SwitchHostLinkChoice.RACK)

	}

}

type fabricPodProfile struct {
	obj             *onexdatamodel.FabricPodProfile
	podSwitchHolder FabricPodSwitch
	rackHolder      FabricRack
}

func NewFabricPodProfile() FabricPodProfile {
	obj := fabricPodProfile{obj: &onexdatamodel.FabricPodProfile{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricPodProfile) Msg() *onexdatamodel.FabricPodProfile {
	return obj.obj
}

func (obj *fabricPodProfile) SetMsg(msg *onexdatamodel.FabricPodProfile) FabricPodProfile {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricPodProfile) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricPodProfile) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *fabricPodProfile) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPodProfile) FromYaml(value string) error {
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

func (obj *fabricPodProfile) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPodProfile) FromJson(value string) error {
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

func (obj *fabricPodProfile) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricPodProfile) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *fabricPodProfile) setNil() {
	obj.podSwitchHolder = nil
	obj.rackHolder = nil
}

// FabricPodProfile is description is TBD
type FabricPodProfile interface {
	Msg() *onexdatamodel.FabricPodProfile
	SetMsg(*onexdatamodel.FabricPodProfile) FabricPodProfile
	// ToPbText marshals FabricPodProfile to protobuf text
	ToPbText() string
	// ToYaml marshals FabricPodProfile to YAML text
	ToYaml() string
	// ToJson marshals FabricPodProfile to JSON text
	ToJson() string
	// FromPbText unmarshals FabricPodProfile from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricPodProfile from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricPodProfile from JSON text
	FromJson(value string) error
	// Validate validates FabricPodProfile
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in FabricPodProfile.
	Name() string
	// SetName assigns string provided by user to FabricPodProfile
	SetName(value string) FabricPodProfile
	// HasName checks if Name has been set in FabricPodProfile
	HasName() bool
	// PodSwitch returns FabricPodSwitch, set in FabricPodProfile.
	// FabricPodSwitch is description is TBD
	PodSwitch() FabricPodSwitch
	// SetPodSwitch assigns FabricPodSwitch provided by user to FabricPodProfile.
	// FabricPodSwitch is description is TBD
	SetPodSwitch(value FabricPodSwitch) FabricPodProfile
	// HasPodSwitch checks if PodSwitch has been set in FabricPodProfile
	HasPodSwitch() bool
	// Rack returns FabricRack, set in FabricPodProfile.
	// FabricRack is description is TBD
	Rack() FabricRack
	// SetRack assigns FabricRack provided by user to FabricPodProfile.
	// FabricRack is description is TBD
	SetRack(value FabricRack) FabricPodProfile
	// HasRack checks if Rack has been set in FabricPodProfile
	HasRack() bool
	setNil()
}

// Name returns a string
// description is TBD
func (obj *fabricPodProfile) Name() string {

	return *obj.obj.Name

}

// Name returns a string
// description is TBD
func (obj *fabricPodProfile) HasName() bool {
	return obj.obj.Name != nil
}

// SetName sets the string value in the FabricPodProfile object
// description is TBD
func (obj *fabricPodProfile) SetName(value string) FabricPodProfile {

	obj.obj.Name = &value
	return obj
}

// PodSwitch returns a FabricPodSwitch
// description is TBD
func (obj *fabricPodProfile) PodSwitch() FabricPodSwitch {
	if obj.obj.PodSwitch == nil {
		obj.obj.PodSwitch = NewFabricPodSwitch().Msg()
	}
	if obj.podSwitchHolder == nil {
		obj.podSwitchHolder = &fabricPodSwitch{obj: obj.obj.PodSwitch}
	}
	return obj.podSwitchHolder
}

// PodSwitch returns a FabricPodSwitch
// description is TBD
func (obj *fabricPodProfile) HasPodSwitch() bool {
	return obj.obj.PodSwitch != nil
}

// SetPodSwitch sets the FabricPodSwitch value in the FabricPodProfile object
// description is TBD
func (obj *fabricPodProfile) SetPodSwitch(value FabricPodSwitch) FabricPodProfile {

	obj.PodSwitch().SetMsg(value.Msg())
	return obj
}

// Rack returns a FabricRack
// description is TBD
func (obj *fabricPodProfile) Rack() FabricRack {
	if obj.obj.Rack == nil {
		obj.obj.Rack = NewFabricRack().Msg()
	}
	if obj.rackHolder == nil {
		obj.rackHolder = &fabricRack{obj: obj.obj.Rack}
	}
	return obj.rackHolder
}

// Rack returns a FabricRack
// description is TBD
func (obj *fabricPodProfile) HasRack() bool {
	return obj.obj.Rack != nil
}

// SetRack sets the FabricRack value in the FabricPodProfile object
// description is TBD
func (obj *fabricPodProfile) SetRack(value FabricRack) FabricPodProfile {

	obj.Rack().SetMsg(value.Msg())
	return obj
}

func (obj *fabricPodProfile) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.PodSwitch != nil {
		obj.PodSwitch().validateObj(set_default)
	}

	if obj.obj.Rack != nil {
		obj.Rack().validateObj(set_default)
	}

}

func (obj *fabricPodProfile) setDefault() {

}

type fabricRackProfile struct {
	obj *onexdatamodel.FabricRackProfile
}

func NewFabricRackProfile() FabricRackProfile {
	obj := fabricRackProfile{obj: &onexdatamodel.FabricRackProfile{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricRackProfile) Msg() *onexdatamodel.FabricRackProfile {
	return obj.obj
}

func (obj *fabricRackProfile) SetMsg(msg *onexdatamodel.FabricRackProfile) FabricRackProfile {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricRackProfile) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricRackProfile) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricRackProfile) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricRackProfile) FromYaml(value string) error {
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

func (obj *fabricRackProfile) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricRackProfile) FromJson(value string) error {
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

func (obj *fabricRackProfile) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricRackProfile) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricRackProfile is description is TBD
type FabricRackProfile interface {
	Msg() *onexdatamodel.FabricRackProfile
	SetMsg(*onexdatamodel.FabricRackProfile) FabricRackProfile
	// ToPbText marshals FabricRackProfile to protobuf text
	ToPbText() string
	// ToYaml marshals FabricRackProfile to YAML text
	ToYaml() string
	// ToJson marshals FabricRackProfile to JSON text
	ToJson() string
	// FromPbText unmarshals FabricRackProfile from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricRackProfile from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricRackProfile from JSON text
	FromJson(value string) error
	// Validate validates FabricRackProfile
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in FabricRackProfile.
	Name() string
	// SetName assigns string provided by user to FabricRackProfile
	SetName(value string) FabricRackProfile
	// HasName checks if Name has been set in FabricRackProfile
	HasName() bool
	// TorUplinkEcmpMode returns FabricRackProfileTorUplinkEcmpModeEnum, set in FabricRackProfile
	TorUplinkEcmpMode() FabricRackProfileTorUplinkEcmpModeEnum
	// SetTorUplinkEcmpMode assigns FabricRackProfileTorUplinkEcmpModeEnum provided by user to FabricRackProfile
	SetTorUplinkEcmpMode(value FabricRackProfileTorUplinkEcmpModeEnum) FabricRackProfile
	// HasTorUplinkEcmpMode checks if TorUplinkEcmpMode has been set in FabricRackProfile
	HasTorUplinkEcmpMode() bool
	// TorDownlinkEcmpMode returns FabricRackProfileTorDownlinkEcmpModeEnum, set in FabricRackProfile
	TorDownlinkEcmpMode() FabricRackProfileTorDownlinkEcmpModeEnum
	// SetTorDownlinkEcmpMode assigns FabricRackProfileTorDownlinkEcmpModeEnum provided by user to FabricRackProfile
	SetTorDownlinkEcmpMode(value FabricRackProfileTorDownlinkEcmpModeEnum) FabricRackProfile
	// HasTorDownlinkEcmpMode checks if TorDownlinkEcmpMode has been set in FabricRackProfile
	HasTorDownlinkEcmpMode() bool
	// TorQosProfileName returns string, set in FabricRackProfile.
	TorQosProfileName() string
	// SetTorQosProfileName assigns string provided by user to FabricRackProfile
	SetTorQosProfileName(value string) FabricRackProfile
	// HasTorQosProfileName checks if TorQosProfileName has been set in FabricRackProfile
	HasTorQosProfileName() bool
	// TorToPodOversubscription returns string, set in FabricRackProfile.
	TorToPodOversubscription() string
	// SetTorToPodOversubscription assigns string provided by user to FabricRackProfile
	SetTorToPodOversubscription(value string) FabricRackProfile
	// HasTorToPodOversubscription checks if TorToPodOversubscription has been set in FabricRackProfile
	HasTorToPodOversubscription() bool
}

// Name returns a string
// description is TBD
func (obj *fabricRackProfile) Name() string {

	return *obj.obj.Name

}

// Name returns a string
// description is TBD
func (obj *fabricRackProfile) HasName() bool {
	return obj.obj.Name != nil
}

// SetName sets the string value in the FabricRackProfile object
// description is TBD
func (obj *fabricRackProfile) SetName(value string) FabricRackProfile {

	obj.obj.Name = &value
	return obj
}

type FabricRackProfileTorUplinkEcmpModeEnum string

//  Enum of TorUplinkEcmpMode on FabricRackProfile
var FabricRackProfileTorUplinkEcmpMode = struct {
	RANDOM_SPRAY FabricRackProfileTorUplinkEcmpModeEnum
	HASH_3_TUPLE FabricRackProfileTorUplinkEcmpModeEnum
	HASH_5_TUPLE FabricRackProfileTorUplinkEcmpModeEnum
}{
	RANDOM_SPRAY: FabricRackProfileTorUplinkEcmpModeEnum("random_spray"),
	HASH_3_TUPLE: FabricRackProfileTorUplinkEcmpModeEnum("hash_3_tuple"),
	HASH_5_TUPLE: FabricRackProfileTorUplinkEcmpModeEnum("hash_5_tuple"),
}

func (obj *fabricRackProfile) TorUplinkEcmpMode() FabricRackProfileTorUplinkEcmpModeEnum {
	return FabricRackProfileTorUplinkEcmpModeEnum(obj.obj.TorUplinkEcmpMode.Enum().String())
}

// TorUplinkEcmpMode returns a string
// The algorithm for packet distribution over ECMP links.
// - random_spray randomly puts each packet on an ECMP member links
// - hash_3_tuple is a 3 tuple hash of ipv4 src, dst, protocol
// - hash_5_tuple is static_hash_ipv4_l4 but a different resulting RTAG7 hash mode
func (obj *fabricRackProfile) HasTorUplinkEcmpMode() bool {
	return obj.obj.TorUplinkEcmpMode != nil
}

func (obj *fabricRackProfile) SetTorUplinkEcmpMode(value FabricRackProfileTorUplinkEcmpModeEnum) FabricRackProfile {
	intValue, ok := onexdatamodel.FabricRackProfile_TorUplinkEcmpMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricRackProfileTorUplinkEcmpModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricRackProfile_TorUplinkEcmpMode_Enum(intValue)
	obj.obj.TorUplinkEcmpMode = &enumValue

	return obj
}

type FabricRackProfileTorDownlinkEcmpModeEnum string

//  Enum of TorDownlinkEcmpMode on FabricRackProfile
var FabricRackProfileTorDownlinkEcmpMode = struct {
	RANDOM_SPRAY FabricRackProfileTorDownlinkEcmpModeEnum
	HASH_3_TUPLE FabricRackProfileTorDownlinkEcmpModeEnum
	HASH_5_TUPLE FabricRackProfileTorDownlinkEcmpModeEnum
}{
	RANDOM_SPRAY: FabricRackProfileTorDownlinkEcmpModeEnum("random_spray"),
	HASH_3_TUPLE: FabricRackProfileTorDownlinkEcmpModeEnum("hash_3_tuple"),
	HASH_5_TUPLE: FabricRackProfileTorDownlinkEcmpModeEnum("hash_5_tuple"),
}

func (obj *fabricRackProfile) TorDownlinkEcmpMode() FabricRackProfileTorDownlinkEcmpModeEnum {
	return FabricRackProfileTorDownlinkEcmpModeEnum(obj.obj.TorDownlinkEcmpMode.Enum().String())
}

// TorDownlinkEcmpMode returns a string
// The algorithm for packet distribution over ECMP links.
// - random_spray randomly puts each packet on an ECMP member links
// - hash_3_tuple is a 3 tuple hash of ipv4 src, dst, protocol
// - hash_5_tuple is static_hash_ipv4_l4 but a different resulting RTAG7 hash mode
func (obj *fabricRackProfile) HasTorDownlinkEcmpMode() bool {
	return obj.obj.TorDownlinkEcmpMode != nil
}

func (obj *fabricRackProfile) SetTorDownlinkEcmpMode(value FabricRackProfileTorDownlinkEcmpModeEnum) FabricRackProfile {
	intValue, ok := onexdatamodel.FabricRackProfile_TorDownlinkEcmpMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricRackProfileTorDownlinkEcmpModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricRackProfile_TorDownlinkEcmpMode_Enum(intValue)
	obj.obj.TorDownlinkEcmpMode = &enumValue

	return obj
}

// TorQosProfileName returns a string
// The name of a qos profile associated with the ToR of this rack
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricRackProfile) TorQosProfileName() string {

	return *obj.obj.TorQosProfileName

}

// TorQosProfileName returns a string
// The name of a qos profile associated with the ToR of this rack
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricRackProfile) HasTorQosProfileName() bool {
	return obj.obj.TorQosProfileName != nil
}

// SetTorQosProfileName sets the string value in the FabricRackProfile object
// The name of a qos profile associated with the ToR of this rack
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricRackProfile) SetTorQosProfileName(value string) FabricRackProfile {

	obj.obj.TorQosProfileName = &value
	return obj
}

// TorToPodOversubscription returns a string
// The oversubscription ratio of the ToR switch
func (obj *fabricRackProfile) TorToPodOversubscription() string {

	return *obj.obj.TorToPodOversubscription

}

// TorToPodOversubscription returns a string
// The oversubscription ratio of the ToR switch
func (obj *fabricRackProfile) HasTorToPodOversubscription() bool {
	return obj.obj.TorToPodOversubscription != nil
}

// SetTorToPodOversubscription sets the string value in the FabricRackProfile object
// The oversubscription ratio of the ToR switch
func (obj *fabricRackProfile) SetTorToPodOversubscription(value string) FabricRackProfile {

	obj.obj.TorToPodOversubscription = &value
	return obj
}

func (obj *fabricRackProfile) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricRackProfile) setDefault() {

}

type fabricQosProfileIngressAdmission struct {
	obj *onexdatamodel.FabricQosProfileIngressAdmission
}

func NewFabricQosProfileIngressAdmission() FabricQosProfileIngressAdmission {
	obj := fabricQosProfileIngressAdmission{obj: &onexdatamodel.FabricQosProfileIngressAdmission{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricQosProfileIngressAdmission) Msg() *onexdatamodel.FabricQosProfileIngressAdmission {
	return obj.obj
}

func (obj *fabricQosProfileIngressAdmission) SetMsg(msg *onexdatamodel.FabricQosProfileIngressAdmission) FabricQosProfileIngressAdmission {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricQosProfileIngressAdmission) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricQosProfileIngressAdmission) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricQosProfileIngressAdmission) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfileIngressAdmission) FromYaml(value string) error {
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

func (obj *fabricQosProfileIngressAdmission) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfileIngressAdmission) FromJson(value string) error {
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

func (obj *fabricQosProfileIngressAdmission) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricQosProfileIngressAdmission) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricQosProfileIngressAdmission is description is TBD
type FabricQosProfileIngressAdmission interface {
	Msg() *onexdatamodel.FabricQosProfileIngressAdmission
	SetMsg(*onexdatamodel.FabricQosProfileIngressAdmission) FabricQosProfileIngressAdmission
	// ToPbText marshals FabricQosProfileIngressAdmission to protobuf text
	ToPbText() string
	// ToYaml marshals FabricQosProfileIngressAdmission to YAML text
	ToYaml() string
	// ToJson marshals FabricQosProfileIngressAdmission to JSON text
	ToJson() string
	// FromPbText unmarshals FabricQosProfileIngressAdmission from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricQosProfileIngressAdmission from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricQosProfileIngressAdmission from JSON text
	FromJson(value string) error
	// Validate validates FabricQosProfileIngressAdmission
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// ReservedBufferBytes returns int32, set in FabricQosProfileIngressAdmission.
	ReservedBufferBytes() int32
	// SetReservedBufferBytes assigns int32 provided by user to FabricQosProfileIngressAdmission
	SetReservedBufferBytes(value int32) FabricQosProfileIngressAdmission
	// HasReservedBufferBytes checks if ReservedBufferBytes has been set in FabricQosProfileIngressAdmission
	HasReservedBufferBytes() bool
	// SharedBufferBytes returns int32, set in FabricQosProfileIngressAdmission.
	SharedBufferBytes() int32
	// SetSharedBufferBytes assigns int32 provided by user to FabricQosProfileIngressAdmission
	SetSharedBufferBytes(value int32) FabricQosProfileIngressAdmission
	// HasSharedBufferBytes checks if SharedBufferBytes has been set in FabricQosProfileIngressAdmission
	HasSharedBufferBytes() bool
}

// ReservedBufferBytes returns a int32
// Buffer space (in bytes) reserved for each port that this Qos profile applies to
func (obj *fabricQosProfileIngressAdmission) ReservedBufferBytes() int32 {

	return *obj.obj.ReservedBufferBytes

}

// ReservedBufferBytes returns a int32
// Buffer space (in bytes) reserved for each port that this Qos profile applies to
func (obj *fabricQosProfileIngressAdmission) HasReservedBufferBytes() bool {
	return obj.obj.ReservedBufferBytes != nil
}

// SetReservedBufferBytes sets the int32 value in the FabricQosProfileIngressAdmission object
// Buffer space (in bytes) reserved for each port that this Qos profile applies to
func (obj *fabricQosProfileIngressAdmission) SetReservedBufferBytes(value int32) FabricQosProfileIngressAdmission {

	obj.obj.ReservedBufferBytes = &value
	return obj
}

// SharedBufferBytes returns a int32
// Amount of shared buffer space (in bytes) available
func (obj *fabricQosProfileIngressAdmission) SharedBufferBytes() int32 {

	return *obj.obj.SharedBufferBytes

}

// SharedBufferBytes returns a int32
// Amount of shared buffer space (in bytes) available
func (obj *fabricQosProfileIngressAdmission) HasSharedBufferBytes() bool {
	return obj.obj.SharedBufferBytes != nil
}

// SetSharedBufferBytes sets the int32 value in the FabricQosProfileIngressAdmission object
// Amount of shared buffer space (in bytes) available
func (obj *fabricQosProfileIngressAdmission) SetSharedBufferBytes(value int32) FabricQosProfileIngressAdmission {

	obj.obj.SharedBufferBytes = &value
	return obj
}

func (obj *fabricQosProfileIngressAdmission) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricQosProfileIngressAdmission) setDefault() {
	if obj.obj.ReservedBufferBytes == nil {
		obj.SetReservedBufferBytes(0)
	}
	if obj.obj.SharedBufferBytes == nil {
		obj.SetSharedBufferBytes(0)
	}

}

type fabricQosProfileScheduler struct {
	obj *onexdatamodel.FabricQosProfileScheduler
}

func NewFabricQosProfileScheduler() FabricQosProfileScheduler {
	obj := fabricQosProfileScheduler{obj: &onexdatamodel.FabricQosProfileScheduler{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricQosProfileScheduler) Msg() *onexdatamodel.FabricQosProfileScheduler {
	return obj.obj
}

func (obj *fabricQosProfileScheduler) SetMsg(msg *onexdatamodel.FabricQosProfileScheduler) FabricQosProfileScheduler {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricQosProfileScheduler) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricQosProfileScheduler) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricQosProfileScheduler) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfileScheduler) FromYaml(value string) error {
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

func (obj *fabricQosProfileScheduler) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfileScheduler) FromJson(value string) error {
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

func (obj *fabricQosProfileScheduler) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricQosProfileScheduler) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricQosProfileScheduler is description is TBD
type FabricQosProfileScheduler interface {
	Msg() *onexdatamodel.FabricQosProfileScheduler
	SetMsg(*onexdatamodel.FabricQosProfileScheduler) FabricQosProfileScheduler
	// ToPbText marshals FabricQosProfileScheduler to protobuf text
	ToPbText() string
	// ToYaml marshals FabricQosProfileScheduler to YAML text
	ToYaml() string
	// ToJson marshals FabricQosProfileScheduler to JSON text
	ToJson() string
	// FromPbText unmarshals FabricQosProfileScheduler from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricQosProfileScheduler from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricQosProfileScheduler from JSON text
	FromJson(value string) error
	// Validate validates FabricQosProfileScheduler
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// SchedulerMode returns FabricQosProfileSchedulerSchedulerModeEnum, set in FabricQosProfileScheduler
	SchedulerMode() FabricQosProfileSchedulerSchedulerModeEnum
	// SetSchedulerMode assigns FabricQosProfileSchedulerSchedulerModeEnum provided by user to FabricQosProfileScheduler
	SetSchedulerMode(value FabricQosProfileSchedulerSchedulerModeEnum) FabricQosProfileScheduler
	// HasSchedulerMode checks if SchedulerMode has been set in FabricQosProfileScheduler
	HasSchedulerMode() bool
	// WeightList returns []int32, set in FabricQosProfileScheduler.
	WeightList() []int32
	// SetWeightList assigns []int32 provided by user to FabricQosProfileScheduler
	SetWeightList(value []int32) FabricQosProfileScheduler
}

type FabricQosProfileSchedulerSchedulerModeEnum string

//  Enum of SchedulerMode on FabricQosProfileScheduler
var FabricQosProfileSchedulerSchedulerMode = struct {
	STRICT_PRIORITY      FabricQosProfileSchedulerSchedulerModeEnum
	WEIGHTED_ROUND_ROBIN FabricQosProfileSchedulerSchedulerModeEnum
}{
	STRICT_PRIORITY:      FabricQosProfileSchedulerSchedulerModeEnum("strict_priority"),
	WEIGHTED_ROUND_ROBIN: FabricQosProfileSchedulerSchedulerModeEnum("weighted_round_robin"),
}

func (obj *fabricQosProfileScheduler) SchedulerMode() FabricQosProfileSchedulerSchedulerModeEnum {
	return FabricQosProfileSchedulerSchedulerModeEnum(obj.obj.SchedulerMode.Enum().String())
}

// SchedulerMode returns a string
// The queue scheduling discipline
func (obj *fabricQosProfileScheduler) HasSchedulerMode() bool {
	return obj.obj.SchedulerMode != nil
}

func (obj *fabricQosProfileScheduler) SetSchedulerMode(value FabricQosProfileSchedulerSchedulerModeEnum) FabricQosProfileScheduler {
	intValue, ok := onexdatamodel.FabricQosProfileScheduler_SchedulerMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricQosProfileSchedulerSchedulerModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricQosProfileScheduler_SchedulerMode_Enum(intValue)
	obj.obj.SchedulerMode = &enumValue

	return obj
}

// WeightList returns a []int32
// A list of queue weights
func (obj *fabricQosProfileScheduler) WeightList() []int32 {
	if obj.obj.WeightList == nil {
		obj.obj.WeightList = make([]int32, 0)
	}
	return obj.obj.WeightList
}

// SetWeightList sets the []int32 value in the FabricQosProfileScheduler object
// A list of queue weights
func (obj *fabricQosProfileScheduler) SetWeightList(value []int32) FabricQosProfileScheduler {

	if obj.obj.WeightList == nil {
		obj.obj.WeightList = make([]int32, 0)
	}
	obj.obj.WeightList = value

	return obj
}

func (obj *fabricQosProfileScheduler) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricQosProfileScheduler) setDefault() {

}

type fabricQosProfilePacketClassification struct {
	obj                          *onexdatamodel.FabricQosProfilePacketClassification
	mapDscpToTrafficClassHolder  FabricQosProfilePacketClassificationMap
	mapTrafficClassToQueueHolder FabricQosProfilePacketClassificationMap
}

func NewFabricQosProfilePacketClassification() FabricQosProfilePacketClassification {
	obj := fabricQosProfilePacketClassification{obj: &onexdatamodel.FabricQosProfilePacketClassification{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricQosProfilePacketClassification) Msg() *onexdatamodel.FabricQosProfilePacketClassification {
	return obj.obj
}

func (obj *fabricQosProfilePacketClassification) SetMsg(msg *onexdatamodel.FabricQosProfilePacketClassification) FabricQosProfilePacketClassification {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricQosProfilePacketClassification) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricQosProfilePacketClassification) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *fabricQosProfilePacketClassification) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfilePacketClassification) FromYaml(value string) error {
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

func (obj *fabricQosProfilePacketClassification) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfilePacketClassification) FromJson(value string) error {
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

func (obj *fabricQosProfilePacketClassification) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricQosProfilePacketClassification) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *fabricQosProfilePacketClassification) setNil() {
	obj.mapDscpToTrafficClassHolder = nil
	obj.mapTrafficClassToQueueHolder = nil
}

// FabricQosProfilePacketClassification is description is TBD
type FabricQosProfilePacketClassification interface {
	Msg() *onexdatamodel.FabricQosProfilePacketClassification
	SetMsg(*onexdatamodel.FabricQosProfilePacketClassification) FabricQosProfilePacketClassification
	// ToPbText marshals FabricQosProfilePacketClassification to protobuf text
	ToPbText() string
	// ToYaml marshals FabricQosProfilePacketClassification to YAML text
	ToYaml() string
	// ToJson marshals FabricQosProfilePacketClassification to JSON text
	ToJson() string
	// FromPbText unmarshals FabricQosProfilePacketClassification from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricQosProfilePacketClassification from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricQosProfilePacketClassification from JSON text
	FromJson(value string) error
	// Validate validates FabricQosProfilePacketClassification
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// MapDscpToTrafficClass returns FabricQosProfilePacketClassificationMap, set in FabricQosProfilePacketClassification.
	// FabricQosProfilePacketClassificationMap is description is TBD
	MapDscpToTrafficClass() FabricQosProfilePacketClassificationMap
	// SetMapDscpToTrafficClass assigns FabricQosProfilePacketClassificationMap provided by user to FabricQosProfilePacketClassification.
	// FabricQosProfilePacketClassificationMap is description is TBD
	SetMapDscpToTrafficClass(value FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassification
	// HasMapDscpToTrafficClass checks if MapDscpToTrafficClass has been set in FabricQosProfilePacketClassification
	HasMapDscpToTrafficClass() bool
	// MapTrafficClassToQueue returns FabricQosProfilePacketClassificationMap, set in FabricQosProfilePacketClassification.
	// FabricQosProfilePacketClassificationMap is description is TBD
	MapTrafficClassToQueue() FabricQosProfilePacketClassificationMap
	// SetMapTrafficClassToQueue assigns FabricQosProfilePacketClassificationMap provided by user to FabricQosProfilePacketClassification.
	// FabricQosProfilePacketClassificationMap is description is TBD
	SetMapTrafficClassToQueue(value FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassification
	// HasMapTrafficClassToQueue checks if MapTrafficClassToQueue has been set in FabricQosProfilePacketClassification
	HasMapTrafficClassToQueue() bool
	setNil()
}

// MapDscpToTrafficClass returns a FabricQosProfilePacketClassificationMap
// description is TBD
func (obj *fabricQosProfilePacketClassification) MapDscpToTrafficClass() FabricQosProfilePacketClassificationMap {
	if obj.obj.MapDscpToTrafficClass == nil {
		obj.obj.MapDscpToTrafficClass = NewFabricQosProfilePacketClassificationMap().Msg()
	}
	if obj.mapDscpToTrafficClassHolder == nil {
		obj.mapDscpToTrafficClassHolder = &fabricQosProfilePacketClassificationMap{obj: obj.obj.MapDscpToTrafficClass}
	}
	return obj.mapDscpToTrafficClassHolder
}

// MapDscpToTrafficClass returns a FabricQosProfilePacketClassificationMap
// description is TBD
func (obj *fabricQosProfilePacketClassification) HasMapDscpToTrafficClass() bool {
	return obj.obj.MapDscpToTrafficClass != nil
}

// SetMapDscpToTrafficClass sets the FabricQosProfilePacketClassificationMap value in the FabricQosProfilePacketClassification object
// description is TBD
func (obj *fabricQosProfilePacketClassification) SetMapDscpToTrafficClass(value FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassification {

	obj.MapDscpToTrafficClass().SetMsg(value.Msg())
	return obj
}

// MapTrafficClassToQueue returns a FabricQosProfilePacketClassificationMap
// description is TBD
func (obj *fabricQosProfilePacketClassification) MapTrafficClassToQueue() FabricQosProfilePacketClassificationMap {
	if obj.obj.MapTrafficClassToQueue == nil {
		obj.obj.MapTrafficClassToQueue = NewFabricQosProfilePacketClassificationMap().Msg()
	}
	if obj.mapTrafficClassToQueueHolder == nil {
		obj.mapTrafficClassToQueueHolder = &fabricQosProfilePacketClassificationMap{obj: obj.obj.MapTrafficClassToQueue}
	}
	return obj.mapTrafficClassToQueueHolder
}

// MapTrafficClassToQueue returns a FabricQosProfilePacketClassificationMap
// description is TBD
func (obj *fabricQosProfilePacketClassification) HasMapTrafficClassToQueue() bool {
	return obj.obj.MapTrafficClassToQueue != nil
}

// SetMapTrafficClassToQueue sets the FabricQosProfilePacketClassificationMap value in the FabricQosProfilePacketClassification object
// description is TBD
func (obj *fabricQosProfilePacketClassification) SetMapTrafficClassToQueue(value FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassification {

	obj.MapTrafficClassToQueue().SetMsg(value.Msg())
	return obj
}

func (obj *fabricQosProfilePacketClassification) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.MapDscpToTrafficClass != nil {
		obj.MapDscpToTrafficClass().validateObj(set_default)
	}

	if obj.obj.MapTrafficClassToQueue != nil {
		obj.MapTrafficClassToQueue().validateObj(set_default)
	}

}

func (obj *fabricQosProfilePacketClassification) setDefault() {

}

type dataflowScatterWorkload struct {
	obj *onexdatamodel.DataflowScatterWorkload
}

func NewDataflowScatterWorkload() DataflowScatterWorkload {
	obj := dataflowScatterWorkload{obj: &onexdatamodel.DataflowScatterWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowScatterWorkload) Msg() *onexdatamodel.DataflowScatterWorkload {
	return obj.obj
}

func (obj *dataflowScatterWorkload) SetMsg(msg *onexdatamodel.DataflowScatterWorkload) DataflowScatterWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowScatterWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowScatterWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowScatterWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowScatterWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowScatterWorkload is description is TBD
type DataflowScatterWorkload interface {
	Msg() *onexdatamodel.DataflowScatterWorkload
	SetMsg(*onexdatamodel.DataflowScatterWorkload) DataflowScatterWorkload
	// ToPbText marshals DataflowScatterWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowScatterWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowScatterWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowScatterWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowScatterWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowScatterWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowScatterWorkload
	Validate() error
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

type dataflowGatherWorkload struct {
	obj *onexdatamodel.DataflowGatherWorkload
}

func NewDataflowGatherWorkload() DataflowGatherWorkload {
	obj := dataflowGatherWorkload{obj: &onexdatamodel.DataflowGatherWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowGatherWorkload) Msg() *onexdatamodel.DataflowGatherWorkload {
	return obj.obj
}

func (obj *dataflowGatherWorkload) SetMsg(msg *onexdatamodel.DataflowGatherWorkload) DataflowGatherWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowGatherWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowGatherWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowGatherWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowGatherWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowGatherWorkload is description is TBD
type DataflowGatherWorkload interface {
	Msg() *onexdatamodel.DataflowGatherWorkload
	SetMsg(*onexdatamodel.DataflowGatherWorkload) DataflowGatherWorkload
	// ToPbText marshals DataflowGatherWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowGatherWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowGatherWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowGatherWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowGatherWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowGatherWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowGatherWorkload
	Validate() error
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

type dataflowLoopWorkload struct {
	obj            *onexdatamodel.DataflowLoopWorkload
	childrenHolder DataflowLoopWorkloadDataflowWorkloadItemIter
}

func NewDataflowLoopWorkload() DataflowLoopWorkload {
	obj := dataflowLoopWorkload{obj: &onexdatamodel.DataflowLoopWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowLoopWorkload) Msg() *onexdatamodel.DataflowLoopWorkload {
	return obj.obj
}

func (obj *dataflowLoopWorkload) SetMsg(msg *onexdatamodel.DataflowLoopWorkload) DataflowLoopWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowLoopWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowLoopWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflowLoopWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowLoopWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowLoopWorkload) setNil() {
	obj.childrenHolder = nil
}

// DataflowLoopWorkload is description is TBD
type DataflowLoopWorkload interface {
	Msg() *onexdatamodel.DataflowLoopWorkload
	SetMsg(*onexdatamodel.DataflowLoopWorkload) DataflowLoopWorkload
	// ToPbText marshals DataflowLoopWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowLoopWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowLoopWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowLoopWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowLoopWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowLoopWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowLoopWorkload
	Validate() error
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
		obj.obj.Children = []*onexdatamodel.DataflowWorkloadItem{}
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
	newObj := &onexdatamodel.DataflowWorkloadItem{}
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
		obj.obj.obj.Children = []*onexdatamodel.DataflowWorkloadItem{}
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

type dataflowComputeWorkload struct {
	obj             *onexdatamodel.DataflowComputeWorkload
	simulatedHolder DataflowSimulatedComputeWorkload
}

func NewDataflowComputeWorkload() DataflowComputeWorkload {
	obj := dataflowComputeWorkload{obj: &onexdatamodel.DataflowComputeWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowComputeWorkload) Msg() *onexdatamodel.DataflowComputeWorkload {
	return obj.obj
}

func (obj *dataflowComputeWorkload) SetMsg(msg *onexdatamodel.DataflowComputeWorkload) DataflowComputeWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowComputeWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowComputeWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflowComputeWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowComputeWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowComputeWorkload) setNil() {
	obj.simulatedHolder = nil
}

// DataflowComputeWorkload is description is TBD
type DataflowComputeWorkload interface {
	Msg() *onexdatamodel.DataflowComputeWorkload
	SetMsg(*onexdatamodel.DataflowComputeWorkload) DataflowComputeWorkload
	// ToPbText marshals DataflowComputeWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowComputeWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowComputeWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowComputeWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowComputeWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowComputeWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowComputeWorkload
	Validate() error
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
	intValue, ok := onexdatamodel.DataflowComputeWorkload_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowComputeWorkloadChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.DataflowComputeWorkload_Choice_Enum(intValue)
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
	obj.Simulated().SetMsg(value.Msg())
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

type dataflowAllReduceWorkload struct {
	obj *onexdatamodel.DataflowAllReduceWorkload
}

func NewDataflowAllReduceWorkload() DataflowAllReduceWorkload {
	obj := dataflowAllReduceWorkload{obj: &onexdatamodel.DataflowAllReduceWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowAllReduceWorkload) Msg() *onexdatamodel.DataflowAllReduceWorkload {
	return obj.obj
}

func (obj *dataflowAllReduceWorkload) SetMsg(msg *onexdatamodel.DataflowAllReduceWorkload) DataflowAllReduceWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowAllReduceWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowAllReduceWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowAllReduceWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowAllReduceWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowAllReduceWorkload is description is TBD
type DataflowAllReduceWorkload interface {
	Msg() *onexdatamodel.DataflowAllReduceWorkload
	SetMsg(*onexdatamodel.DataflowAllReduceWorkload) DataflowAllReduceWorkload
	// ToPbText marshals DataflowAllReduceWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowAllReduceWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowAllReduceWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowAllReduceWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowAllReduceWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowAllReduceWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowAllReduceWorkload
	Validate() error
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
	ALL_TO_ALL DataflowAllReduceWorkloadTypeEnum
	RING       DataflowAllReduceWorkloadTypeEnum
	BUTTERFLY  DataflowAllReduceWorkloadTypeEnum
	PIPELINE   DataflowAllReduceWorkloadTypeEnum
}{
	ALL_TO_ALL: DataflowAllReduceWorkloadTypeEnum("all_to_all"),
	RING:       DataflowAllReduceWorkloadTypeEnum("ring"),
	BUTTERFLY:  DataflowAllReduceWorkloadTypeEnum("butterfly"),
	PIPELINE:   DataflowAllReduceWorkloadTypeEnum("pipeline"),
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
	intValue, ok := onexdatamodel.DataflowAllReduceWorkload_Type_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowAllReduceWorkloadTypeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.DataflowAllReduceWorkload_Type_Enum(intValue)
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
		obj.SetType(DataflowAllReduceWorkloadType.ALL_TO_ALL)

	}

}

type dataflowBroadcastWorkload struct {
	obj *onexdatamodel.DataflowBroadcastWorkload
}

func NewDataflowBroadcastWorkload() DataflowBroadcastWorkload {
	obj := dataflowBroadcastWorkload{obj: &onexdatamodel.DataflowBroadcastWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowBroadcastWorkload) Msg() *onexdatamodel.DataflowBroadcastWorkload {
	return obj.obj
}

func (obj *dataflowBroadcastWorkload) SetMsg(msg *onexdatamodel.DataflowBroadcastWorkload) DataflowBroadcastWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowBroadcastWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowBroadcastWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowBroadcastWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowBroadcastWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowBroadcastWorkload is description is TBD
type DataflowBroadcastWorkload interface {
	Msg() *onexdatamodel.DataflowBroadcastWorkload
	SetMsg(*onexdatamodel.DataflowBroadcastWorkload) DataflowBroadcastWorkload
	// ToPbText marshals DataflowBroadcastWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowBroadcastWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowBroadcastWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowBroadcastWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowBroadcastWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowBroadcastWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowBroadcastWorkload
	Validate() error
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

type dataflowFlowProfileEthernet struct {
	obj *onexdatamodel.DataflowFlowProfileEthernet
}

func NewDataflowFlowProfileEthernet() DataflowFlowProfileEthernet {
	obj := dataflowFlowProfileEthernet{obj: &onexdatamodel.DataflowFlowProfileEthernet{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileEthernet) Msg() *onexdatamodel.DataflowFlowProfileEthernet {
	return obj.obj
}

func (obj *dataflowFlowProfileEthernet) SetMsg(msg *onexdatamodel.DataflowFlowProfileEthernet) DataflowFlowProfileEthernet {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileEthernet) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowFlowProfileEthernet) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileEthernet) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfileEthernet) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowFlowProfileEthernet is description is TBD
type DataflowFlowProfileEthernet interface {
	Msg() *onexdatamodel.DataflowFlowProfileEthernet
	SetMsg(*onexdatamodel.DataflowFlowProfileEthernet) DataflowFlowProfileEthernet
	// ToPbText marshals DataflowFlowProfileEthernet to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowFlowProfileEthernet to YAML text
	ToYaml() string
	// ToJson marshals DataflowFlowProfileEthernet to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowFlowProfileEthernet from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileEthernet from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileEthernet from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileEthernet
	Validate() error
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

type dataflowFlowProfileTcp struct {
	obj                   *onexdatamodel.DataflowFlowProfileTcp
	destinationPortHolder L4PortRange
	sourcePortHolder      L4PortRange
}

func NewDataflowFlowProfileTcp() DataflowFlowProfileTcp {
	obj := dataflowFlowProfileTcp{obj: &onexdatamodel.DataflowFlowProfileTcp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileTcp) Msg() *onexdatamodel.DataflowFlowProfileTcp {
	return obj.obj
}

func (obj *dataflowFlowProfileTcp) SetMsg(msg *onexdatamodel.DataflowFlowProfileTcp) DataflowFlowProfileTcp {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileTcp) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowFlowProfileTcp) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *dataflowFlowProfileTcp) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfileTcp) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfileTcp) setNil() {
	obj.destinationPortHolder = nil
	obj.sourcePortHolder = nil
}

// DataflowFlowProfileTcp is description is TBD
type DataflowFlowProfileTcp interface {
	Msg() *onexdatamodel.DataflowFlowProfileTcp
	SetMsg(*onexdatamodel.DataflowFlowProfileTcp) DataflowFlowProfileTcp
	// ToPbText marshals DataflowFlowProfileTcp to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowFlowProfileTcp to YAML text
	ToYaml() string
	// ToJson marshals DataflowFlowProfileTcp to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowFlowProfileTcp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileTcp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileTcp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileTcp
	Validate() error
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
	intValue, ok := onexdatamodel.DataflowFlowProfileTcp_CongestionAlgorithm_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on DataflowFlowProfileTcpCongestionAlgorithmEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.DataflowFlowProfileTcp_CongestionAlgorithm_Enum(intValue)
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

	obj.DestinationPort().SetMsg(value.Msg())
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

	obj.SourcePort().SetMsg(value.Msg())
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

type dataflowFlowProfileUdp struct {
	obj *onexdatamodel.DataflowFlowProfileUdp
}

func NewDataflowFlowProfileUdp() DataflowFlowProfileUdp {
	obj := dataflowFlowProfileUdp{obj: &onexdatamodel.DataflowFlowProfileUdp{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowFlowProfileUdp) Msg() *onexdatamodel.DataflowFlowProfileUdp {
	return obj.obj
}

func (obj *dataflowFlowProfileUdp) SetMsg(msg *onexdatamodel.DataflowFlowProfileUdp) DataflowFlowProfileUdp {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowFlowProfileUdp) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowFlowProfileUdp) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowFlowProfileUdp) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowFlowProfileUdp) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowFlowProfileUdp is description is TBD
type DataflowFlowProfileUdp interface {
	Msg() *onexdatamodel.DataflowFlowProfileUdp
	SetMsg(*onexdatamodel.DataflowFlowProfileUdp) DataflowFlowProfileUdp
	// ToPbText marshals DataflowFlowProfileUdp to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowFlowProfileUdp to YAML text
	ToYaml() string
	// ToJson marshals DataflowFlowProfileUdp to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowFlowProfileUdp from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowFlowProfileUdp from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowFlowProfileUdp from JSON text
	FromJson(value string) error
	// Validate validates DataflowFlowProfileUdp
	Validate() error
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

type chaosBackgroundTrafficFlow struct {
	obj                    *onexdatamodel.ChaosBackgroundTrafficFlow
	fabricEntryPointHolder ChaosBackgroundTrafficFlowEntryPoint
	statelessHolder        ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
}

func NewChaosBackgroundTrafficFlow() ChaosBackgroundTrafficFlow {
	obj := chaosBackgroundTrafficFlow{obj: &onexdatamodel.ChaosBackgroundTrafficFlow{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlow) Msg() *onexdatamodel.ChaosBackgroundTrafficFlow {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlow) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficFlow {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlow) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlow) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaosBackgroundTrafficFlow) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlow) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlow) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlow) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlow) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlow) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlow) setNil() {
	obj.fabricEntryPointHolder = nil
	obj.statelessHolder = nil
}

// ChaosBackgroundTrafficFlow is description is TBD
type ChaosBackgroundTrafficFlow interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlow
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlow) ChaosBackgroundTrafficFlow
	// ToPbText marshals ChaosBackgroundTrafficFlow to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlow to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlow to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlow from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlow from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlow from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlow
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in ChaosBackgroundTrafficFlow.
	Name() string
	// SetName assigns string provided by user to ChaosBackgroundTrafficFlow
	SetName(value string) ChaosBackgroundTrafficFlow
	// HasName checks if Name has been set in ChaosBackgroundTrafficFlow
	HasName() bool
	// FabricEntryPoint returns ChaosBackgroundTrafficFlowEntryPoint, set in ChaosBackgroundTrafficFlow.
	// ChaosBackgroundTrafficFlowEntryPoint is description is TBD
	FabricEntryPoint() ChaosBackgroundTrafficFlowEntryPoint
	// SetFabricEntryPoint assigns ChaosBackgroundTrafficFlowEntryPoint provided by user to ChaosBackgroundTrafficFlow.
	// ChaosBackgroundTrafficFlowEntryPoint is description is TBD
	SetFabricEntryPoint(value ChaosBackgroundTrafficFlowEntryPoint) ChaosBackgroundTrafficFlow
	// HasFabricEntryPoint checks if FabricEntryPoint has been set in ChaosBackgroundTrafficFlow
	HasFabricEntryPoint() bool
	// Choice returns ChaosBackgroundTrafficFlowChoiceEnum, set in ChaosBackgroundTrafficFlow
	Choice() ChaosBackgroundTrafficFlowChoiceEnum
	// SetChoice assigns ChaosBackgroundTrafficFlowChoiceEnum provided by user to ChaosBackgroundTrafficFlow
	SetChoice(value ChaosBackgroundTrafficFlowChoiceEnum) ChaosBackgroundTrafficFlow
	// HasChoice checks if Choice has been set in ChaosBackgroundTrafficFlow
	HasChoice() bool
	// Stateless returns ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter, set in ChaosBackgroundTrafficFlow
	Stateless() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	setNil()
}

// Name returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlow) Name() string {

	return *obj.obj.Name

}

// Name returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlow) HasName() bool {
	return obj.obj.Name != nil
}

// SetName sets the string value in the ChaosBackgroundTrafficFlow object
// description is TBD
func (obj *chaosBackgroundTrafficFlow) SetName(value string) ChaosBackgroundTrafficFlow {

	obj.obj.Name = &value
	return obj
}

// FabricEntryPoint returns a ChaosBackgroundTrafficFlowEntryPoint
// description is TBD
func (obj *chaosBackgroundTrafficFlow) FabricEntryPoint() ChaosBackgroundTrafficFlowEntryPoint {
	if obj.obj.FabricEntryPoint == nil {
		obj.obj.FabricEntryPoint = NewChaosBackgroundTrafficFlowEntryPoint().Msg()
	}
	if obj.fabricEntryPointHolder == nil {
		obj.fabricEntryPointHolder = &chaosBackgroundTrafficFlowEntryPoint{obj: obj.obj.FabricEntryPoint}
	}
	return obj.fabricEntryPointHolder
}

// FabricEntryPoint returns a ChaosBackgroundTrafficFlowEntryPoint
// description is TBD
func (obj *chaosBackgroundTrafficFlow) HasFabricEntryPoint() bool {
	return obj.obj.FabricEntryPoint != nil
}

// SetFabricEntryPoint sets the ChaosBackgroundTrafficFlowEntryPoint value in the ChaosBackgroundTrafficFlow object
// description is TBD
func (obj *chaosBackgroundTrafficFlow) SetFabricEntryPoint(value ChaosBackgroundTrafficFlowEntryPoint) ChaosBackgroundTrafficFlow {

	obj.FabricEntryPoint().SetMsg(value.Msg())
	return obj
}

type ChaosBackgroundTrafficFlowChoiceEnum string

//  Enum of Choice on ChaosBackgroundTrafficFlow
var ChaosBackgroundTrafficFlowChoice = struct {
	STATELESS ChaosBackgroundTrafficFlowChoiceEnum
}{
	STATELESS: ChaosBackgroundTrafficFlowChoiceEnum("stateless"),
}

func (obj *chaosBackgroundTrafficFlow) Choice() ChaosBackgroundTrafficFlowChoiceEnum {
	return ChaosBackgroundTrafficFlowChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlow) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *chaosBackgroundTrafficFlow) SetChoice(value ChaosBackgroundTrafficFlowChoiceEnum) ChaosBackgroundTrafficFlow {
	intValue, ok := onexdatamodel.ChaosBackgroundTrafficFlow_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ChaosBackgroundTrafficFlowChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.ChaosBackgroundTrafficFlow_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Stateless = nil
	obj.statelessHolder = nil

	if value == ChaosBackgroundTrafficFlowChoice.STATELESS {
		obj.obj.Stateless = []*onexdatamodel.ChaosBackgroundTrafficFlowStateless{}
	}

	return obj
}

// Stateless returns a []ChaosBackgroundTrafficFlowStateless
// description is TBD
func (obj *chaosBackgroundTrafficFlow) Stateless() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	if len(obj.obj.Stateless) == 0 {
		obj.SetChoice(ChaosBackgroundTrafficFlowChoice.STATELESS)
	}
	if obj.statelessHolder == nil {
		obj.statelessHolder = newChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter().setMsg(obj)
	}
	return obj.statelessHolder
}

type chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter struct {
	obj                                      *chaosBackgroundTrafficFlow
	chaosBackgroundTrafficFlowStatelessSlice []ChaosBackgroundTrafficFlowStateless
}

func newChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	return &chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter{}
}

type ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter interface {
	setMsg(*chaosBackgroundTrafficFlow) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	Items() []ChaosBackgroundTrafficFlowStateless
	Add() ChaosBackgroundTrafficFlowStateless
	Append(items ...ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	Set(index int, newObj ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	Clear() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	clearHolderSlice() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
	appendHolderSlice(item ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter
}

func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) setMsg(msg *chaosBackgroundTrafficFlow) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	obj.clearHolderSlice()
	for _, val := range msg.obj.Stateless {
		obj.appendHolderSlice(&chaosBackgroundTrafficFlowStateless{obj: val})
	}
	obj.obj = msg
	return obj
}

func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) Items() []ChaosBackgroundTrafficFlowStateless {
	return obj.chaosBackgroundTrafficFlowStatelessSlice
}

func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) Add() ChaosBackgroundTrafficFlowStateless {
	newObj := &onexdatamodel.ChaosBackgroundTrafficFlowStateless{}
	obj.obj.obj.Stateless = append(obj.obj.obj.Stateless, newObj)
	newLibObj := &chaosBackgroundTrafficFlowStateless{obj: newObj}
	newLibObj.setDefault()
	obj.chaosBackgroundTrafficFlowStatelessSlice = append(obj.chaosBackgroundTrafficFlowStatelessSlice, newLibObj)
	return newLibObj
}

func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) Append(items ...ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	for _, item := range items {
		newObj := item.Msg()
		obj.obj.obj.Stateless = append(obj.obj.obj.Stateless, newObj)
		obj.chaosBackgroundTrafficFlowStatelessSlice = append(obj.chaosBackgroundTrafficFlowStatelessSlice, item)
	}
	return obj
}

func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) Set(index int, newObj ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	obj.obj.obj.Stateless[index] = newObj.Msg()
	obj.chaosBackgroundTrafficFlowStatelessSlice[index] = newObj
	return obj
}
func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) Clear() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	if len(obj.obj.obj.Stateless) > 0 {
		obj.obj.obj.Stateless = []*onexdatamodel.ChaosBackgroundTrafficFlowStateless{}
		obj.chaosBackgroundTrafficFlowStatelessSlice = []ChaosBackgroundTrafficFlowStateless{}
	}
	return obj
}
func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) clearHolderSlice() ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	if len(obj.chaosBackgroundTrafficFlowStatelessSlice) > 0 {
		obj.chaosBackgroundTrafficFlowStatelessSlice = []ChaosBackgroundTrafficFlowStateless{}
	}
	return obj
}
func (obj *chaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter) appendHolderSlice(item ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowChaosBackgroundTrafficFlowStatelessIter {
	obj.chaosBackgroundTrafficFlowStatelessSlice = append(obj.chaosBackgroundTrafficFlowStatelessSlice, item)
	return obj
}

func (obj *chaosBackgroundTrafficFlow) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.FabricEntryPoint != nil {
		obj.FabricEntryPoint().validateObj(set_default)
	}

	if obj.obj.Stateless != nil {

		if set_default {
			obj.Stateless().clearHolderSlice()
			for _, item := range obj.obj.Stateless {
				obj.Stateless().appendHolderSlice(&chaosBackgroundTrafficFlowStateless{obj: item})
			}
		}
		for _, item := range obj.Stateless().Items() {
			item.validateObj(set_default)
		}

	}

}

func (obj *chaosBackgroundTrafficFlow) setDefault() {

}

type metricsResponseFlowResultTcpInfo struct {
	obj *onexdatamodel.MetricsResponseFlowResultTcpInfo
}

func NewMetricsResponseFlowResultTcpInfo() MetricsResponseFlowResultTcpInfo {
	obj := metricsResponseFlowResultTcpInfo{obj: &onexdatamodel.MetricsResponseFlowResultTcpInfo{}}
	obj.setDefault()
	return &obj
}

func (obj *metricsResponseFlowResultTcpInfo) Msg() *onexdatamodel.MetricsResponseFlowResultTcpInfo {
	return obj.obj
}

func (obj *metricsResponseFlowResultTcpInfo) SetMsg(msg *onexdatamodel.MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResultTcpInfo {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *metricsResponseFlowResultTcpInfo) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *metricsResponseFlowResultTcpInfo) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *metricsResponseFlowResultTcpInfo) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *metricsResponseFlowResultTcpInfo) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// MetricsResponseFlowResultTcpInfo is tCP information for this flow
type MetricsResponseFlowResultTcpInfo interface {
	Msg() *onexdatamodel.MetricsResponseFlowResultTcpInfo
	SetMsg(*onexdatamodel.MetricsResponseFlowResultTcpInfo) MetricsResponseFlowResultTcpInfo
	// ToPbText marshals MetricsResponseFlowResultTcpInfo to protobuf text
	ToPbText() string
	// ToYaml marshals MetricsResponseFlowResultTcpInfo to YAML text
	ToYaml() string
	// ToJson marshals MetricsResponseFlowResultTcpInfo to JSON text
	ToJson() string
	// FromPbText unmarshals MetricsResponseFlowResultTcpInfo from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals MetricsResponseFlowResultTcpInfo from YAML text
	FromYaml(value string) error
	// FromJson unmarshals MetricsResponseFlowResultTcpInfo from JSON text
	FromJson(value string) error
	// Validate validates MetricsResponseFlowResultTcpInfo
	Validate() error
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

type switchHostLinkSwitchRef struct {
	obj *onexdatamodel.SwitchHostLinkSwitchRef
}

func NewSwitchHostLinkSwitchRef() SwitchHostLinkSwitchRef {
	obj := switchHostLinkSwitchRef{obj: &onexdatamodel.SwitchHostLinkSwitchRef{}}
	obj.setDefault()
	return &obj
}

func (obj *switchHostLinkSwitchRef) Msg() *onexdatamodel.SwitchHostLinkSwitchRef {
	return obj.obj
}

func (obj *switchHostLinkSwitchRef) SetMsg(msg *onexdatamodel.SwitchHostLinkSwitchRef) SwitchHostLinkSwitchRef {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *switchHostLinkSwitchRef) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *switchHostLinkSwitchRef) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *switchHostLinkSwitchRef) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *switchHostLinkSwitchRef) FromYaml(value string) error {
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

func (obj *switchHostLinkSwitchRef) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *switchHostLinkSwitchRef) FromJson(value string) error {
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

func (obj *switchHostLinkSwitchRef) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *switchHostLinkSwitchRef) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// SwitchHostLinkSwitchRef is location of the switch based on pod and switch index
type SwitchHostLinkSwitchRef interface {
	Msg() *onexdatamodel.SwitchHostLinkSwitchRef
	SetMsg(*onexdatamodel.SwitchHostLinkSwitchRef) SwitchHostLinkSwitchRef
	// ToPbText marshals SwitchHostLinkSwitchRef to protobuf text
	ToPbText() string
	// ToYaml marshals SwitchHostLinkSwitchRef to YAML text
	ToYaml() string
	// ToJson marshals SwitchHostLinkSwitchRef to JSON text
	ToJson() string
	// FromPbText unmarshals SwitchHostLinkSwitchRef from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals SwitchHostLinkSwitchRef from YAML text
	FromYaml(value string) error
	// FromJson unmarshals SwitchHostLinkSwitchRef from JSON text
	FromJson(value string) error
	// Validate validates SwitchHostLinkSwitchRef
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// PodIndex returns int32, set in SwitchHostLinkSwitchRef.
	PodIndex() int32
	// SetPodIndex assigns int32 provided by user to SwitchHostLinkSwitchRef
	SetPodIndex(value int32) SwitchHostLinkSwitchRef
	// HasPodIndex checks if PodIndex has been set in SwitchHostLinkSwitchRef
	HasPodIndex() bool
	// SwitchIndex returns int32, set in SwitchHostLinkSwitchRef.
	SwitchIndex() int32
	// SetSwitchIndex assigns int32 provided by user to SwitchHostLinkSwitchRef
	SetSwitchIndex(value int32) SwitchHostLinkSwitchRef
	// HasSwitchIndex checks if SwitchIndex has been set in SwitchHostLinkSwitchRef
	HasSwitchIndex() bool
}

// PodIndex returns a int32
// One-based index of the pod based on the number of pods in the fabric
func (obj *switchHostLinkSwitchRef) PodIndex() int32 {

	return *obj.obj.PodIndex

}

// PodIndex returns a int32
// One-based index of the pod based on the number of pods in the fabric
func (obj *switchHostLinkSwitchRef) HasPodIndex() bool {
	return obj.obj.PodIndex != nil
}

// SetPodIndex sets the int32 value in the SwitchHostLinkSwitchRef object
// One-based index of the pod based on the number of pods in the fabric
func (obj *switchHostLinkSwitchRef) SetPodIndex(value int32) SwitchHostLinkSwitchRef {

	obj.obj.PodIndex = &value
	return obj
}

// SwitchIndex returns a int32
// One-based index of the pod or rack switch in the indicated pod
func (obj *switchHostLinkSwitchRef) SwitchIndex() int32 {

	return *obj.obj.SwitchIndex

}

// SwitchIndex returns a int32
// One-based index of the pod or rack switch in the indicated pod
func (obj *switchHostLinkSwitchRef) HasSwitchIndex() bool {
	return obj.obj.SwitchIndex != nil
}

// SetSwitchIndex sets the int32 value in the SwitchHostLinkSwitchRef object
// One-based index of the pod or rack switch in the indicated pod
func (obj *switchHostLinkSwitchRef) SetSwitchIndex(value int32) SwitchHostLinkSwitchRef {

	obj.obj.SwitchIndex = &value
	return obj
}

func (obj *switchHostLinkSwitchRef) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.PodIndex != nil {
		if *obj.obj.PodIndex < 1 || *obj.obj.PodIndex > 2147483647 {
			validation = append(
				validation,
				fmt.Sprintf("1 <= SwitchHostLinkSwitchRef.PodIndex <= 2147483647 but Got %d", *obj.obj.PodIndex))
		}

	}

	if obj.obj.SwitchIndex != nil {
		if *obj.obj.SwitchIndex < 1 || *obj.obj.SwitchIndex > 2147483647 {
			validation = append(
				validation,
				fmt.Sprintf("1 <= SwitchHostLinkSwitchRef.SwitchIndex <= 2147483647 but Got %d", *obj.obj.SwitchIndex))
		}

	}

}

func (obj *switchHostLinkSwitchRef) setDefault() {
	if obj.obj.PodIndex == nil {
		obj.SetPodIndex(1)
	}
	if obj.obj.SwitchIndex == nil {
		obj.SetSwitchIndex(1)
	}

}

type fabricPodSwitch struct {
	obj *onexdatamodel.FabricPodSwitch
}

func NewFabricPodSwitch() FabricPodSwitch {
	obj := fabricPodSwitch{obj: &onexdatamodel.FabricPodSwitch{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricPodSwitch) Msg() *onexdatamodel.FabricPodSwitch {
	return obj.obj
}

func (obj *fabricPodSwitch) SetMsg(msg *onexdatamodel.FabricPodSwitch) FabricPodSwitch {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricPodSwitch) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricPodSwitch) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricPodSwitch) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPodSwitch) FromYaml(value string) error {
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

func (obj *fabricPodSwitch) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricPodSwitch) FromJson(value string) error {
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

func (obj *fabricPodSwitch) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricPodSwitch) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricPodSwitch is description is TBD
type FabricPodSwitch interface {
	Msg() *onexdatamodel.FabricPodSwitch
	SetMsg(*onexdatamodel.FabricPodSwitch) FabricPodSwitch
	// ToPbText marshals FabricPodSwitch to protobuf text
	ToPbText() string
	// ToYaml marshals FabricPodSwitch to YAML text
	ToYaml() string
	// ToJson marshals FabricPodSwitch to JSON text
	ToJson() string
	// FromPbText unmarshals FabricPodSwitch from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricPodSwitch from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricPodSwitch from JSON text
	FromJson(value string) error
	// Validate validates FabricPodSwitch
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Count returns int32, set in FabricPodSwitch.
	Count() int32
	// SetCount assigns int32 provided by user to FabricPodSwitch
	SetCount(value int32) FabricPodSwitch
	// HasCount checks if Count has been set in FabricPodSwitch
	HasCount() bool
	// PodToSpineOversubscription returns string, set in FabricPodSwitch.
	PodToSpineOversubscription() string
	// SetPodToSpineOversubscription assigns string provided by user to FabricPodSwitch
	SetPodToSpineOversubscription(value string) FabricPodSwitch
	// HasPodToSpineOversubscription checks if PodToSpineOversubscription has been set in FabricPodSwitch
	HasPodToSpineOversubscription() bool
	// UplinkEcmpMode returns FabricPodSwitchUplinkEcmpModeEnum, set in FabricPodSwitch
	UplinkEcmpMode() FabricPodSwitchUplinkEcmpModeEnum
	// SetUplinkEcmpMode assigns FabricPodSwitchUplinkEcmpModeEnum provided by user to FabricPodSwitch
	SetUplinkEcmpMode(value FabricPodSwitchUplinkEcmpModeEnum) FabricPodSwitch
	// HasUplinkEcmpMode checks if UplinkEcmpMode has been set in FabricPodSwitch
	HasUplinkEcmpMode() bool
	// DownlinkEcmpMode returns FabricPodSwitchDownlinkEcmpModeEnum, set in FabricPodSwitch
	DownlinkEcmpMode() FabricPodSwitchDownlinkEcmpModeEnum
	// SetDownlinkEcmpMode assigns FabricPodSwitchDownlinkEcmpModeEnum provided by user to FabricPodSwitch
	SetDownlinkEcmpMode(value FabricPodSwitchDownlinkEcmpModeEnum) FabricPodSwitch
	// HasDownlinkEcmpMode checks if DownlinkEcmpMode has been set in FabricPodSwitch
	HasDownlinkEcmpMode() bool
	// QosProfileName returns string, set in FabricPodSwitch.
	QosProfileName() string
	// SetQosProfileName assigns string provided by user to FabricPodSwitch
	SetQosProfileName(value string) FabricPodSwitch
	// HasQosProfileName checks if QosProfileName has been set in FabricPodSwitch
	HasQosProfileName() bool
}

// Count returns a int32
// description is TBD
func (obj *fabricPodSwitch) Count() int32 {

	return *obj.obj.Count

}

// Count returns a int32
// description is TBD
func (obj *fabricPodSwitch) HasCount() bool {
	return obj.obj.Count != nil
}

// SetCount sets the int32 value in the FabricPodSwitch object
// description is TBD
func (obj *fabricPodSwitch) SetCount(value int32) FabricPodSwitch {

	obj.obj.Count = &value
	return obj
}

// PodToSpineOversubscription returns a string
// oversubscription ratio of the pod switches
func (obj *fabricPodSwitch) PodToSpineOversubscription() string {

	return *obj.obj.PodToSpineOversubscription

}

// PodToSpineOversubscription returns a string
// oversubscription ratio of the pod switches
func (obj *fabricPodSwitch) HasPodToSpineOversubscription() bool {
	return obj.obj.PodToSpineOversubscription != nil
}

// SetPodToSpineOversubscription sets the string value in the FabricPodSwitch object
// oversubscription ratio of the pod switches
func (obj *fabricPodSwitch) SetPodToSpineOversubscription(value string) FabricPodSwitch {

	obj.obj.PodToSpineOversubscription = &value
	return obj
}

type FabricPodSwitchUplinkEcmpModeEnum string

//  Enum of UplinkEcmpMode on FabricPodSwitch
var FabricPodSwitchUplinkEcmpMode = struct {
	RANDOM_SPRAY FabricPodSwitchUplinkEcmpModeEnum
	HASH_3_TUPLE FabricPodSwitchUplinkEcmpModeEnum
	HASH_5_TUPLE FabricPodSwitchUplinkEcmpModeEnum
}{
	RANDOM_SPRAY: FabricPodSwitchUplinkEcmpModeEnum("random_spray"),
	HASH_3_TUPLE: FabricPodSwitchUplinkEcmpModeEnum("hash_3_tuple"),
	HASH_5_TUPLE: FabricPodSwitchUplinkEcmpModeEnum("hash_5_tuple"),
}

func (obj *fabricPodSwitch) UplinkEcmpMode() FabricPodSwitchUplinkEcmpModeEnum {
	return FabricPodSwitchUplinkEcmpModeEnum(obj.obj.UplinkEcmpMode.Enum().String())
}

// UplinkEcmpMode returns a string
// The algorithm for packet distribution over ECMP links.
// - random_spray randomly puts each packet on an ECMP member links
// - hash_3_tuple is a 3 tuple hash of ipv4 src, dst, protocol
// - hash_5_tuple is static_hash_ipv4_l4 but a different resulting RTAG7 hash mode
func (obj *fabricPodSwitch) HasUplinkEcmpMode() bool {
	return obj.obj.UplinkEcmpMode != nil
}

func (obj *fabricPodSwitch) SetUplinkEcmpMode(value FabricPodSwitchUplinkEcmpModeEnum) FabricPodSwitch {
	intValue, ok := onexdatamodel.FabricPodSwitch_UplinkEcmpMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricPodSwitchUplinkEcmpModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricPodSwitch_UplinkEcmpMode_Enum(intValue)
	obj.obj.UplinkEcmpMode = &enumValue

	return obj
}

type FabricPodSwitchDownlinkEcmpModeEnum string

//  Enum of DownlinkEcmpMode on FabricPodSwitch
var FabricPodSwitchDownlinkEcmpMode = struct {
	RANDOM_SPRAY FabricPodSwitchDownlinkEcmpModeEnum
	HASH_3_TUPLE FabricPodSwitchDownlinkEcmpModeEnum
	HASH_5_TUPLE FabricPodSwitchDownlinkEcmpModeEnum
}{
	RANDOM_SPRAY: FabricPodSwitchDownlinkEcmpModeEnum("random_spray"),
	HASH_3_TUPLE: FabricPodSwitchDownlinkEcmpModeEnum("hash_3_tuple"),
	HASH_5_TUPLE: FabricPodSwitchDownlinkEcmpModeEnum("hash_5_tuple"),
}

func (obj *fabricPodSwitch) DownlinkEcmpMode() FabricPodSwitchDownlinkEcmpModeEnum {
	return FabricPodSwitchDownlinkEcmpModeEnum(obj.obj.DownlinkEcmpMode.Enum().String())
}

// DownlinkEcmpMode returns a string
// The algorithm for packet distribution over ECMP links.
// - random_spray randomly puts each packet on an ECMP member links
// - hash_3_tuple is a 3 tuple hash of ipv4 src, dst, protocol
// - hash_5_tuple is static_hash_ipv4_l4 but a different resulting RTAG7 hash mode
func (obj *fabricPodSwitch) HasDownlinkEcmpMode() bool {
	return obj.obj.DownlinkEcmpMode != nil
}

func (obj *fabricPodSwitch) SetDownlinkEcmpMode(value FabricPodSwitchDownlinkEcmpModeEnum) FabricPodSwitch {
	intValue, ok := onexdatamodel.FabricPodSwitch_DownlinkEcmpMode_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on FabricPodSwitchDownlinkEcmpModeEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.FabricPodSwitch_DownlinkEcmpMode_Enum(intValue)
	obj.obj.DownlinkEcmpMode = &enumValue

	return obj
}

// QosProfileName returns a string
// The name of a qos profile associated with this pod switch.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricPodSwitch) QosProfileName() string {

	return *obj.obj.QosProfileName

}

// QosProfileName returns a string
// The name of a qos profile associated with this pod switch.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricPodSwitch) HasQosProfileName() bool {
	return obj.obj.QosProfileName != nil
}

// SetQosProfileName sets the string value in the FabricPodSwitch object
// The name of a qos profile associated with this pod switch.
//
// x-constraint:
// - #/components/schemas/QosProfile/properties/name
//
func (obj *fabricPodSwitch) SetQosProfileName(value string) FabricPodSwitch {

	obj.obj.QosProfileName = &value
	return obj
}

func (obj *fabricPodSwitch) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricPodSwitch) setDefault() {
	if obj.obj.Count == nil {
		obj.SetCount(1)
	}

}

type fabricRack struct {
	obj *onexdatamodel.FabricRack
}

func NewFabricRack() FabricRack {
	obj := fabricRack{obj: &onexdatamodel.FabricRack{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricRack) Msg() *onexdatamodel.FabricRack {
	return obj.obj
}

func (obj *fabricRack) SetMsg(msg *onexdatamodel.FabricRack) FabricRack {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricRack) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricRack) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricRack) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricRack) FromYaml(value string) error {
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

func (obj *fabricRack) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricRack) FromJson(value string) error {
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

func (obj *fabricRack) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricRack) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricRack is description is TBD
type FabricRack interface {
	Msg() *onexdatamodel.FabricRack
	SetMsg(*onexdatamodel.FabricRack) FabricRack
	// ToPbText marshals FabricRack to protobuf text
	ToPbText() string
	// ToYaml marshals FabricRack to YAML text
	ToYaml() string
	// ToJson marshals FabricRack to JSON text
	ToJson() string
	// FromPbText unmarshals FabricRack from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricRack from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricRack from JSON text
	FromJson(value string) error
	// Validate validates FabricRack
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Count returns int32, set in FabricRack.
	Count() int32
	// SetCount assigns int32 provided by user to FabricRack
	SetCount(value int32) FabricRack
	// HasCount checks if Count has been set in FabricRack
	HasCount() bool
	// RackProfileNames returns []string, set in FabricRack.
	RackProfileNames() []string
	// SetRackProfileNames assigns []string provided by user to FabricRack
	SetRackProfileNames(value []string) FabricRack
}

// Count returns a int32
// number of racks (and thus ToRs) in the pod
func (obj *fabricRack) Count() int32 {

	return *obj.obj.Count

}

// Count returns a int32
// number of racks (and thus ToRs) in the pod
func (obj *fabricRack) HasCount() bool {
	return obj.obj.Count != nil
}

// SetCount sets the int32 value in the FabricRack object
// number of racks (and thus ToRs) in the pod
func (obj *fabricRack) SetCount(value int32) FabricRack {

	obj.obj.Count = &value
	return obj
}

// RackProfileNames returns a []string
// The names of rack profiles associated with this rack.
func (obj *fabricRack) RackProfileNames() []string {
	if obj.obj.RackProfileNames == nil {
		obj.obj.RackProfileNames = make([]string, 0)
	}
	return obj.obj.RackProfileNames
}

// SetRackProfileNames sets the []string value in the FabricRack object
// The names of rack profiles associated with this rack.
func (obj *fabricRack) SetRackProfileNames(value []string) FabricRack {

	if obj.obj.RackProfileNames == nil {
		obj.obj.RackProfileNames = make([]string, 0)
	}
	obj.obj.RackProfileNames = value

	return obj
}

func (obj *fabricRack) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricRack) setDefault() {
	if obj.obj.Count == nil {
		obj.SetCount(1)
	}

}

type fabricQosProfilePacketClassificationMap struct {
	obj *onexdatamodel.FabricQosProfilePacketClassificationMap
}

func NewFabricQosProfilePacketClassificationMap() FabricQosProfilePacketClassificationMap {
	obj := fabricQosProfilePacketClassificationMap{obj: &onexdatamodel.FabricQosProfilePacketClassificationMap{}}
	obj.setDefault()
	return &obj
}

func (obj *fabricQosProfilePacketClassificationMap) Msg() *onexdatamodel.FabricQosProfilePacketClassificationMap {
	return obj.obj
}

func (obj *fabricQosProfilePacketClassificationMap) SetMsg(msg *onexdatamodel.FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassificationMap {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *fabricQosProfilePacketClassificationMap) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *fabricQosProfilePacketClassificationMap) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *fabricQosProfilePacketClassificationMap) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfilePacketClassificationMap) FromYaml(value string) error {
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

func (obj *fabricQosProfilePacketClassificationMap) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *fabricQosProfilePacketClassificationMap) FromJson(value string) error {
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

func (obj *fabricQosProfilePacketClassificationMap) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *fabricQosProfilePacketClassificationMap) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// FabricQosProfilePacketClassificationMap is description is TBD
type FabricQosProfilePacketClassificationMap interface {
	Msg() *onexdatamodel.FabricQosProfilePacketClassificationMap
	SetMsg(*onexdatamodel.FabricQosProfilePacketClassificationMap) FabricQosProfilePacketClassificationMap
	// ToPbText marshals FabricQosProfilePacketClassificationMap to protobuf text
	ToPbText() string
	// ToYaml marshals FabricQosProfilePacketClassificationMap to YAML text
	ToYaml() string
	// ToJson marshals FabricQosProfilePacketClassificationMap to JSON text
	ToJson() string
	// FromPbText unmarshals FabricQosProfilePacketClassificationMap from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals FabricQosProfilePacketClassificationMap from YAML text
	FromYaml(value string) error
	// FromJson unmarshals FabricQosProfilePacketClassificationMap from JSON text
	FromJson(value string) error
	// Validate validates FabricQosProfilePacketClassificationMap
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
}

func (obj *fabricQosProfilePacketClassificationMap) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *fabricQosProfilePacketClassificationMap) setDefault() {

}

type dataflowSimulatedComputeWorkload struct {
	obj *onexdatamodel.DataflowSimulatedComputeWorkload
}

func NewDataflowSimulatedComputeWorkload() DataflowSimulatedComputeWorkload {
	obj := dataflowSimulatedComputeWorkload{obj: &onexdatamodel.DataflowSimulatedComputeWorkload{}}
	obj.setDefault()
	return &obj
}

func (obj *dataflowSimulatedComputeWorkload) Msg() *onexdatamodel.DataflowSimulatedComputeWorkload {
	return obj.obj
}

func (obj *dataflowSimulatedComputeWorkload) SetMsg(msg *onexdatamodel.DataflowSimulatedComputeWorkload) DataflowSimulatedComputeWorkload {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *dataflowSimulatedComputeWorkload) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *dataflowSimulatedComputeWorkload) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *dataflowSimulatedComputeWorkload) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *dataflowSimulatedComputeWorkload) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// DataflowSimulatedComputeWorkload is description is TBD
type DataflowSimulatedComputeWorkload interface {
	Msg() *onexdatamodel.DataflowSimulatedComputeWorkload
	SetMsg(*onexdatamodel.DataflowSimulatedComputeWorkload) DataflowSimulatedComputeWorkload
	// ToPbText marshals DataflowSimulatedComputeWorkload to protobuf text
	ToPbText() string
	// ToYaml marshals DataflowSimulatedComputeWorkload to YAML text
	ToYaml() string
	// ToJson marshals DataflowSimulatedComputeWorkload to JSON text
	ToJson() string
	// FromPbText unmarshals DataflowSimulatedComputeWorkload from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals DataflowSimulatedComputeWorkload from YAML text
	FromYaml(value string) error
	// FromJson unmarshals DataflowSimulatedComputeWorkload from JSON text
	FromJson(value string) error
	// Validate validates DataflowSimulatedComputeWorkload
	Validate() error
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

type l4PortRange struct {
	obj               *onexdatamodel.L4PortRange
	singleValueHolder L4PortRangeSingleValue
	rangeHolder       L4PortRangeRange
}

func NewL4PortRange() L4PortRange {
	obj := l4PortRange{obj: &onexdatamodel.L4PortRange{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRange) Msg() *onexdatamodel.L4PortRange {
	return obj.obj
}

func (obj *l4PortRange) SetMsg(msg *onexdatamodel.L4PortRange) L4PortRange {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRange) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *l4PortRange) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *l4PortRange) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *l4PortRange) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *l4PortRange) setNil() {
	obj.singleValueHolder = nil
	obj.rangeHolder = nil
}

// L4PortRange is layer4 protocol source or destination port values
type L4PortRange interface {
	Msg() *onexdatamodel.L4PortRange
	SetMsg(*onexdatamodel.L4PortRange) L4PortRange
	// ToPbText marshals L4PortRange to protobuf text
	ToPbText() string
	// ToYaml marshals L4PortRange to YAML text
	ToYaml() string
	// ToJson marshals L4PortRange to JSON text
	ToJson() string
	// FromPbText unmarshals L4PortRange from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRange from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRange from JSON text
	FromJson(value string) error
	// Validate validates L4PortRange
	Validate() error
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
	intValue, ok := onexdatamodel.L4PortRange_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on L4PortRangeChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.L4PortRange_Choice_Enum(intValue)
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
	obj.SingleValue().SetMsg(value.Msg())
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
	obj.Range().SetMsg(value.Msg())
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

type chaosBackgroundTrafficFlowEntryPoint struct {
	obj                   *onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint
	switchReferenceHolder ChaosBackgroundTrafficFlowEntryPointSwitchReference
	frontPanelPortHolder  ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
}

func NewChaosBackgroundTrafficFlowEntryPoint() ChaosBackgroundTrafficFlowEntryPoint {
	obj := chaosBackgroundTrafficFlowEntryPoint{obj: &onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint) ChaosBackgroundTrafficFlowEntryPoint {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaosBackgroundTrafficFlowEntryPoint) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPoint) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPoint) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) setNil() {
	obj.switchReferenceHolder = nil
	obj.frontPanelPortHolder = nil
}

// ChaosBackgroundTrafficFlowEntryPoint is description is TBD
type ChaosBackgroundTrafficFlowEntryPoint interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint) ChaosBackgroundTrafficFlowEntryPoint
	// ToPbText marshals ChaosBackgroundTrafficFlowEntryPoint to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowEntryPoint to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowEntryPoint to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowEntryPoint from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowEntryPoint from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowEntryPoint from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowEntryPoint
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Choice returns ChaosBackgroundTrafficFlowEntryPointChoiceEnum, set in ChaosBackgroundTrafficFlowEntryPoint
	Choice() ChaosBackgroundTrafficFlowEntryPointChoiceEnum
	// SetChoice assigns ChaosBackgroundTrafficFlowEntryPointChoiceEnum provided by user to ChaosBackgroundTrafficFlowEntryPoint
	SetChoice(value ChaosBackgroundTrafficFlowEntryPointChoiceEnum) ChaosBackgroundTrafficFlowEntryPoint
	// HasChoice checks if Choice has been set in ChaosBackgroundTrafficFlowEntryPoint
	HasChoice() bool
	// SwitchReference returns ChaosBackgroundTrafficFlowEntryPointSwitchReference, set in ChaosBackgroundTrafficFlowEntryPoint.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReference is description is TBD
	SwitchReference() ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// SetSwitchReference assigns ChaosBackgroundTrafficFlowEntryPointSwitchReference provided by user to ChaosBackgroundTrafficFlowEntryPoint.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReference is description is TBD
	SetSwitchReference(value ChaosBackgroundTrafficFlowEntryPointSwitchReference) ChaosBackgroundTrafficFlowEntryPoint
	// HasSwitchReference checks if SwitchReference has been set in ChaosBackgroundTrafficFlowEntryPoint
	HasSwitchReference() bool
	// FrontPanelPort returns ChaosBackgroundTrafficFlowEntryPointFrontPanelPort, set in ChaosBackgroundTrafficFlowEntryPoint.
	// ChaosBackgroundTrafficFlowEntryPointFrontPanelPort is description is TBD
	FrontPanelPort() ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	// SetFrontPanelPort assigns ChaosBackgroundTrafficFlowEntryPointFrontPanelPort provided by user to ChaosBackgroundTrafficFlowEntryPoint.
	// ChaosBackgroundTrafficFlowEntryPointFrontPanelPort is description is TBD
	SetFrontPanelPort(value ChaosBackgroundTrafficFlowEntryPointFrontPanelPort) ChaosBackgroundTrafficFlowEntryPoint
	// HasFrontPanelPort checks if FrontPanelPort has been set in ChaosBackgroundTrafficFlowEntryPoint
	HasFrontPanelPort() bool
	setNil()
}

type ChaosBackgroundTrafficFlowEntryPointChoiceEnum string

//  Enum of Choice on ChaosBackgroundTrafficFlowEntryPoint
var ChaosBackgroundTrafficFlowEntryPointChoice = struct {
	SWITCH_REFERENCE ChaosBackgroundTrafficFlowEntryPointChoiceEnum
	FRONT_PANEL_PORT ChaosBackgroundTrafficFlowEntryPointChoiceEnum
}{
	SWITCH_REFERENCE: ChaosBackgroundTrafficFlowEntryPointChoiceEnum("switch_reference"),
	FRONT_PANEL_PORT: ChaosBackgroundTrafficFlowEntryPointChoiceEnum("front_panel_port"),
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) Choice() ChaosBackgroundTrafficFlowEntryPointChoiceEnum {
	return ChaosBackgroundTrafficFlowEntryPointChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) SetChoice(value ChaosBackgroundTrafficFlowEntryPointChoiceEnum) ChaosBackgroundTrafficFlowEntryPoint {
	intValue, ok := onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ChaosBackgroundTrafficFlowEntryPointChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.ChaosBackgroundTrafficFlowEntryPoint_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.FrontPanelPort = nil
	obj.frontPanelPortHolder = nil
	obj.obj.SwitchReference = nil
	obj.switchReferenceHolder = nil

	if value == ChaosBackgroundTrafficFlowEntryPointChoice.SWITCH_REFERENCE {
		obj.obj.SwitchReference = NewChaosBackgroundTrafficFlowEntryPointSwitchReference().Msg()
	}

	if value == ChaosBackgroundTrafficFlowEntryPointChoice.FRONT_PANEL_PORT {
		obj.obj.FrontPanelPort = NewChaosBackgroundTrafficFlowEntryPointFrontPanelPort().Msg()
	}

	return obj
}

// SwitchReference returns a ChaosBackgroundTrafficFlowEntryPointSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) SwitchReference() ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	if obj.obj.SwitchReference == nil {
		obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointChoice.SWITCH_REFERENCE)
	}
	if obj.switchReferenceHolder == nil {
		obj.switchReferenceHolder = &chaosBackgroundTrafficFlowEntryPointSwitchReference{obj: obj.obj.SwitchReference}
	}
	return obj.switchReferenceHolder
}

// SwitchReference returns a ChaosBackgroundTrafficFlowEntryPointSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) HasSwitchReference() bool {
	return obj.obj.SwitchReference != nil
}

// SetSwitchReference sets the ChaosBackgroundTrafficFlowEntryPointSwitchReference value in the ChaosBackgroundTrafficFlowEntryPoint object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) SetSwitchReference(value ChaosBackgroundTrafficFlowEntryPointSwitchReference) ChaosBackgroundTrafficFlowEntryPoint {
	obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointChoice.SWITCH_REFERENCE)
	obj.SwitchReference().SetMsg(value.Msg())
	return obj
}

// FrontPanelPort returns a ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) FrontPanelPort() ChaosBackgroundTrafficFlowEntryPointFrontPanelPort {
	if obj.obj.FrontPanelPort == nil {
		obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointChoice.FRONT_PANEL_PORT)
	}
	if obj.frontPanelPortHolder == nil {
		obj.frontPanelPortHolder = &chaosBackgroundTrafficFlowEntryPointFrontPanelPort{obj: obj.obj.FrontPanelPort}
	}
	return obj.frontPanelPortHolder
}

// FrontPanelPort returns a ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) HasFrontPanelPort() bool {
	return obj.obj.FrontPanelPort != nil
}

// SetFrontPanelPort sets the ChaosBackgroundTrafficFlowEntryPointFrontPanelPort value in the ChaosBackgroundTrafficFlowEntryPoint object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPoint) SetFrontPanelPort(value ChaosBackgroundTrafficFlowEntryPointFrontPanelPort) ChaosBackgroundTrafficFlowEntryPoint {
	obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointChoice.FRONT_PANEL_PORT)
	obj.FrontPanelPort().SetMsg(value.Msg())
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPoint) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.SwitchReference != nil {
		obj.SwitchReference().validateObj(set_default)
	}

	if obj.obj.FrontPanelPort != nil {
		obj.FrontPanelPort().validateObj(set_default)
	}

}

func (obj *chaosBackgroundTrafficFlowEntryPoint) setDefault() {

}

type chaosBackgroundTrafficFlowStateless struct {
	obj          *onexdatamodel.ChaosBackgroundTrafficFlowStateless
	packetHolder ChaosBackgroundTrafficFlowStatelessPacket
}

func NewChaosBackgroundTrafficFlowStateless() ChaosBackgroundTrafficFlowStateless {
	obj := chaosBackgroundTrafficFlowStateless{obj: &onexdatamodel.ChaosBackgroundTrafficFlowStateless{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowStateless) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowStateless {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowStateless) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowStateless {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowStateless) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowStateless) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaosBackgroundTrafficFlowStateless) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowStateless) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowStateless) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowStateless) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowStateless) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowStateless) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowStateless) setNil() {
	obj.packetHolder = nil
}

// ChaosBackgroundTrafficFlowStateless is description is TBD
type ChaosBackgroundTrafficFlowStateless interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowStateless
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowStateless) ChaosBackgroundTrafficFlowStateless
	// ToPbText marshals ChaosBackgroundTrafficFlowStateless to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowStateless to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowStateless to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowStateless from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowStateless from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowStateless from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowStateless
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Name returns string, set in ChaosBackgroundTrafficFlowStateless.
	Name() string
	// SetName assigns string provided by user to ChaosBackgroundTrafficFlowStateless
	SetName(value string) ChaosBackgroundTrafficFlowStateless
	// HasName checks if Name has been set in ChaosBackgroundTrafficFlowStateless
	HasName() bool
	// Rate returns int32, set in ChaosBackgroundTrafficFlowStateless.
	Rate() int32
	// SetRate assigns int32 provided by user to ChaosBackgroundTrafficFlowStateless
	SetRate(value int32) ChaosBackgroundTrafficFlowStateless
	// HasRate checks if Rate has been set in ChaosBackgroundTrafficFlowStateless
	HasRate() bool
	// RateUnit returns ChaosBackgroundTrafficFlowStatelessRateUnitEnum, set in ChaosBackgroundTrafficFlowStateless
	RateUnit() ChaosBackgroundTrafficFlowStatelessRateUnitEnum
	// SetRateUnit assigns ChaosBackgroundTrafficFlowStatelessRateUnitEnum provided by user to ChaosBackgroundTrafficFlowStateless
	SetRateUnit(value ChaosBackgroundTrafficFlowStatelessRateUnitEnum) ChaosBackgroundTrafficFlowStateless
	// HasRateUnit checks if RateUnit has been set in ChaosBackgroundTrafficFlowStateless
	HasRateUnit() bool
	// Packet returns ChaosBackgroundTrafficFlowStatelessPacket, set in ChaosBackgroundTrafficFlowStateless.
	// ChaosBackgroundTrafficFlowStatelessPacket is description is TBD
	Packet() ChaosBackgroundTrafficFlowStatelessPacket
	// SetPacket assigns ChaosBackgroundTrafficFlowStatelessPacket provided by user to ChaosBackgroundTrafficFlowStateless.
	// ChaosBackgroundTrafficFlowStatelessPacket is description is TBD
	SetPacket(value ChaosBackgroundTrafficFlowStatelessPacket) ChaosBackgroundTrafficFlowStateless
	// HasPacket checks if Packet has been set in ChaosBackgroundTrafficFlowStateless
	HasPacket() bool
	setNil()
}

// Name returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) Name() string {

	return *obj.obj.Name

}

// Name returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) HasName() bool {
	return obj.obj.Name != nil
}

// SetName sets the string value in the ChaosBackgroundTrafficFlowStateless object
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) SetName(value string) ChaosBackgroundTrafficFlowStateless {

	obj.obj.Name = &value
	return obj
}

// Rate returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) Rate() int32 {

	return *obj.obj.Rate

}

// Rate returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) HasRate() bool {
	return obj.obj.Rate != nil
}

// SetRate sets the int32 value in the ChaosBackgroundTrafficFlowStateless object
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) SetRate(value int32) ChaosBackgroundTrafficFlowStateless {

	obj.obj.Rate = &value
	return obj
}

type ChaosBackgroundTrafficFlowStatelessRateUnitEnum string

//  Enum of RateUnit on ChaosBackgroundTrafficFlowStateless
var ChaosBackgroundTrafficFlowStatelessRateUnit = struct {
	BPS  ChaosBackgroundTrafficFlowStatelessRateUnitEnum
	KBPS ChaosBackgroundTrafficFlowStatelessRateUnitEnum
	MBPS ChaosBackgroundTrafficFlowStatelessRateUnitEnum
	GBPS ChaosBackgroundTrafficFlowStatelessRateUnitEnum
}{
	BPS:  ChaosBackgroundTrafficFlowStatelessRateUnitEnum("bps"),
	KBPS: ChaosBackgroundTrafficFlowStatelessRateUnitEnum("Kbps"),
	MBPS: ChaosBackgroundTrafficFlowStatelessRateUnitEnum("Mbps"),
	GBPS: ChaosBackgroundTrafficFlowStatelessRateUnitEnum("Gbps"),
}

func (obj *chaosBackgroundTrafficFlowStateless) RateUnit() ChaosBackgroundTrafficFlowStatelessRateUnitEnum {
	return ChaosBackgroundTrafficFlowStatelessRateUnitEnum(obj.obj.RateUnit.Enum().String())
}

// RateUnit returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) HasRateUnit() bool {
	return obj.obj.RateUnit != nil
}

func (obj *chaosBackgroundTrafficFlowStateless) SetRateUnit(value ChaosBackgroundTrafficFlowStatelessRateUnitEnum) ChaosBackgroundTrafficFlowStateless {
	intValue, ok := onexdatamodel.ChaosBackgroundTrafficFlowStateless_RateUnit_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ChaosBackgroundTrafficFlowStatelessRateUnitEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.ChaosBackgroundTrafficFlowStateless_RateUnit_Enum(intValue)
	obj.obj.RateUnit = &enumValue

	return obj
}

// Packet returns a ChaosBackgroundTrafficFlowStatelessPacket
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) Packet() ChaosBackgroundTrafficFlowStatelessPacket {
	if obj.obj.Packet == nil {
		obj.obj.Packet = NewChaosBackgroundTrafficFlowStatelessPacket().Msg()
	}
	if obj.packetHolder == nil {
		obj.packetHolder = &chaosBackgroundTrafficFlowStatelessPacket{obj: obj.obj.Packet}
	}
	return obj.packetHolder
}

// Packet returns a ChaosBackgroundTrafficFlowStatelessPacket
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) HasPacket() bool {
	return obj.obj.Packet != nil
}

// SetPacket sets the ChaosBackgroundTrafficFlowStatelessPacket value in the ChaosBackgroundTrafficFlowStateless object
// description is TBD
func (obj *chaosBackgroundTrafficFlowStateless) SetPacket(value ChaosBackgroundTrafficFlowStatelessPacket) ChaosBackgroundTrafficFlowStateless {

	obj.Packet().SetMsg(value.Msg())
	return obj
}

func (obj *chaosBackgroundTrafficFlowStateless) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Packet != nil {
		obj.Packet().validateObj(set_default)
	}

}

func (obj *chaosBackgroundTrafficFlowStateless) setDefault() {

}

type l4PortRangeSingleValue struct {
	obj *onexdatamodel.L4PortRangeSingleValue
}

func NewL4PortRangeSingleValue() L4PortRangeSingleValue {
	obj := l4PortRangeSingleValue{obj: &onexdatamodel.L4PortRangeSingleValue{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRangeSingleValue) Msg() *onexdatamodel.L4PortRangeSingleValue {
	return obj.obj
}

func (obj *l4PortRangeSingleValue) SetMsg(msg *onexdatamodel.L4PortRangeSingleValue) L4PortRangeSingleValue {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRangeSingleValue) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *l4PortRangeSingleValue) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *l4PortRangeSingleValue) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *l4PortRangeSingleValue) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// L4PortRangeSingleValue is description is TBD
type L4PortRangeSingleValue interface {
	Msg() *onexdatamodel.L4PortRangeSingleValue
	SetMsg(*onexdatamodel.L4PortRangeSingleValue) L4PortRangeSingleValue
	// ToPbText marshals L4PortRangeSingleValue to protobuf text
	ToPbText() string
	// ToYaml marshals L4PortRangeSingleValue to YAML text
	ToYaml() string
	// ToJson marshals L4PortRangeSingleValue to JSON text
	ToJson() string
	// FromPbText unmarshals L4PortRangeSingleValue from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRangeSingleValue from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRangeSingleValue from JSON text
	FromJson(value string) error
	// Validate validates L4PortRangeSingleValue
	Validate() error
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

type l4PortRangeRange struct {
	obj *onexdatamodel.L4PortRangeRange
}

func NewL4PortRangeRange() L4PortRangeRange {
	obj := l4PortRangeRange{obj: &onexdatamodel.L4PortRangeRange{}}
	obj.setDefault()
	return &obj
}

func (obj *l4PortRangeRange) Msg() *onexdatamodel.L4PortRangeRange {
	return obj.obj
}

func (obj *l4PortRangeRange) SetMsg(msg *onexdatamodel.L4PortRangeRange) L4PortRangeRange {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *l4PortRangeRange) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *l4PortRangeRange) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *l4PortRangeRange) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
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

func (obj *l4PortRangeRange) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
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

// L4PortRangeRange is description is TBD
type L4PortRangeRange interface {
	Msg() *onexdatamodel.L4PortRangeRange
	SetMsg(*onexdatamodel.L4PortRangeRange) L4PortRangeRange
	// ToPbText marshals L4PortRangeRange to protobuf text
	ToPbText() string
	// ToYaml marshals L4PortRangeRange to YAML text
	ToYaml() string
	// ToJson marshals L4PortRangeRange to JSON text
	ToJson() string
	// FromPbText unmarshals L4PortRangeRange from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals L4PortRangeRange from YAML text
	FromYaml(value string) error
	// FromJson unmarshals L4PortRangeRange from JSON text
	FromJson(value string) error
	// Validate validates L4PortRangeRange
	Validate() error
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

type chaosBackgroundTrafficFlowEntryPointSwitchReference struct {
	obj         *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference
	spineHolder ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	podHolder   ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	torHolder   ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
}

func NewChaosBackgroundTrafficFlowEntryPointSwitchReference() ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	obj := chaosBackgroundTrafficFlowEntryPointSwitchReference{obj: &onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) setNil() {
	obj.spineHolder = nil
	obj.podHolder = nil
	obj.torHolder = nil
}

// ChaosBackgroundTrafficFlowEntryPointSwitchReference is description is TBD
type ChaosBackgroundTrafficFlowEntryPointSwitchReference interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// ToPbText marshals ChaosBackgroundTrafficFlowEntryPointSwitchReference to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowEntryPointSwitchReference to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowEntryPointSwitchReference to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReference from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReference from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReference from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowEntryPointSwitchReference
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// Choice returns ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum, set in ChaosBackgroundTrafficFlowEntryPointSwitchReference
	Choice() ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum
	// SetChoice assigns ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReference
	SetChoice(value ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum) ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// HasChoice checks if Choice has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReference
	HasChoice() bool
	// Spine returns ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine, set in ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine is description is TBD
	Spine() ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	// SetSpine assigns ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine is description is TBD
	SetSpine(value ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// HasSpine checks if Spine has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReference
	HasSpine() bool
	// Pod returns ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference, set in ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference is description is TBD
	Pod() ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	// SetPod assigns ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference is description is TBD
	SetPod(value ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// HasPod checks if Pod has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReference
	HasPod() bool
	// Tor returns ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference, set in ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference is description is TBD
	Tor() ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	// SetTor assigns ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReference.
	// ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference is description is TBD
	SetTor(value ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference
	// HasTor checks if Tor has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReference
	HasTor() bool
	setNil()
}

type ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum string

//  Enum of Choice on ChaosBackgroundTrafficFlowEntryPointSwitchReference
var ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice = struct {
	SPINE ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum
	POD   ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum
	TOR   ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum
}{
	SPINE: ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum("spine"),
	POD:   ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum("pod"),
	TOR:   ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum("tor"),
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Choice() ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum {
	return ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum(obj.obj.Choice.Enum().String())
}

// Choice returns a string
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) HasChoice() bool {
	return obj.obj.Choice != nil
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) SetChoice(value ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum) ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	intValue, ok := onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference_Choice_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoiceEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReference_Choice_Enum(intValue)
	obj.obj.Choice = &enumValue
	obj.obj.Tor = nil
	obj.torHolder = nil
	obj.obj.Pod = nil
	obj.podHolder = nil
	obj.obj.Spine = nil
	obj.spineHolder = nil

	if value == ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.SPINE {
		obj.obj.Spine = NewChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine().Msg()
	}

	if value == ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.POD {
		obj.obj.Pod = NewChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference().Msg()
	}

	if value == ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.TOR {
		obj.obj.Tor = NewChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference().Msg()
	}

	return obj
}

// Spine returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Spine() ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine {
	if obj.obj.Spine == nil {
		obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.SPINE)
	}
	if obj.spineHolder == nil {
		obj.spineHolder = &chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine{obj: obj.obj.Spine}
	}
	return obj.spineHolder
}

// Spine returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) HasSpine() bool {
	return obj.obj.Spine != nil
}

// SetSpine sets the ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine value in the ChaosBackgroundTrafficFlowEntryPointSwitchReference object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) SetSpine(value ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.SPINE)
	obj.Spine().SetMsg(value.Msg())
	return obj
}

// Pod returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Pod() ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {
	if obj.obj.Pod == nil {
		obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.POD)
	}
	if obj.podHolder == nil {
		obj.podHolder = &chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference{obj: obj.obj.Pod}
	}
	return obj.podHolder
}

// Pod returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) HasPod() bool {
	return obj.obj.Pod != nil
}

// SetPod sets the ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference value in the ChaosBackgroundTrafficFlowEntryPointSwitchReference object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) SetPod(value ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.POD)
	obj.Pod().SetMsg(value.Msg())
	return obj
}

// Tor returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) Tor() ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {
	if obj.obj.Tor == nil {
		obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.TOR)
	}
	if obj.torHolder == nil {
		obj.torHolder = &chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference{obj: obj.obj.Tor}
	}
	return obj.torHolder
}

// Tor returns a ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) HasTor() bool {
	return obj.obj.Tor != nil
}

// SetTor sets the ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference value in the ChaosBackgroundTrafficFlowEntryPointSwitchReference object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) SetTor(value ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReference {
	obj.SetChoice(ChaosBackgroundTrafficFlowEntryPointSwitchReferenceChoice.TOR)
	obj.Tor().SetMsg(value.Msg())
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.Spine != nil {
		obj.Spine().validateObj(set_default)
	}

	if obj.obj.Pod != nil {
		obj.Pod().validateObj(set_default)
	}

	if obj.obj.Tor != nil {
		obj.Tor().validateObj(set_default)
	}

}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReference) setDefault() {

}

type chaosBackgroundTrafficFlowEntryPointFrontPanelPort struct {
	obj *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
}

func NewChaosBackgroundTrafficFlowEntryPointFrontPanelPort() ChaosBackgroundTrafficFlowEntryPointFrontPanelPort {
	obj := chaosBackgroundTrafficFlowEntryPointFrontPanelPort{obj: &onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort) ChaosBackgroundTrafficFlowEntryPointFrontPanelPort {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// ChaosBackgroundTrafficFlowEntryPointFrontPanelPort is description is TBD
type ChaosBackgroundTrafficFlowEntryPointFrontPanelPort interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowEntryPointFrontPanelPort) ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	// ToPbText marshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowEntryPointFrontPanelPort from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// FrontPanelPort returns int32, set in ChaosBackgroundTrafficFlowEntryPointFrontPanelPort.
	FrontPanelPort() int32
	// SetFrontPanelPort assigns int32 provided by user to ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	SetFrontPanelPort(value int32) ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	// HasFrontPanelPort checks if FrontPanelPort has been set in ChaosBackgroundTrafficFlowEntryPointFrontPanelPort
	HasFrontPanelPort() bool
}

// FrontPanelPort returns a int32
// Front panel port number (if fabric is emulated on a device with front panel ports)
func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) FrontPanelPort() int32 {

	return *obj.obj.FrontPanelPort

}

// FrontPanelPort returns a int32
// Front panel port number (if fabric is emulated on a device with front panel ports)
func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) HasFrontPanelPort() bool {
	return obj.obj.FrontPanelPort != nil
}

// SetFrontPanelPort sets the int32 value in the ChaosBackgroundTrafficFlowEntryPointFrontPanelPort object
// Front panel port number (if fabric is emulated on a device with front panel ports)
func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) SetFrontPanelPort(value int32) ChaosBackgroundTrafficFlowEntryPointFrontPanelPort {

	obj.obj.FrontPanelPort = &value
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *chaosBackgroundTrafficFlowEntryPointFrontPanelPort) setDefault() {

}

type chaosBackgroundTrafficFlowStatelessPacket struct {
	obj *onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket
}

func NewChaosBackgroundTrafficFlowStatelessPacket() ChaosBackgroundTrafficFlowStatelessPacket {
	obj := chaosBackgroundTrafficFlowStatelessPacket{obj: &onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket) ChaosBackgroundTrafficFlowStatelessPacket {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowStatelessPacket) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowStatelessPacket) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// ChaosBackgroundTrafficFlowStatelessPacket is description is TBD
type ChaosBackgroundTrafficFlowStatelessPacket interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket) ChaosBackgroundTrafficFlowStatelessPacket
	// ToPbText marshals ChaosBackgroundTrafficFlowStatelessPacket to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowStatelessPacket to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowStatelessPacket to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowStatelessPacket from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowStatelessPacket from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowStatelessPacket from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowStatelessPacket
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// SrcAddress returns string, set in ChaosBackgroundTrafficFlowStatelessPacket.
	SrcAddress() string
	// SetSrcAddress assigns string provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetSrcAddress(value string) ChaosBackgroundTrafficFlowStatelessPacket
	// HasSrcAddress checks if SrcAddress has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasSrcAddress() bool
	// DstAddress returns string, set in ChaosBackgroundTrafficFlowStatelessPacket.
	DstAddress() string
	// SetDstAddress assigns string provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetDstAddress(value string) ChaosBackgroundTrafficFlowStatelessPacket
	// HasDstAddress checks if DstAddress has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasDstAddress() bool
	// SrcPort returns int32, set in ChaosBackgroundTrafficFlowStatelessPacket.
	SrcPort() int32
	// SetSrcPort assigns int32 provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetSrcPort(value int32) ChaosBackgroundTrafficFlowStatelessPacket
	// HasSrcPort checks if SrcPort has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasSrcPort() bool
	// DstPort returns int32, set in ChaosBackgroundTrafficFlowStatelessPacket.
	DstPort() int32
	// SetDstPort assigns int32 provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetDstPort(value int32) ChaosBackgroundTrafficFlowStatelessPacket
	// HasDstPort checks if DstPort has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasDstPort() bool
	// Size returns int32, set in ChaosBackgroundTrafficFlowStatelessPacket.
	Size() int32
	// SetSize assigns int32 provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetSize(value int32) ChaosBackgroundTrafficFlowStatelessPacket
	// HasSize checks if Size has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasSize() bool
	// L4Protocol returns ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum, set in ChaosBackgroundTrafficFlowStatelessPacket
	L4Protocol() ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum
	// SetL4Protocol assigns ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum provided by user to ChaosBackgroundTrafficFlowStatelessPacket
	SetL4Protocol(value ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum) ChaosBackgroundTrafficFlowStatelessPacket
	// HasL4Protocol checks if L4Protocol has been set in ChaosBackgroundTrafficFlowStatelessPacket
	HasL4Protocol() bool
}

// SrcAddress returns a string
// source IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SrcAddress() string {

	return *obj.obj.SrcAddress

}

// SrcAddress returns a string
// source IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasSrcAddress() bool {
	return obj.obj.SrcAddress != nil
}

// SetSrcAddress sets the string value in the ChaosBackgroundTrafficFlowStatelessPacket object
// source IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetSrcAddress(value string) ChaosBackgroundTrafficFlowStatelessPacket {

	obj.obj.SrcAddress = &value
	return obj
}

// DstAddress returns a string
// destination IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) DstAddress() string {

	return *obj.obj.DstAddress

}

// DstAddress returns a string
// destination IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasDstAddress() bool {
	return obj.obj.DstAddress != nil
}

// SetDstAddress sets the string value in the ChaosBackgroundTrafficFlowStatelessPacket object
// destination IP address
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetDstAddress(value string) ChaosBackgroundTrafficFlowStatelessPacket {

	obj.obj.DstAddress = &value
	return obj
}

// SrcPort returns a int32
// Layer 4 source port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SrcPort() int32 {

	return *obj.obj.SrcPort

}

// SrcPort returns a int32
// Layer 4 source port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasSrcPort() bool {
	return obj.obj.SrcPort != nil
}

// SetSrcPort sets the int32 value in the ChaosBackgroundTrafficFlowStatelessPacket object
// Layer 4 source port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetSrcPort(value int32) ChaosBackgroundTrafficFlowStatelessPacket {

	obj.obj.SrcPort = &value
	return obj
}

// DstPort returns a int32
// Layer 4 destination port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) DstPort() int32 {

	return *obj.obj.DstPort

}

// DstPort returns a int32
// Layer 4 destination port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasDstPort() bool {
	return obj.obj.DstPort != nil
}

// SetDstPort sets the int32 value in the ChaosBackgroundTrafficFlowStatelessPacket object
// Layer 4 destination port
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetDstPort(value int32) ChaosBackgroundTrafficFlowStatelessPacket {

	obj.obj.DstPort = &value
	return obj
}

// Size returns a int32
// total packet size
func (obj *chaosBackgroundTrafficFlowStatelessPacket) Size() int32 {

	return *obj.obj.Size

}

// Size returns a int32
// total packet size
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasSize() bool {
	return obj.obj.Size != nil
}

// SetSize sets the int32 value in the ChaosBackgroundTrafficFlowStatelessPacket object
// total packet size
func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetSize(value int32) ChaosBackgroundTrafficFlowStatelessPacket {

	obj.obj.Size = &value
	return obj
}

type ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum string

//  Enum of L4Protocol on ChaosBackgroundTrafficFlowStatelessPacket
var ChaosBackgroundTrafficFlowStatelessPacketL4Protocol = struct {
	TCP ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum
	UDP ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum
}{
	TCP: ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum("tcp"),
	UDP: ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum("udp"),
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) L4Protocol() ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum {
	return ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum(obj.obj.L4Protocol.Enum().String())
}

// L4Protocol returns a string
// Layer 4 transport protocol
func (obj *chaosBackgroundTrafficFlowStatelessPacket) HasL4Protocol() bool {
	return obj.obj.L4Protocol != nil
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) SetL4Protocol(value ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum) ChaosBackgroundTrafficFlowStatelessPacket {
	intValue, ok := onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket_L4Protocol_Enum_value[string(value)]
	if !ok {
		validation = append(validation, fmt.Sprintf(
			"%s is not a valid choice on ChaosBackgroundTrafficFlowStatelessPacketL4ProtocolEnum", string(value)))
		return obj
	}
	enumValue := onexdatamodel.ChaosBackgroundTrafficFlowStatelessPacket_L4Protocol_Enum(intValue)
	obj.obj.L4Protocol = &enumValue

	return obj
}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

	if obj.obj.SrcPort != nil {
		if *obj.obj.SrcPort < -2147483648 || *obj.obj.SrcPort > 65535 {
			validation = append(
				validation,
				fmt.Sprintf("-2147483648 <= ChaosBackgroundTrafficFlowStatelessPacket.SrcPort <= 65535 but Got %d", *obj.obj.SrcPort))
		}

	}

	if obj.obj.DstPort != nil {
		if *obj.obj.DstPort < -2147483648 || *obj.obj.DstPort > 65535 {
			validation = append(
				validation,
				fmt.Sprintf("-2147483648 <= ChaosBackgroundTrafficFlowStatelessPacket.DstPort <= 65535 but Got %d", *obj.obj.DstPort))
		}

	}

	if obj.obj.Size != nil {
		if *obj.obj.Size < 64 || *obj.obj.Size > 2147483647 {
			validation = append(
				validation,
				fmt.Sprintf("64 <= ChaosBackgroundTrafficFlowStatelessPacket.Size <= 2147483647 but Got %d", *obj.obj.Size))
		}

	}

}

func (obj *chaosBackgroundTrafficFlowStatelessPacket) setDefault() {
	if obj.obj.SrcPort == nil {
		obj.SetSrcPort(1024)
	}
	if obj.obj.DstPort == nil {
		obj.SetDstPort(1024)
	}
	if obj.obj.Size == nil {
		obj.SetSize(1000)
	}

}

type chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine struct {
	obj *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
}

func NewChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine() ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine {
	obj := chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine{obj: &onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine is description is TBD
type ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	// ToPbText marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// SwitchIndex returns int32, set in ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine.
	SwitchIndex() int32
	// SetSwitchIndex assigns int32 provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	SetSwitchIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	// HasSwitchIndex checks if SwitchIndex has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine
	HasSwitchIndex() bool
}

// SwitchIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) SwitchIndex() int32 {

	return *obj.obj.SwitchIndex

}

// SwitchIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) HasSwitchIndex() bool {
	return obj.obj.SwitchIndex != nil
}

// SetSwitchIndex sets the int32 value in the ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) SetSwitchIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine {

	obj.obj.SwitchIndex = &value
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferenceSpine) setDefault() {

}

type chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference struct {
	obj *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
}

func NewChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference() ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {
	obj := chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference{obj: &onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference{}}
	obj.setDefault()
	return &obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {
	return obj.obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) SetMsg(msg *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {
	proto.Merge(obj.obj, msg)
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ToPbText() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	return proto.MarshalTextString(obj.Msg())
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) FromPbText(value string) error {
	retObj := proto.UnmarshalText(value, obj.Msg())
	if retObj != nil {
		return retObj
	}

	vErr := obj.validateFromText()
	if vErr != nil {
		return vErr
	}
	return retObj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ToYaml() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) FromYaml(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ToJson() string {
	vErr := obj.Validate()
	if vErr != nil {
		panic(vErr)
	}
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	data, err := opts.Marshal(obj.Msg())
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) FromJson(value string) error {
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

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) validateFromText() error {
	obj.validateObj(true)
	return validationResult()
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) Validate() error {
	obj.validateObj(false)
	return validationResult()
}

// ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference is description is TBD
type ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference interface {
	Msg() *onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	SetMsg(*onexdatamodel.ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	// ToPbText marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference to protobuf text
	ToPbText() string
	// ToYaml marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference to YAML text
	ToYaml() string
	// ToJson marshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference to JSON text
	ToJson() string
	// FromPbText unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference from protobuf text
	FromPbText(value string) error
	// FromYaml unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference from YAML text
	FromYaml(value string) error
	// FromJson unmarshals ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference from JSON text
	FromJson(value string) error
	// Validate validates ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	Validate() error
	validateFromText() error
	validateObj(set_default bool)
	setDefault()
	// PodIndex returns int32, set in ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference.
	PodIndex() int32
	// SetPodIndex assigns int32 provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	SetPodIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	// HasPodIndex checks if PodIndex has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	HasPodIndex() bool
	// SwitchIndex returns int32, set in ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference.
	SwitchIndex() int32
	// SetSwitchIndex assigns int32 provided by user to ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	SetSwitchIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	// HasSwitchIndex checks if SwitchIndex has been set in ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference
	HasSwitchIndex() bool
}

// PodIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) PodIndex() int32 {

	return *obj.obj.PodIndex

}

// PodIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) HasPodIndex() bool {
	return obj.obj.PodIndex != nil
}

// SetPodIndex sets the int32 value in the ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) SetPodIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {

	obj.obj.PodIndex = &value
	return obj
}

// SwitchIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) SwitchIndex() int32 {

	return *obj.obj.SwitchIndex

}

// SwitchIndex returns a int32
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) HasSwitchIndex() bool {
	return obj.obj.SwitchIndex != nil
}

// SetSwitchIndex sets the int32 value in the ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference object
// description is TBD
func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) SetSwitchIndex(value int32) ChaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference {

	obj.obj.SwitchIndex = &value
	return obj
}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) validateObj(set_default bool) {
	if set_default {
		obj.setDefault()
	}

}

func (obj *chaosBackgroundTrafficFlowEntryPointSwitchReferencePodSwitchReference) setDefault() {

}
