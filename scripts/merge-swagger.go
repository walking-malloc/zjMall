package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Swagger struct {
	Swagger      string                 `json:"swagger"`
	Info         map[string]interface{} `json:"info"`
	Tags         []interface{}          `json:"tags"`
	Consumes     []string               `json:"consumes"`
	Produces     []string               `json:"produces"`
	Paths        map[string]interface{} `json:"paths"`
	Definitions  map[string]interface{} `json:"definitions"`
	Security     []interface{}          `json:"security,omitempty"`
	SecurityDefs map[string]interface{} `json:"securityDefinitions,omitempty"`
}

func main() {
	outputDir := "docs/openapi"
	outputFile := filepath.Join(outputDir, "api.swagger.json")

	// 读取所有 swagger 文件
	files := []string{
		filepath.Join(outputDir, "api/proto/common/health.swagger.json"),
		filepath.Join(outputDir, "api/proto/user/user.swagger.json"),
	}

	var merged Swagger
	merged.Swagger = "2.0"
	merged.Info = map[string]interface{}{
		"title":   "zjMall API",
		"version": "1.0.0",
		"description": "zjMall 电商平台 API 文档",
	}
	merged.Tags = []interface{}{}
	merged.Consumes = []string{"application/json"}
	merged.Produces = []string{"application/json"}
	merged.Paths = make(map[string]interface{})
	merged.Definitions = make(map[string]interface{})

	// 合并所有文件
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				continue // 文件不存在，跳过
			}
			fmt.Printf("Error reading %s: %v\n", file, err)
			continue
		}

		var swagger Swagger
		if err := json.Unmarshal(data, &swagger); err != nil {
			fmt.Printf("Error parsing %s: %v\n", file, err)
			continue
		}

		// 合并 tags
		merged.Tags = append(merged.Tags, swagger.Tags...)

		// 合并 paths
		for path, methods := range swagger.Paths {
			merged.Paths[path] = methods
		}

		// 合并 definitions
		for def, schema := range swagger.Definitions {
			merged.Definitions[def] = schema
		}
	}

	// 写入合并后的文件
	outputData, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(outputFile, outputData, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully merged swagger files into %s\n", outputFile)
}
