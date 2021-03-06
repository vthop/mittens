//Copyright 2019 Expedia, Inc.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package http

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttp_FlagToHttpRequest(t *testing.T) {

	requestFlag := `post:/db:{"db": "true"}`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, request.Method)
	assert.Equal(t, "/db", request.Path)
	assert.Equal(t, `{"db": "true"}`, *request.Body)
}

func TestHttp_FlagWithoutBodyToHttpRequest(t *testing.T) {

	requestFlag := `get:ping`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "ping", request.Path)
	assert.Nil(t, request.Body)
}

func TestHttp_RandomNumbersInterpolation(t *testing.T) {

	requestFlag := `post:/path_{numbers-5}:{"body": "{numbers-10}"}`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, request.Method)

	var numbersRegex = regexp.MustCompile("\\d+")
	matchPath := numbersRegex.MatchString(request.Path)
	matchBody := numbersRegex.MatchString(*request.Body)

	assert.True(t, matchPath)
	assert.True(t, matchBody)
	assert.Equal(t, len(request.Path), 11)  //  "path_ + 5 numbers"
	assert.Equal(t, len(*request.Body), 22) // { "body": 10 numbers }
}

func TestHttp_RandomLetersInterpolation(t *testing.T) {

	requestFlag := `post:/path_{chars-5}:{"body": "{chars-10}"}`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, request.Method)

	var numbersRegex = regexp.MustCompile("[a-zA-Z0-9]+")
	matchPath := numbersRegex.MatchString(request.Path)
	matchBody := numbersRegex.MatchString(*request.Body)

	assert.True(t, matchPath)
	assert.True(t, matchBody)
	assert.Equal(t, len(request.Path), 11)  //  "path_ + 5 chars"
	assert.Equal(t, len(*request.Body), 22) // { "body": 10 chars }
}

func TestHttp_TodayInterpolation(t *testing.T) {

	requestFlag := `post:/db_{today+5}:{"db": "{today+5}"}`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, request.Method)
	date := time.Now().Add(time.Duration(5) * 24 * time.Hour).Format("2006-01-02") // today + 5
	assert.Equal(t, "/db_"+date, request.Path)
	assert.Equal(t, fmt.Sprintf(`{"db": "%s"}`, date), *request.Body)
}

func TestHttp_TomorrowInterpolation(t *testing.T) {

	requestFlag := `post:/db_{tomorrow}:{"db": "{tomorrow}"}`
	request, err := ToHttpRequest(requestFlag)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, request.Method)
	date := time.Now().Add(time.Duration(1) * 24 * time.Hour).Format("2006-01-02") // today + 1
	assert.Equal(t, "/db_"+date, request.Path)
	assert.Equal(t, fmt.Sprintf(`{"db": "%s"}`, date), *request.Body)
}

func TestHttp_FlagWithInvalidMethodToHttpRequest(t *testing.T) {

	requestFlag := `hmm:/ping:all=true`
	_, err := ToHttpRequest(requestFlag)
	require.Error(t, err)
}
