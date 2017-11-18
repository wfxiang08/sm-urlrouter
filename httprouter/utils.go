package httprouter

import (
	"fmt"
	"github.com/wfxiang08/cyutils/utils/errors"
	log "github.com/wfxiang08/cyutils/utils/rolling_log"
	"net/url"
	"strings"
)

var (
	ValidHttpMethods map[string]bool
)

func init() {
	ValidHttpMethods = map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
		"DELETE":  true,
	}
}

// 是否为有效的 http method
func IsHttpMethodValid(method string) bool {
	_, ok := ValidHttpMethods[method]
	return ok
}

type Segment struct {
	Segment string
	IsKey   bool
}

type UrlPattern struct {
	Url           string
	EndsWithSlash bool
	Action        string
	Methods       []string // 全部为大写, 且有效
	Segments      []Segment
}

func (this *UrlPattern) Compile() error {
	segments, err := PreprocessUrlInterpolation(this.Url)
	if err != nil {
		return err
	}
	this.Segments = segments
	return nil
}

// 然后的Path必须以"/"开头
func (this *UrlPattern) GetInterpolatedUrl(params map[string]string) (string, error) {
	results := make([]string, len(this.Segments))
	for i, segment := range this.Segments {
		if segment.IsKey {
			if v, ok := params[segment.Segment]; ok {
				results[i] = url.PathEscape(v)
				delete(params, segment.Segment)
			} else {
				return "", errors.New(fmt.Sprintf("param not specified: %s", segment.Segment))
			}
		} else {
			results[i] = segment.Segment
		}
	}

	// 处理path部分
	requestUri := "/" + strings.Join(results, "/")
	if this.EndsWithSlash {
		requestUri = requestUri + "/"
	}

	// 处理Query部分
	results = results[0:0]
	anchor, hasAnchor := params["#"]
	if hasAnchor {
		delete(params, "#")
	}
	if len(params) > 0 {
		for key, value := range params {
			results = append(results, key+"="+url.QueryEscape(value))
		}
	}
	if len(results) > 0 {
		requestUri = requestUri + "?" + strings.Join(results, "&")
	}

	// 处理Anchor
	if hasAnchor {
		requestUri = requestUri + "#" + anchor
	}
	return requestUri, nil
}

func (this *UrlPattern) String() string {
	if this.EndsWithSlash {
		return fmt.Sprintf("Url: %s/ --> %s, Action: %s", this.Url, this.Action, strings.Join(this.Methods, ","))
	} else {
		return fmt.Sprintf("Url: %s --> %s, Action: %s", this.Url, this.Action, strings.Join(this.Methods, ","))
	}
}

func (this *UrlPattern) GetUrl() string {
	if this.EndsWithSlash {
		return this.Url + "/"
	} else {
		return this.Url
	}
}

func PreprocessPhpUrl(url string) (*UrlPattern, error) {

	fields := strings.SplitN(url, " ", 2)

	pattern := &UrlPattern{}

	// 解析Methods
	if len(fields) > 1 {
		// 统一为大写
		methods := strings.Split(strings.ToUpper(fields[0]), ",")
		for _, method := range methods {
			// 逐项检查
			if IsHttpMethodValid(method) {
				pattern.Methods = append(pattern.Methods, method)
			} else {
				log.Errorf("invalid method %d for url: %s", method, url)
			}
		}
	} else {
		pattern.Methods = kDefaultMethods
	}

	urlAction := fields[len(fields)-1]
	fields = strings.SplitN(urlAction, kActionSeperator, 2)
	if len(fields) != 2 {
		return nil, errors.New(fmt.Sprintf("No action found in url: %s", url))
	}

	pattern.Action = strings.TrimSpace(fields[1])
	pattern.Url = strings.Trim(fields[0], "? ")

	urlLen := len(pattern.Url)
	if urlLen > 0 {
		if pattern.Url[urlLen-1] == '/' {
			pattern.Url = pattern.Url[0 : urlLen-1]
			pattern.EndsWithSlash = true
		}
	}
	return pattern, nil

}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// 统计path中的参数的个数
func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	if n >= 255 {
		return 255
	}
	return uint8(n)
}

func SplitUrlSegments(url string) []string {
	if len(url) > 0 && url[0] == '/' {
		url = url[1:]
	}
	if len(url) > 0 && url[len(url)-1] == '/' {
		url = url[0 : len(url)-1]
	}
	if len(url) == 0 {
		return nil
	}
	fields := strings.Split(url, "/")
	// fmt.Printf("Process url: %s\n", strings.Join(fields, "/"))
	return fields
}

//Input:  `<deviceDensity>`,
//Output: `(?P<deviceDensity>[^/]*)`,
//
// 最多只支持一个Pattern
// return: GoUrlPattern, isRegex, error
//
func PreprocessRegexPattern(pattern string) (string, bool, error) {
	// 可能是Root Node的Pattern, 长度为0
	if len(pattern) <= 1 {
		log.Printf("WARNING: pattern short: %s", pattern)
	}

	start := len(pattern) > 0 && pattern[0] == '<'
	end := len(pattern) > 1 && pattern[len(pattern)-1] == '>'

	// 如果<>配对
	if start && end {
		pattern = pattern[1 : len(pattern)-1]
		index := strings.Index(pattern, ":")
		if index != -1 {
			name := pattern[0:index]
			regPattern := pattern[index+1:]
			if regPattern == ".*" {
				regPattern = `[^/]+`
			}
			return fmt.Sprintf(`(?P<%s>%s)`, name, regPattern), true, nil
		} else {
			return fmt.Sprintf(`(?P<%s>[^/]+)`, pattern), true, nil
		}

	} else if !start && !end {
		// 没有pattern的情况
		return pattern, false, nil
	}

	// <>不配对
	return "", false, errors.New("Invalid pattern")
}

func ParseRegexPattern(pattern string) (string, bool, error) {
	// 可能是Root Node的Pattern, 长度为0
	if len(pattern) <= 1 {
		log.Printf("WARNING: pattern short: %s", pattern)
	}

	start := len(pattern) > 0 && pattern[0] == '<'
	end := len(pattern) > 1 && pattern[len(pattern)-1] == '>'

	// 如果<>配对
	if start && end {
		pattern = pattern[1 : len(pattern)-1]
		index := strings.Index(pattern, ":")
		if index != -1 {
			name := pattern[0:index]
			return name, true, nil
		} else {
			return pattern, true, nil
		}

	} else if !start && !end {
		// 没有pattern的情况
		return pattern, false, nil
	}

	// <>不配对
	return "", false, errors.New("Invalid pattern")
}

func PreprocessUrlInterpolation(url string) ([]Segment, error) {
	segments := SplitUrlSegments(url)

	results := make([]Segment, len(segments))
	for i, segment := range segments {
		name, isReg, err := ParseRegexPattern(segment)
		if err != nil {
			return nil, err
		}
		results[i] = Segment{
			Segment: name,
			IsKey:   isReg,
		}
	}
	return results, nil
}
