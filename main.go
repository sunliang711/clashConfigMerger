package main

import (
	"fmt"
	"github.com/Dreamacro/clash/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"regexp"
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
			if k == "name" {
				allName = append(allName, v.(string))
				break __INNER
			}
		}
	}
	logrus.Debugf("allName: %v", allName)
	// TODO 把所有节点信息添加到模板中templateCfg
	allIdx := []int{}
	for idx, proxy := range templateCfg.ProxyGroup {
		for k, v := range proxy {
			if k == "name" {
				if v == "Auto" || v == "Select" {
					allIdx = append(allIdx, idx)
				}
			}
		}
	}
	logrus.Debugf("allIdx: %v", allIdx)
	for _, idx := range allIdx {
		templateCfg.ProxyGroup[idx]["proxies"] = allName
	}

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
