package main

import (
	"fmt"
	"github.com/Dreamacro/clash/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"regexp"
)

func main() {
	inputFile := pflag.StringP("config", "c", "config.yaml", "clash config file (yaml type)")
	templateFile := pflag.StringP("template", "t", "template.yaml", "template config file")

	pflag.Parse()
	logrus.Debugf("clash config file: %v", *inputFile)
	logrus.Debugf("clash template file: %v", *templateFile)

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
		fmt.Printf("id: %v proxies: %v", idx, templateCfg.ProxyGroup[idx]["proxies"])
		templateCfg.ProxyGroup[idx]["proxies"] = allName
	}

	bs, err := yaml.Marshal(templateCfg)
	if err != nil {
		logrus.Fatalf("marshal template config error: %v", err)
	}
	ioutil.WriteFile("out.yaml", bs, 0600)
}

func preprocessConfigFile(filename string) (*config.RawConfig, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	bs = RemoveEmoji(bs)
	return config.UnmarshalRawConfig(bs)
}
func RemoveEmoji(bs []byte) []byte {
	var emojiRe = regexp.MustCompile(`[\x{F000}-\x{FFFFF}]`)
	return emojiRe.ReplaceAll(bs, nil)
}
