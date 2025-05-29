package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, getMoscowLenFor100()}, // Should do one more request if moscow's cafe list bigger than 100
	}

	requestLines := make([]string, len(requests))

	for i := 0; i < len(requests); i++ {
		requestLines[i] = fmt.Sprintf("/cafe?city=moscow&count=%d", requests[i].count)
	}

	for i := 0; i < len(requests); i++ {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", requestLines[i], nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, requests[i].want, getLenOfResponse(response))
	}
}

// Because moscow's cafe list can become bigger than 100
func getMoscowLenFor100() int {
	if len(cafeList["moscow"]) > 100 {
		return 100
	}
	return len(cafeList["moscow"])
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	requestLines := make([]string, len(requests))

	for i := 0; i < len(requests); i++ {
		requestLines[i] = fmt.Sprintf("/cafe?city=moscow&search=%s", requests[i].search)
	}

	for i := 0; i < len(requests); i++ {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", requestLines[i], nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, requests[i].wantCount, getLenOfResponse(response))
	}
}

func getLenOfResponse(response *httptest.ResponseRecorder) int {
	if response.Body.String() == "" {
		return 0
	}

	return len(strings.Split(response.Body.String(), ","))
}
