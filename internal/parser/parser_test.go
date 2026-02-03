package parser

import (
	"strings"
	"testing"
)

func TestParseCurl_RealWorldCases(t *testing.T) {
	tests := []struct {
		name           string
		curlCmd        string
		wantMethod     string
		wantURL        string
		wantBodySubstr string // check if body contains this substring
		wantNoHeaders  []string
	}{
		{
			name: "Chrome - Data Raw",
			curlCmd: `curl 'https://demopgsgw.jtcargo.co.id/global-hr-base/hrEmp/page' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'Authorization: bearer 9c2cda5970a348339e9855de7458c5fa' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -H 'Origin: https://demohr.jtcargo.co.id' \
  -H 'Referer: https://demohr.jtcargo.co.id/' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-site' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36' \
  -H 'country: indonesia' \
  -H 'lang: EN' \
  -H 'routeName: /emp/emp' \
  -H 'sec-ch-ua: "Not(A:Brand";v="8", "Chromium";v="144", "Google Chrome";v="144"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  --data-raw '{"data":{"birthdayEnd":"","birthdayStart":"","empCode":null,"empName":null,"empTypeList":[],"empTypeTextList":[],"entryDate":[],"entryDateEnd":"","entryDateStart":"","mobile":null,"orgName":null,"orgCode":null,"positionCode":null,"positionName":null,"region":null,"regularDateEnd":"","regularDateStart":"","resignDateEnd":"","resignDateStart":"","statusList":[],"statusTextList":[],"regularStatusList":[],"regularStatusTextList":[]},"pageIndex":1,"pageSize":50}'`,
			wantMethod:     "POST",
			wantURL:        "https://demopgsgw.jtcargo.co.id/global-hr-base/hrEmp/page",
			wantBodySubstr: "empCode",
			wantNoHeaders:  []string{"Connection", "Content-Length", "Accept-Encoding", "Host"},
		},
		{
			name: "Safari - Data Binary and Blacklisted Headers",
			curlCmd: `curl 'https://demopgsgw.jtcargo.co.id/global-hr-base/hrEmp/page' \
-X 'POST' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json, text/plain, */*' \
-H 'Authorization: bearer 9c2cda5970a348339e9855de7458c5fa' \
-H 'Sec-Fetch-Site: same-site' \
-H 'Accept-Language: zh-CN,zh-Hans;q=0.9' \
-H 'Accept-Encoding: gzip, deflate, br' \
-H 'Sec-Fetch-Mode: cors' \
-H 'Host: demopgsgw.jtcargo.co.id' \
-H 'Origin: https://demohr.jtcargo.co.id' \
-H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Safari/605.1.15' \
-H 'Referer: https://demohr.jtcargo.co.id/' \
-H 'Content-Length: 461' \
-H 'Sec-Fetch-Dest: empty' \
-H 'Connection: keep-alive' \
-H 'routeName: /emp/emp' \
-H 'lang: CN' \
-H 'country: indonesia' \
--data-binary '{"data":{"birthdayEnd":"","birthdayStart":"","empCode":null,"empName":null,"empTypeList":[],"empTypeTextList":[],"entryDate":[],"entryDateEnd":"","entryDateStart":"","mobile":null,"orgName":null,"orgCode":null,"positionCode":null,"positionName":null,"region":null,"regularDateEnd":"","regularDateStart":"","resignDateEnd":"","resignDateStart":"","statusList":[],"statusTextList":[],"regularStatusList":[],"regularStatusTextList":[]},"pageIndex":1,"pageSize":50}'`,
			wantMethod:     "POST",
			wantURL:        "https://demopgsgw.jtcargo.co.id/global-hr-base/hrEmp/page",
			wantBodySubstr: "empCode",
			wantNoHeaders:  []string{"Connection", "Content-Length", "Accept-Encoding", "Host"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurl(tt.curlCmd)
			if err != nil {
				t.Fatalf("ParseCurl() error = %v", err)
			}

			if req.Method != tt.wantMethod {
				t.Errorf("Method = %v, want %v", req.Method, tt.wantMethod)
			}

			if req.URL != tt.wantURL {
				t.Errorf("URL = %v, want %v", req.URL, tt.wantURL)
			}

			if !strings.Contains(req.Body, tt.wantBodySubstr) {
				t.Errorf("Body missing expected content. Got: %v", req.Body)
			}

			// Verify blacklisted headers are NOT present
			for _, h := range tt.wantNoHeaders {
				// Check case-insensitive, although map keys might be preserved as original case
				// The parser implementation stores keys as provided in curl, but `headerBlacklist` logic prevents adding them.
				// So we iterate over all headers in `req.Headers` and check if they match any blacklisted key.
				for k := range req.Headers {
					if strings.EqualFold(k, h) {
						t.Errorf("Blacklisted header found: %s", k)
					}
				}
			}
		})
	}
}
