package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// 用假接口验证请求体形状与响应解析（不访问真实服务）。
func TestCertificateLookupParsesResult(t *testing.T) {
	var gotBody map[string]map[string]string
	var gotPath, gotCT string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 10000,
			"msg": "成功",
			"res": {
				"prcList": [{
					"name": "张三",
					"idCarNumber": "110101199001011234",
					"sex": "男",
					"type": "B",
					"certificateNo": "B1100123456",
					"passDate": 1714521600000,
					"issueDate": "2024-05-01 00:00:00",
					"addr": "江苏省工业和信息化厅"
				}],
				"count": 1
			}
		}`))
	}))
	defer server.Close()

	svc := NewCertificateService()
	svc.apiBase = server.URL + "/app/"

	info, err := svc.Lookup("张三", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/app/businessSupport/cracOperationCert/getOperCertByParamWeb" {
		t.Fatalf("请求路径不对: %s", gotPath)
	}
	if !strings.Contains(gotCT, "application/json") {
		t.Fatalf("Content-Type 不对: %s", gotCT)
	}
	if gotBody["req"]["name"] != "张三" || gotBody["req"]["idCarNumber"] != "110101199001011234" {
		t.Fatalf("请求体不对: %+v", gotBody)
	}
	if gotBody["req"]["page_no"] != "1" || gotBody["req"]["page_size"] != "10" {
		t.Fatalf("分页参数缺失: %+v", gotBody)
	}
	if !info.Found || info.CertificateNo != "B1100123456" || info.Category != "B" {
		t.Fatalf("解析结果不对: %+v", info)
	}
	if info.IssueDate != "2024-05-01" {
		t.Fatalf("发证日期解析不对: %q", info.IssueDate)
	}
	if info.Province != "江苏" || info.Issuer != "江苏省工业和信息化厅" {
		t.Fatalf("核发机构解析不对: %+v", info)
	}
}

func TestCertificateLookupNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":10010,"res":{"prcList":[]},"msg":"未查询到对应数据"}`))
	}))
	defer server.Close()

	svc := NewCertificateService()
	svc.apiBase = server.URL

	info, err := svc.Lookup("张三", "110101199001011234")
	if err != nil {
		t.Fatal(err)
	}
	if info.Found {
		t.Fatalf("空结果不应标记为找到: %+v", info)
	}
	if info.Message == "" {
		t.Fatal("空结果应带上提示信息")
	}
}

func TestCertificateLookupValidatesInput(t *testing.T) {
	svc := NewCertificateService()
	if _, err := svc.Lookup("", "110101199001011234"); err == nil {
		t.Fatal("缺少姓名应报错")
	}
	if _, err := svc.Lookup("张三", "  "); err == nil {
		t.Fatal("缺少证件号应报错")
	}
}

func TestProvinceFromIssuer(t *testing.T) {
	cases := map[string]string{
		"江苏省工业和信息化厅":    "江苏",
		"北京市经济和信息化局":    "北京",
		"广东省工业和信息化厅":    "广东",
		"内蒙古自治区工业和信息化厅": "内蒙古",
		"":    "",
		"xx厅": "xx厅",
	}
	for in, want := range cases {
		if got := provinceFromIssuer(in); got != want {
			t.Errorf("provinceFromIssuer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCertDate(t *testing.T) {
	cases := map[string]string{
		"1714521600000":       "2024-05-01",
		"1714521600":          "2024-05-01",
		"2024-05-01 00:00:00": "2024-05-01",
		"2024-05-01":          "2024-05-01",
		"":                    "",
		"0":                   "",
	}
	for in, want := range cases {
		if got := certDate(in); got != want {
			t.Errorf("certDate(%q) = %q, want %q", in, got, want)
		}
	}
}

// 只有显式设置 CRAC_TEST_CERT=1 时才真的访问官方接口，用假数据验证请求格式能被接受。
func TestCertificateLookupLive(t *testing.T) {
	if os.Getenv("CRAC_TEST_CERT") != "1" {
		t.Skip("set CRAC_TEST_CERT=1 to hit the real CRAC endpoint")
	}
	svc := NewCertificateService()
	info, err := svc.Lookup("测试姓名", "000000000000000000")
	if err != nil {
		t.Fatalf("接口调用失败（请求格式可能不对）：%v", err)
	}
	t.Logf("found=%v message=%s", info.Found, info.Message)
}
