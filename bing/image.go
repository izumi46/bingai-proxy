package bing

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	bingImageCreateUrl = "%s/images/create?q=%s&rt=4&FORM=GENCRE"
	bingImageResult    = "%s/images/create/async/results/%s"
)

func NewImage(cookies string) *Image {
	return &Image{
		cookies:     cookies,
		BingBaseUrl: bingBaseUrl,
	}
}

func (image *Image) Image(q string) ([]string, string, error) {
	URL := fmt.Sprintf(bingImageCreateUrl, image.BingBaseUrl, url.QueryEscape(q))
	body := strings.NewReader(url.QueryEscape(fmt.Sprintf("q=%s&qs=ds", q)))
	req, err := http.NewRequest("POST", URL, body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Origin", "https://www.bing.com")
	req.Header.Set("Referer", "https://www.bing.com/images/create/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", image.cookies)
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	if res.StatusCode != 302 {
		return nil, "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}
	newURL, err := url.Parse(fmt.Sprintf("%s%s", image.BingBaseUrl, req.Header.Get("Location")))
	if err != nil {
		return nil, "", err
	}
	req.Method = "GET"
	req.URL = newURL
	res, err = client.Do(req)
	if err != nil {
		return nil, "", err
	}
	if res.StatusCode != 200 {
		return nil, "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}

	id := newURL.Query().Get("id")
	// fmt.Println(id)

	var bytes []byte
	for i := 0; i < 120; i++ {
		time.Sleep(1 * time.Second)

		req.URL, err = url.Parse(fmt.Sprintf(bingImageResult, image.BingBaseUrl, id))
		if err != nil {
			continue
		}
		res, err = client.Do(req)
		if err != nil {
			continue
		}
		defer res.Body.Close()
		bytes, err = io.ReadAll(res.Body)
		if err != nil {
			continue
		}

		if len(string(bytes)) > 1 && strings.Contains(res.Header.Get("Content-Type"), "text/html") {
			break
		}
	}

	if err != nil {
		return nil, id, err
	}
	if !(len(string(bytes)) > 1 && strings.Contains(res.Header.Get("Content-Type"), "text/html")) {
		return nil, "", fmt.Errorf("timeout")
	}

	node, err := html.Parse(strings.NewReader(string(bytes)))
	if err != nil {
		return nil, id, err
	}

	var images []string
	findImgs(node, &images)

	var results []string
	for i := range images {
		if !strings.Contains(images[i], "/rp/") {
			url, _ := url.Parse(images[i])
			url.RawQuery = ""
			results = append(results, url.String())
		}
	}

	return results, id, nil
}

func findImgs(n *html.Node, vals *[]string) {
	if n.Type == html.ElementNode && n.Data == "img" {
		for _, a := range n.Attr {
			if a.Key == "src" {
				*vals = append(*vals, a.Val)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findImgs(c, vals)
	}
}
