package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go test -v -run ^TestGojsCall_BasicExecution$ todoplus/api/internal/utils

func TestGojsCall_File(t *testing.T) {

	// 从外部文件读取JavaScript代码
	scriptContent, err := ioutil.ReadFile("/mnt/12t/code/src.go/coolpeople.com.cn/api/rag.js")
	if err != nil {
		fmt.Printf("读取脚本文件错误: %v\n", err)
		os.Exit(1)
	}

	query := "风机维修"
	stringContent := strings.ReplaceAll(string(scriptContent), "{{query}}", query)
	stringContent = strings.ReplaceAll(string(stringContent), "{{keys}}", "AAA,BBB")

	result := GojsCall(stringContent)

	jsonContent, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("json marshal错误: %v\n", err)
		os.Exit(1)
	}
	t.Logf("jsonContent: %s", string(jsonContent))
	// t.Logf("result[jsonfileinfo]: %+v", result["jsonfileinfo"])
	t.Logf("result[jsonfileinfo]: %+v", reflect.TypeOf(result["jsonfileinfo"]))

	// []interface {},逐个转为字符[]string
	jsonfileinfo := result["jsonfileinfo"].([]interface{})
	var jsonfileinfoList []string
	for _, v := range jsonfileinfo {
		jsonfileinfoList = append(jsonfileinfoList, v.(string))
	}
	for _, v := range jsonfileinfoList {
		t.Logf(">>>%s<<<", v)
	}

	t.Logf("result[error]: %+v", result["error"])

	//	t.Logf("result: %+v", result)

}
func TestGojsCall_BasicExecution(t *testing.T) {
	// Test basic JavaScript execution
	script := `
		var result = {
			success: true,
			message: "Hello from JavaScript",
			value: 42
		};
		result;
	`

	result := GojsCall(script)
	t.Logf("result: %+v", result)
	t.Logf("result['value']: %+v", result["value"])
	typeOf := reflect.TypeOf(result["value"])
	t.Logf("typeOf: %+v", typeOf)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "Hello from JavaScript", result["message"])
	assert.Equal(t, int64(42), result["value"])

}

func TestGojsCall_ConsoleLog(t *testing.T) {
	// Test console.log functionality
	script := `
		console.log("Test message 1", "Test message 2");
		var result = { success: true, logged: true };
		result;
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, true, result["logged"])
}

func TestGojsCall_ConsoleError(t *testing.T) {
	// Test console.error functionality
	script := `
		console.error("Error message 1", "Error message 2");
		var result = { success: true, errorLogged: true };
		result;
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, true, result["errorLogged"])
}

func TestGojsCall_HttpGet(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from test server",
			"status":  "ok",
		})
	}))
	defer server.Close()

	script := `
		var response = httpGet("` + server.URL + `");
		var result = {
			success: true,
			response: response,
			length: response.length
		};
		result;
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Contains(t, result["response"], "Hello from test server")
	assert.Greater(t, result["length"], float64(0))
}

func TestGojsCall_HttpPost(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "POST request received",
			"received": requestBody,
			"status":   "ok",
		})
	}))
	defer server.Close()

	script := `
		var postData = JSON.stringify({ test: "data", number: 123 });
		var response = httpPost("` + server.URL + `", postData);
		var result = {
			success: true,
			response: response,
			length: response.length
		};
		result;
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Contains(t, result["response"], "POST request received")
	assert.Contains(t, result["response"], "test")
	assert.Contains(t, result["response"], "data")
	assert.Greater(t, result["length"], float64(0))
}

func TestGojsCall_ComplexReturn(t *testing.T) {
	// Test complex return object with nested structures
	script := `
		var result = {
			success: true,
			data: {
				users: [
					{ id: 1, name: "Alice", age: 25 },
					{ id: 2, name: "Bob", age: 30 }
				],
				total: 2,
				metadata: {
					version: "1.0.0",
					timestamp: Date.now()
				}
			},
			message: "Data retrieved successfully"
		};
		result;
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "Data retrieved successfully", result["message"])

	// Check nested data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)

	users, ok := data["users"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, users, 2)

	// Check first user
	user1, ok := users[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), user1["id"])
	assert.Equal(t, "Alice", user1["name"])
	assert.Equal(t, float64(25), user1["age"])
}

func TestGojsCall_ErrorHandling(t *testing.T) {
	// Test JavaScript error handling
	script := `
		try {
			var result = {
				success: true,
				message: "No error occurred"
			};
			result;
		} catch (error) {
			var result = {
				success: false,
				error: error.message
			};
			result;
		}
	`

	result := GojsCall(script)

	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "No error occurred", result["message"])
}

func TestGojsCall_EmptyScript(t *testing.T) {
	// Test with empty script
	script := ""

	result := GojsCall(script)
	t.Logf("result: %+v", result)

	// Should return nil or empty result for empty script
	assert.NotNil(t, result)
}

func TestGojsCall_UndefinedReturn(t *testing.T) {
	// Test script that doesn't return anything
	script := `
		var x = 10;
		var y = 20;
		// No return statement
	`

	result := GojsCall(script)

	// Should handle undefined return gracefully
	assert.NotNil(t, result)
}

func TestGojsCall_StringReturn(t *testing.T) {
	// Test script that returns a string
	script := `
		"Hello World";
	`

	result := GojsCall(script)

	// Should handle string return
	assert.NotNil(t, result)
}

func TestGojsCall_NumberReturn(t *testing.T) {
	// Test script that returns a number
	script := `
		42;
	`

	result := GojsCall(script)

	// Should handle number return
	assert.NotNil(t, result)
}

func TestGojsCall_BooleanReturn(t *testing.T) {
	// Test script that returns a boolean
	script := `
		true;
	`

	result := GojsCall(script)

	// Should handle boolean return
	assert.NotNil(t, result)
}

func TestGojsCall_ArrayReturn(t *testing.T) {
	// Test script that returns an array
	script := `
		[1, 2, 3, "hello", true];
	`

	result := GojsCall(script)

	// Should handle array return
	assert.NotNil(t, result)
}

func TestGojsCall_HttpError(t *testing.T) {
	// Test HTTP error handling with invalid URL
	script := `
		try {
			var response = httpGet("http://invalid-url-that-does-not-exist.com");
			var result = {
				success: true,
				response: response
			};
			result;
		} catch (error) {
			var result = {
				success: false,
				error: error.message
			};
			result;
		}
	`

	result := GojsCall(script)

	// Should handle HTTP errors gracefully
	assert.NotNil(t, result)
}

func TestGojsCall_JavaScriptError(t *testing.T) {
	// Test JavaScript syntax error
	script := `
		var result = {
			success: true,
			message: "This should work"
		};
		// Intentional syntax error
		var x = ;
		result;
	`

	// This should cause a JavaScript error and exit
	// We'll test that the function handles errors gracefully
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic due to os.Exit(1) in the original function
			t.Logf("Expected panic occurred: %v", r)
		}
	}()

	result := GojsCall(script)

	// If we reach here, the error was handled gracefully
	assert.NotNil(t, result)
}

func TestGojsCall_ConcurrentExecution(t *testing.T) {
	// Test concurrent execution of multiple scripts
	script := `
		var result = {
			success: true,
			threadId: Math.random(),
			message: "Concurrent execution test"
		};
		result;
	`

	// Run multiple goroutines
	results := make(chan map[string]interface{}, 5)

	for i := 0; i < 5; i++ {
		go func() {
			result := GojsCall(script)
			results <- result
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		result := <-results
		assert.NotNil(t, result)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, "Concurrent execution test", result["message"])
	}
}

// Benchmark test for performance
func BenchmarkGojsCall(b *testing.B) {
	script := `
		var result = {
			success: true,
			message: "Benchmark test",
			value: Math.random() * 1000
		};
		result;
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GojsCall(script)
	}
}
