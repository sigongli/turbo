package turbo

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
)

const grpcServiceName string = "grpc_service_name"
const grpcServiceAddress string = "grpc_service_address"
const thriftServiceName string = "thrift_service_name"
const thriftServiceAddress string = "thrift_service_address"
const httpPort string = "http_port"
const filterProtoJson string = "filter_proto_json"
const filterProtoJsonEmitZeroValues string = "filter_proto_json_emit_zerovalues"
const filterProtoJsonInt64AsNumber string = "filter_proto_json_int64_as_number"

var Config *config = &config{}

type config struct {
	GOPATH          string // the GOPATH used by Turbo
	RpcType         string // "grpc"/"thrift"
	ConfigFileName  string // yaml file name, exclude extention
	ServiceRootPath string // absolute path
	ServicePkgPath  string // package path, e.g. "github.com/vaporz/turbo"

	configs        map[string]string
	urlServiceMaps [][3]string
	fieldMappings  map[string][]string
}

func (c *config) GrpcServiceName() string {
	return c.configs[grpcServiceName]
}

func (c *config) SetGrpcServiceName(name string) {
	c.configs[grpcServiceName] = name
}

func (c *config) GrpcServiceAddress() string {
	return c.configs[grpcServiceAddress]
}

func (c *config) SetGrpcServiceAddress(address string) {
	c.configs[grpcServiceAddress] = address
}

func (c *config) ThriftServiceName() string {
	return c.configs[thriftServiceName]
}

func (c *config) SetThriftServiceName(name string) {
	c.configs[thriftServiceName] = name
}

func (c *config) ThriftServiceAddress() string {
	return c.configs[thriftServiceAddress]
}

func (c *config) SetThriftServiceAddress(address string) {
	c.configs[thriftServiceAddress] = address
}

func (c *config) HTTPPort() int64 {
	i, err := strconv.ParseInt(c.configs[httpPort], 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	return i
}

func (c *config) HTTPPortStr() string {
	return ":" + c.configs[httpPort]
}

func (c *config) SetHTTPPort(p int64) {
	c.configs[httpPort] = strconv.FormatInt(p, 10)
}

func (c *config) FilterProtoJson() bool {
	option, ok := c.configs[filterProtoJson]
	if !ok || option != "true" {
		return false
	}
	return true
}

func (c *config) SetFilterProtoJson(filterJson bool) {
	c.configs[filterProtoJson] = strconv.FormatBool(filterJson)
}

func (c *config) FilterProtoJsonEmitZeroValues() bool {
	option, ok := c.configs[filterProtoJson]
	if !ok || option != "true" {
		return false
	}
	option, ok = c.configs[filterProtoJsonEmitZeroValues]
	if ok && option == "false" {
		return false
	}
	return true
}

func (c *config) SetFilterProtoJsonEmitZeroValues(emitZeroValues bool) {
	c.configs[filterProtoJsonEmitZeroValues] = strconv.FormatBool(emitZeroValues)
}

func (c *config) FilterProtoJsonInt64AsNumber() bool {
	option, ok := c.configs[filterProtoJson]
	if !ok || option != "true" {
		return false
	}
	option, ok = c.configs[filterProtoJsonInt64AsNumber]
	if ok && option == "false" {
		return false
	}
	return true
}

func (c *config) SetFilterProtoJsonInt64AsNumber(asNumber bool) {
	c.configs[filterProtoJsonInt64AsNumber] = strconv.FormatBool(asNumber)
}

// LoadServiceConfigWith accepts a package path, then load service.yaml in that path
func LoadServiceConfig(rpcType, pkgPath, configFileName string) {
	initRpcType(rpcType)
	initConfigFileName(configFileName)
	initPkgPath(pkgPath)
	loadServiceConfig()
}

func initConfigFileName(name string) {
	Config.ConfigFileName = name
}

func initRpcType(r string) {
	Config.RpcType = r
}

func initPkgPath(pkgPath string) {
	goPath := os.Getenv("GOPATH")
	paths := strings.Split(goPath, ":")
	Config.GOPATH = paths[0]
	Config.ServiceRootPath = Config.GOPATH + "/src/" + pkgPath
	Config.ServicePkgPath = pkgPath
}

func loadServiceConfig() {
	// TODO reload config at runtime
	viper.SetConfigName(Config.ConfigFileName)
	viper.AddConfigPath(Config.ServiceRootPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	initUrlMap()
	initConfigs()
	initFieldMapping()
}

func initUrlMap() {
	Config.urlServiceMaps = make([][3]string, 0)
	urlMap := viper.GetStringSlice("urlmapping")
	for _, line := range urlMap {
		appendUrlServiceMap(strings.TrimSpace(line))
	}
}

func appendUrlServiceMap(line string) {
	values := strings.Split(line, " ")
	HTTPMethod := strings.TrimSpace(values[0])
	url := strings.TrimSpace(values[1])
	methodName := strings.TrimSpace(values[2])
	Config.urlServiceMaps = append(Config.urlServiceMaps, [3]string{HTTPMethod, url, methodName})
}

func initConfigs() {
	Config.configs = viper.GetStringMapString("config")
}

func initFieldMapping() {
	//TODO viper is case-insensitive, CamelCase map key is lower cased
	//Config.fieldMappings = viper.GetStringMapStringSlice(Config.RpcType + "-fieldmapping")

	conf, err := yaml.ReadFile(Config.ServiceRootPath + "/" + Config.ConfigFileName + ".yaml")
	if err != nil {
		log.Fatalf("readfile(%q): %s", Config.ServiceRootPath+"/service.yaml", err)
	}
	configFile := *conf
	node, err := yaml.Child(configFile.Root, Config.RpcType+"-fieldmapping")
	if err != nil {
		return
	}
	Config.fieldMappings = make(map[string][]string)
	fieldMappingMap := node.(yaml.Map)
	for k, v := range fieldMappingMap {
		valueStrList := make([]string, 0)
		if v != nil {
			valueList := v.(yaml.List)
			for _, line := range valueList {
				valueStrList = append(valueStrList, strings.TrimSpace(yaml.Render(line)))
			}
		}
		Config.fieldMappings[k] = valueStrList
	}
}
