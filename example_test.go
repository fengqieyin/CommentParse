package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type expectResult struct {
	total  int
	inline int
	block  int
}

func getExpectResult() map[string]expectResult {
	return map[string]expectResult{
		"testing/cpp/lib_json/json_reader.cpp": {
			total:  1992,
			inline: 134,
			block:  0,
		},
		"testing/cpp/lib_json/json_tool.h": {
			total:  138,
			inline: 13,
			block:  19,
		},
		"testing/cpp/lib_json/json_value.cpp": {
			total:  1634,
			inline: 111,
			block:  18,
		},
		"testing/cpp/lib_json/json_writer.cpp": {
			total:  1259,
			inline: 89,
			block:  0,
		},
		"testing/cpp/special_cases.cpp": {
			total:  62,
			inline: 6,
			block:  34,
		},
		"testing/cpp/test_lib_json/fuzz.cpp": {
			total:  54,
			inline: 5,
			block:  0,
		},
		"testing/cpp/test_lib_json/fuzz.h": {
			total:  14,
			inline: 5,
			block:  0,
		},
		"testing/cpp/test_lib_json/jsontest.cpp": {
			total:  430,
			inline: 54,
			block:  1,
		},
		"testing/cpp/test_lib_json/jsontest.h": {
			total:  288,
			inline: 52,
			block:  8,
		},
		"testing/cpp/test_lib_json/main.cpp": {
			total:  3971,
			inline: 182,
			block:  0,
		},
	}
}

func TestSimpleAdd(t *testing.T) {
	assert.Equal(t, 2, 1+1, "wrong calculation")
}

func TestCountSingleChar(t *testing.T) {
	result := []*calculate{}
	err := filepath.Walk("testing", func(fileName string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("An error occurred while accessing the file: %s\n", err.Error())
			return err
		}

		// Check if it is a file
		if !info.Mode().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(fileName))
		if ext == ".c" || ext == ".cpp" || ext == ".h" || ext == ".hpp" {
			// Read file contents
			fileContent, err := ioutil.ReadFile(fileName)
			if err != nil {
				t.Errorf("打开文件时出错: %s\n", err.Error())
				return err
			}

			// Count the number of comment lines
			cal := countCommentLinesInContent(fileContent)
			cal.name = fileName
			result = append(result, cal)
		}
		if err != nil {
			t.Errorf("An error occurred while accessing the directory: %s\n", err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		t.Errorf("An error occurred while accessing the directory: %s\n", err.Error())
		return
	}

	expectResult := getExpectResult()

	for _, item := range result {
		expect := expectResult[item.name]
		assert.Equal(t, expect.total, item.total, item.name)
		assert.Equal(t, expect.inline, len(item.inlineSet), item.name)
		assert.Equal(t, expect.block, len(item.blockSet), item.name)
	}
}
