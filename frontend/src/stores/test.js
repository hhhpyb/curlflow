
var curl = 'curl ^"https://demovipgw.jtcargo.co.id/vip/workOrder/save^" ^\n' +
    '  -H ^"Accept: application/json, text/plain, */*^" ^\n' +
    '  -H ^"Accept-Language: zh-CN,zh;q=0.9^" ^\n' +
    '  -H ^"Cache-Control: max-age=2, must-revalidate^" ^\n' +
    '  -H ^"Connection: keep-alive^" ^\n' +
    '  -H ^"Content-Type: application/json;charset=UTF-8^" ^\n' +
    '  -H ^"Origin: http://localhost:8090^" ^\n' +
    '  -H ^"Referer: http://localhost:8090/^" ^\n' +
    '  -H ^"Sec-Fetch-Dest: empty^" ^\n' +
    '  -H ^"Sec-Fetch-Mode: cors^" ^\n' +
    '  -H ^"Sec-Fetch-Site: cross-site^" ^\n' +
    '  -H ^"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36^" ^\n' +
    '  -H ^"authToken: eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiJ7XCJjcmVhdGVUaW1lXCI6XCIyMDI0LTExLTI2VDE1OjU4OjU5XCIsXCJjdXN0b21lckNvZGVcIjpcIkowMDg2MDI0NTU1XCIsXCJjdXN0b21lck5hbWVcIjpcInlhbmppZVwiLFwiZnJlaWdodFN0YXR1c1wiOjEsXCJpZFwiOjE2NDMzNDgyMjQsXCJpbnN1cmFuY2VGbGFnXCI6MSxcImlzQWRtaW5cIjowLFwiaXNQcml2YWN5XCI6MCxcIm5lZWREaXNwYXRjaFwiOjIsXCJuZXR3b3JrSWRcIjoxNjE4LFwicGFyZW50SWRcIjowLFwicXVvdGVJZFwiOjQ2LFwic3RhZmZDb2RlXCI6XCJcIixcInN0YXR1c1wiOjEsXCJzdXJmYWNlRm9ybWF0XCI6MSxcInRhcmdldFByaW50ZXJcIjpcIkhQUlQgU0w0MlwiLFwidXBkYXRlVGltZVwiOlwiMjAyNC0xMS0yNlQxNjowMDo1N1wifSIsImlhdCI6MTc3MDI3Mjk1OH0.gZdzQ6DuBZCeUuXsR42USYkLdVDXmkYZEdPa_LyofCU^" ^\n' +
    '  -H ^"language: CN^" ^\n' +
    '  -H ^"routeName: workOrder^" ^\n' +
    '  -H ^"sec-ch-ua: ^\\^"Not(A:Brand^\\^";v=^\\^"8^\\^", ^\\^"Chromium^\\^";v=^\\^"144^\\^", ^\\^"Google Chrome^\\^";v=^\\^"144^\\^"^" ^\n' +
    '  -H ^"sec-ch-ua-mobile: ?0^" ^\n' +
    '  -H ^"sec-ch-ua-platform: ^\\^"Windows^\\^"^" ^\n' +
    '  --data-raw ^"^{^\\^"waybillNo^\\^":^\\^"3333333333333^\\^",^\\^"customerPhone^\\^":^\\^"3333333333^\\^",^\\^"customerName^\\^":^\\^"3333333333^\\^",^\\^"customerSex^\\^":^\\^"2^\\^",^\\^"customerType^\\^":^\\^"1^\\^",^\\^"firstTypeCode^\\^":^\\^"GD5^\\^",^\\^"secondTypeCode^\\^":^\\^"511^\\^",^\\^"problemDescription^\\^":^\\^"3333333333333^\\^"^}^"'

function preprocessCurl() {
    if (!curl) return "";

    // 1. First, handle the line continuation character '^' at the very end of lines
    // This merges multi-line commands into a single line.
    let cleaned = curl.replace(/\^\s*[\r\n]+/g, " ");

    // 2. Handle Windows CMD special escape sequences
    // We use sequential replacements to handle nested escapes correctly.

    // Case A: Handle nested escaped quotes: ^\^" -> \" (often found in JSON keys/values)
    cleaned = cleaned.replace(/\^\\\^"/g, '\\"');

    // Case B: Handle standard escaped quotes: ^" -> "
    cleaned = cleaned.replace(/\^"/g, '"');

    // Case C: Handle escaped braces for JSON: ^{ -> { and ^} -> }
    cleaned = cleaned.replace(/\^{/g, '{').replace(/\^}/g, '}');

    // Case D: Handle escaped carets: ^^ -> ^
    cleaned = cleaned.replace(/\^\^/g, '^');

    // 3. Finally, strip any remaining carets that are used for escaping other characters
    // (e.g., ^&, ^|, etc. in CMD). In curl commands, these are almost always shell escapes.
    cleaned = cleaned.replace(/\^/g, '');

    return cleaned;
}

console.log(preprocessCurl(curl));

