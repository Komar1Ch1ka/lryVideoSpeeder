package h5p

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	cookie     string
	sesskey    string
	baseURL    string
}

func NewClient(cookie, sesskey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 50,
				MaxConnsPerHost:     50,
			},
		},
		cookie:  cookie,
		sesskey: sesskey,
		baseURL: "https://moodle.scnu.edu.cn",
	}
}

func (c *Client) newRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Cookie", c.cookie)

	return req, nil
}

func (c *Client) GetPage(pageURL string) (string, error) {
	req, err := c.newRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *Client) SendProgress(cmid string, total int, progress float64, time int, finish int) (*ProgressResponse, error) {
	data := []ProgressRequest{
		{
			Index:      0,
			Methodname: "report_h5pstats_set_time",
			Args: ProgressArgs{
				Time:     time,
				Finish:   finish,
				Cmid:     cmid,
				Total:    total,
				Progress: progress,
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/lib/ajax/service.php?sesskey=%s", c.baseURL, c.sesskey)

	req, err := c.newRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []ProgressResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result) > 0 {
		return &result[0], nil
	}

	return nil, fmt.Errorf("empty response")
}

func (c *Client) GetVideoDuration(pageURL string) (int, error) {
	html, err := c.GetPage(pageURL)
	if err != nil {
		return 0, err
	}

	re := regexp.MustCompile(`"duration"\s*:\s*(\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		var duration float64
		fmt.Sscanf(matches[1], "%f", &duration)
		return int(duration), nil
	}

	re2 := regexp.MustCompile(`"maxScore"\s*:\s*\d+[^}]*"duration"\s*:\s*(\d+(?:\.\d+)?)`)
	matches2 := re2.FindStringSubmatch(html)
	if len(matches2) > 1 {
		var duration float64
		fmt.Sscanf(matches2[1], "%f", &duration)
		return int(duration), nil
	}

	return 600, nil
}

func ExtractCmidFromURL(courseURL string) (string, error) {
	u, err := url.Parse(courseURL)
	if err != nil {
		return "", err
	}

	id := u.Query().Get("id")
	if id == "" {
		return "", fmt.Errorf("cannot find id parameter in URL")
	}

	return id, nil
}
