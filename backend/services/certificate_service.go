package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/utils"

	"crac_exam_v3/backend/config"
)

const (
	// cracAPIBase CRAC 操作证查询系统的接口根地址（页面里 baseUrl 同源）。
	cracAPIBase = "https://zhipu.allspectrum.cn:9528/CRAC/app/"
	// cracQueryPage 官方查询页，供用户手动核对。
	cracQueryPage = "https://zhipu.allspectrum.cn:9528/CRAC/crac/pages/list_cert.html"
	// cracLookupPath 按姓名 + 证件号查询证书信息的接口。
	cracLookupPath = "businessSupport/cracOperationCert/getOperCertByParamWeb"

	cracCodeOK      = 10000
	cracCodeNoData  = 10010
	cracCodeNeedArg = 10001
)

// CertificateService 通过 CRAC 官方查询接口取回本人在册的操作证书信息。
// 仅在用户点击时按需查询，不做任何后台轮询，也不落库。
type CertificateService struct {
	client  *http.Client
	apiBase string // 测试可覆盖
}

func NewCertificateService() *CertificateService {
	return &CertificateService{
		client:  &http.Client{Timeout: 20 * time.Second},
		apiBase: cracAPIBase,
	}
}

func (s *CertificateService) ServiceName() string { return "CertificateService" }

// OpenQueryPage 用系统浏览器打开官方查询页。
func (s *CertificateService) OpenQueryPage() error {
	if err := openExternalURL(cracQueryPage); err != nil {
		return fmt.Errorf("打开浏览器失败：%w", err)
	}
	return nil
}

// Lookup 按姓名 + 身份证号到 CRAC 官方接口查询证书信息。
func (s *CertificateService) Lookup(name, idCard string) (*models.CertificateInfo, error) {
	name = strings.TrimSpace(name)
	idCard = strings.TrimSpace(idCard)
	if name == "" {
		return nil, fmt.Errorf("请先填写姓名")
	}
	if idCard == "" {
		return nil, fmt.Errorf("请先填写证件号码")
	}

	payload := map[string]any{
		"req": map[string]string{
			"name":        name,
			"idCarNumber": idCard,
			"page_no":     "1",
			"page_size":   "10",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	base := s.apiBase
	if base == "" {
		base = cracAPIBase
	}
	endpoint := strings.TrimRight(base, "/") + "/" + cracLookupPath
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "crac-exam-v3/"+config.AppVersion)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("访问 CRAC 查询接口失败：%w", err)
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if readErr != nil {
		return nil, fmt.Errorf("读取查询结果失败：%w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CRAC 查询接口返回 HTTP %d", resp.StatusCode)
	}

	var parsed struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Res  struct {
			PrcList []map[string]any `json:"prcList"`
		} `json:"res"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("解析查询结果失败：%w", err)
	}

	switch parsed.Code {
	case cracCodeOK:
		// 继续
	case cracCodeNoData:
		return &models.CertificateInfo{
			Found:   false,
			Name:    name,
			IdCard:  idCard,
			Message: "官方库中未查到该姓名 + 证件号的记录，请核对后重试",
		}, nil
	case cracCodeNeedArg:
		return nil, fmt.Errorf("%s", nonEmptyString(parsed.Msg, "姓名或证件号不允许为空"))
	default:
		return nil, fmt.Errorf("%s", nonEmptyString(parsed.Msg, "CRAC 查询接口返回异常"))
	}

	if len(parsed.Res.PrcList) == 0 {
		return &models.CertificateInfo{
			Found:   false,
			Name:    name,
			IdCard:  idCard,
			Message: "未查询到对应数据",
		}, nil
	}

	item := parsed.Res.PrcList[0]
	info := &models.CertificateInfo{
		Found:         true,
		Name:          nonEmptyString(jsonText(item["name"]), name),
		IdCard:        nonEmptyString(jsonText(item["idCarNumber"]), idCard),
		Category:      normalizeCertCategory(jsonText(item["type"])),
		CertificateNo: jsonText(item["certificateNo"]),
		IssueDate:     certDate(jsonText(item["issueDate"])),
		Issuer:        jsonText(item["addr"]),
		Message:       "已从 CRAC 官方库读取到证书信息",
	}
	info.Province = provinceFromIssuer(info.Issuer)
	if info.CertificateNo == "" && info.IssueDate == "" {
		info.Message = "查到了记录，但证书编号与发证日期为空，请到官网核对"
	}
	utils.Info("CertificateService", "查询操作证信息",
		"found", info.Found, "category", info.Category, "has_no", info.CertificateNo != "")
	return info, nil
}

func nonEmptyString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

// jsonText 把接口返回的任意值转成字符串（可能是字符串，也可能是数字）。
func jsonText(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

// normalizeCertCategory 把接口返回的类别归一到 A/B/C。
func normalizeCertCategory(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	for _, c := range []string{"C", "B", "A"} {
		if strings.Contains(v, c) {
			return c
		}
	}
	return ""
}

// certDate 兼容毫秒/秒时间戳与日期字符串，统一成 2006-01-02。
func certDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return ""
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
		switch {
		case n > 1e12: // 毫秒
			return time.UnixMilli(n).Format("2006-01-02")
		case n > 1e9: // 秒
			return time.Unix(n, 0).Format("2006-01-02")
		}
	}
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}

// provinceFromIssuer 从“江苏省工业和信息化厅”里取出“江苏”，用于证书上的核发机构栏。
func provinceFromIssuer(issuer string) string {
	issuer = strings.TrimSpace(issuer)
	if issuer == "" {
		return ""
	}
	runes := []rune(issuer)
	for _, mark := range []string{"省", "市", "自治区"} {
		if i := strings.Index(issuer, mark); i > 0 {
			head := []rune(issuer[:i])
			if len(head) >= 2 && len(head) <= 6 {
				return string(head)
			}
		}
	}
	if len(runes) > 6 {
		return string(runes[:6])
	}
	return issuer
}
