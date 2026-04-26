/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var client = &http.Client{}

func GetRemoteAddr(r *http.Request) string {
	remoteAddr := r.Header.Get("X-Forwarded-For")
	if remoteAddr == "" {
		remoteAddr = r.Header.Get("X-Real-IP")
	}
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}
	return remoteAddr
}

func Download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, err
}

func PerformHTTPRequest(req *http.Request, retryCounts ...int) (*http.Response, error) {
	// 设置重试次数
	retryCount := 3
	if len(retryCounts) > 0 && retryCounts[0] > 0 {
		retryCount = retryCounts[0]
	}
	var err error
	var resp *http.Response
	var lastStatusCode int
	for i := range retryCount {
		resp, err = client.Do(req)
		if err != nil {
			// 请求失败，继续重试
			if i < retryCount-1 {
				time.Sleep(300 * time.Millisecond)
			}
			continue
		}
		if resp.StatusCode == http.StatusOK {
			// 请求成功，返回响应
			return resp, nil
		}
		lastStatusCode = resp.StatusCode
		resp.Body.Close()
		// 如果不是最后一次重试，等待一段时间后重试
		if i < retryCount-1 {
			time.Sleep(300 * time.Millisecond)
		}
	}

	// 所有重试都失败，返回带有 HTTP 状态码的错误消息
	if err != nil {
		return nil, fmt.Errorf("failed after %d attempts. Last error: %v", retryCount, err)
	}

	return nil, fmt.Errorf("failed after %d attempts. (HTTP Status Code: %d)", retryCount, lastStatusCode)
}
