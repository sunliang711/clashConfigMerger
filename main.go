package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"clashConfigMerger/utils"
	"github.com/Dreamacro/clash/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	GitHash   string
	BuildTime string
)

func main() {
	inputFile := pflag.StringP("config", "c", "config.yaml", "clash config file (yaml type)")
	templateFile := pflag.StringP("template", "t", "template.yaml", "template config file")
	version := pflag.BoolP("version", "v", false, "version info")
	outputFile := pflag.StringP("output", "o", "output.yaml", "output config file")
	logLevel := pflag.StringP("loglevel", "l", "info", "log level")
	groupPrefixes := pflag.StringArrayP("prefix", "p", nil, "group prefixes")

	pflag.Parse()

	if *version {
		fmt.Fprintf(os.Stderr, `version info:
git hash: %v
build time: %v`, GitHash, BuildTime)
		return
	}

	setLogLevel(*logLevel)

	logrus.Infof("clash config file: %v", *inputFile)
	logrus.Infof("clash template file: %v", *templateFile)

	templateCfg, err := preprocessConfigFile(*templateFile)
	if err != nil {
		logrus.Fatalf("preprocess template file error: %v", err)
	}

	inputCfg, err := preprocessConfigFile(*inputFile)
	if err != nil {
		logrus.Fatalf("preprocess input file error: %v", err)
	}

	// 输入配置文件的节点添加到模板文件中
	templateCfg.Proxy = inputCfg.Proxy

	allName := []string{}
	for _, proxy := range inputCfg.Proxy {
		// get all name
	__INNER:
		for k, v := range proxy {
			// 找到name对应的v，把它放到allName中，用来统计所有的proxy name
			if k == "name" {
				allName = append(allName, v.(string))
				break __INNER
			}
		}
	}
	logrus.Debugf("all proxy names: %v", allName)
	// 把allName替换name叫做All的proxy-group中
	// 把name为All的proxy-group替换成

	// 1. 构造name为All的proxy-group (用map表示),并插入到templateCfg.ProxyGroup中
	allGroup := make(map[string]interface{})
	allGroup["name"] = "All"
	allGroup["type"] = "select"
	// 插入到第一个位置

	// 2. 把其它在模板中的proxy-group(比如name为Apple,Emby,Telegram)的proxies中追加allName中的所有name
	for idx := range templateCfg.ProxyGroup {
		templateCfg.ProxyGroup[idx]["proxies"] = append(templateCfg.ProxyGroup[idx]["proxies"].([]interface{}), []interface{}{allName}...)
	}

	var allLcp []string
	// 3. 把allName中的name按照最长前缀分组，每个组就是一个proxy group
	var customGroups []map[string]interface{}
	groups := utils.GroupStringArray(allName, *groupPrefixes)
	for idx := range groups {
		newProxyGroup := make(map[string]interface{})
		lcp := utils.LongestCommonPrefix(groups[idx])
		lcp = strings.Trim(lcp, " ")
		allLcp = append(allLcp, lcp)
		newProxyGroup["name"] = lcp
		newProxyGroup["type"] = "url-test"
		newProxyGroup["url"] = "http://www.gstatic.com/generate_204"
		newProxyGroup["interval"] = 300
		newProxyGroup["tolerance"] = 100
		newProxyGroup["proxies"] = groups[idx]
		// 插入
		customGroups = append(customGroups, newProxyGroup)
	}
	allGroup["proxies"] = allLcp

	tmpGroups := append([]map[string]interface{}{allGroup}, customGroups...)
	templateCfg.ProxyGroup = append(tmpGroups, templateCfg.ProxyGroup...)

	bs, err := yaml.Marshal(templateCfg)
	if err != nil {
		logrus.Fatalf("marshal template config error: %v", err)
	}
	logrus.Infof("output to file: %v", *outputFile)
	ioutil.WriteFile(*outputFile, bs, 0600)
}

func preprocessConfigFile(filename string) (*config.RawConfig, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	bs = removeEmoji(bs)
	return config.UnmarshalRawConfig(bs)
}
func removeEmoji(bs []byte) []byte {
	var emojiRe = regexp.MustCompile(`[\x{F000}-\x{FFFFF}]`)
	return emojiRe.ReplaceAll(bs, nil)
}

func setLogLevel(lvl string) {
	switch lvl {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
