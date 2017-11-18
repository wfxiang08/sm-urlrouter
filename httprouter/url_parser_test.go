package httprouter

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/wfxiang08/cyutils/utils"
	log "github.com/wfxiang08/cyutils/utils/rolling_log"
	"os"
	"strings"
	"testing"
)

//
// go test url_manager/httprouter -v -run "TestUrlRouterParser$"
//
func TestUrlRouterParser(t *testing.T) {

	router, err := ParseUrlConfig("url.txt")
	if err != nil {
		assert.NoError(t, err, "Parse url.txt failed")
	}

	if router == nil {
		fmt.Printf("Router is nil\n")
		return
	}

	testcases := []struct {
		Method         string
		Url            string
		ExpectedAction string
	}{
		{
			Method:         "GET",
			Url:            "/api/v17/android/sm/zh-CN/phone/xxhdpi/config",
			ExpectedAction: "app/config",
		},
		{
			Method:         "GET",
			Url:            "/api/v17/android/sm/bn-IN/phone/xhdpi/tagging/user_list?recording_id=4785074249695578",
			ExpectedAction: "friend/tagging-user-list",
		},
		{
			Method:         "PATCH",
			Url:            "/api/v17/android/sm/en-GB/phone/xhdpi/users/5348024334183745/notification/40000002",
			ExpectedAction: "notification/read",
		},
		{
			Method:         "GET",
			Url:            "/api/v17/android/sm/zh-CN/phone/xxhdpi/songs/most-sung",
			ExpectedAction: "song/most-sung",
		},
		{
			Method:         "POST",
			Url:            "/api/v17/android/sm/zh-CN/phone/xxhdpi/multipart-uploading",
			ExpectedAction: "recording/complete",
		},
	}
	for _, testcase := range testcases {
		fmt.Printf(utils.Cyan("----------------------\n"))

		action, params := router.ResolvePath(testcase.Url, testcase.Method)

		if action != nil {
			log.Printf("Action: %s, Params: %v", action.String(), params)
			if action.Action != testcase.ExpectedAction {
				log.Printf(utils.Red("Error: %s vs. E: %s\n"), action.Action, testcase.ExpectedAction)
			}
		} else {
			log.Printf("Action not found for url: %s", testcase.Url)
		}
	}

}

//
// go test url_manager/httprouter -v -run "TestUrlRouterResolve$"
//
func TestUrlRouterResolve(t *testing.T) {

	router, err := ParseUrlConfig("url.txt")
	if err != nil {
		assert.NoError(t, err, "Parse url.txt failed")
	}

	// 读取测试文件
	filename := "router.txt"
	file, err := os.Open(filename)
	if err != nil {
		assert.NoError(t, err, "Open router.txt failed")
	}
	defer file.Close()

	var reader *bufio.Reader
	if strings.HasSuffix(filename, ".gz") {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			log.ErrorErrorf(err, "gzip file open failed: %s", filename)
		}
		reader = bufio.NewReader(gzipReader)
	} else {
		reader = bufio.NewReader(file)
	}
	totalLine := 0
	okLine := 0
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line := strings.Trim(line, "\"' \n")
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "---->")
		if len(fields) != 3 {
			continue
		}

		totalLine++
		method, url, action := fields[0], fields[1], fields[2]

		pattern, _ := router.ResolvePath(url, method)
		if pattern != nil {
			if action != pattern.Action {
				log.Printf(utils.Magenta("Action resolve failed for: %s --> %s, Exp: %s vs. %s"), method, url, action, pattern.Action)
				assert.Equal(t, action, pattern.Action, "Action not the same")
			} else {
				// 正常的就不打印结果了
				okLine++
			}
		} else {
			log.Printf(utils.Red("Url resolve faield for: %s --> %s"), method, url)
		}
	}

	log.Printf("Total line: %d, Ok Line: %d", totalLine, okLine)

}
