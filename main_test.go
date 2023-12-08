package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var cafeList = map[string][]string{
	"moscow": []string{"Мир кофе", "Сладкоежка", "Кофе и завтраки", "Сытый студент"},
}

func mainHandle(w http.ResponseWriter, req *http.Request) {
	countStr := req.URL.Query().Get("count")
	if countStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("count missing"))
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong count value"))
		return
	}

	city := req.URL.Query().Get("city")

	cafe, ok := cafeList[city]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong city value"))
		return
	}

	if count > len(cafe) {
		count = len(cafe)
	}

	answer := strings.Join(cafe[:count], ",")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(answer))
}

func serverGetHandler(t *testing.T, r string) (int, string) {
	req := httptest.NewRequest("GET", r, nil)
	responseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(mainHandle)
	handler.ServeHTTP(responseRecorder, req)

	resp := responseRecorder.Result()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, body, "Response can't be empty")

	return resp.StatusCode, string(body)
}

func TestMainHandlerWhenCountMoreThanTotal(t *testing.T) {
	totalCount := 4
	status, body := serverGetHandler(t, "/cafe?count=5&city=moscow")

	assert.Equal(t, http.StatusOK, status, "expected OK Status")

	cafes := strings.Split(body, ",")
	assert.Equal(t, len(cafes), totalCount, "Result count eq to total count")
	assert.Equal(t, cafes, cafeList["moscow"])
}

func TestMainHandlerWhenCountEqualToTwo(t *testing.T) {
	reqCount := 2
	status, body := serverGetHandler(t, "/cafe?count="+strconv.Itoa(reqCount)+"&city=moscow")

	assert.Equal(t, http.StatusOK, status, "expected 200 Status")

	cafes := strings.Split(body, ",")
	assert.Equal(t, len(cafes), reqCount, "Result count eq to req count")
	assert.Equal(t, cafes, cafeList["moscow"][:reqCount])
}

func TestMainHandlerWhenUsingUnknCity(t *testing.T) {
	status, body := serverGetHandler(t, "/cafe?count=1&city=spb")

	assert.Equal(t, http.StatusBadRequest, status, "expected BadRequest Status")
	assert.Equal(t, string(body), "wrong city value")
}

func TestMainHandlerWhenUsingEmptyCount(t *testing.T) {
	status, body := serverGetHandler(t, "/cafe?count=&city=moscow")

	assert.Equal(t, http.StatusBadRequest, status, "expected BadRequest Status")
	assert.Equal(t, string(body), "count missing")
}

func TestMainHandlerWhenUsingBadCount(t *testing.T) {
	status, body := serverGetHandler(t, "/cafe?count=all&city=moscow")

	assert.Equal(t, http.StatusBadRequest, status, "expected BadRequest Status")
	assert.Equal(t, string(body), "wrong count value")
}
