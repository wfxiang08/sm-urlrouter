package httprouter

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

//
// go test url_manager/httprouter -v -run "TestUrlPatternInterpolate$"
//
func TestUrlPatternInterpolate(t *testing.T) {
	line := `POST api/v17/<platform>/<client>/<language>/<deviceType>/<deviceDensity>/users/<user_id:\d+>/followees/<label:(facebook|contacts|user_ids)>###follow/multi-follow`
	action, err := PreprocessPhpUrl(line)
	assert.NoError(t, err)

	err = action.Compile()
	assert.NoError(t, err)

	params := map[string]string{
		"platform":      "ios",
		"client":        "sm",
		"language":      "zh-CN",
		"deviceType":    "ios",
		"deviceDensity": "3x",
		"user_id":       "12121212",
		"label":         "facebook",
	}
	url, err1 := action.GetInterpolatedUrl(params)
	assert.NoError(t, err1)

	fmt.Printf("Url: %s", url)
}

//
// go test url_manager/httprouter -v -run "TestUrlPatternProcess$"
//
func TestUrlPatternProcess(t *testing.T) {

	testCases := []struct {
		Input  string
		Output string
		Test   string
		Param  Param
	}{
		{
			Input:  `<deviceDensity>`,
			Output: `(?P<deviceDensity>[^/]*)`,
			Test:   "xx",
		},
		{
			Input:  `<user_id:\d+>`,
			Output: `(?P<user_id>\d+)`,
			Test:   "1212",
		},
		{
			Input:  `<stage_name:.*>`,
			Output: `(?P<stage_name>[^/]*)`,
			Test:   "hello",
		},
		{
			Input:  `<label:(hot|new|collab)>`,
			Output: `(?P<label>(hot|new|collab))`,
			Test:   "hello",
		},
		{
			Input:  `<label:(hot|new|collab)>`,
			Output: `(?P<label>(hot|new|collab))`,
			Test:   "collab",
		},
	}
	for _, testCase := range testCases {
		output, _, err := PreprocessRegexPattern(testCase.Input)
		assert.NoError(t, err)
		fmt.Printf("Output: %s\n", output)

		reg, _ := regexp.Compile(output)

		match := reg.FindStringSubmatch(testCase.Test)
		if match != nil {
			names := reg.SubexpNames()
			for i, name := range names {
				if i != 0 {
					fmt.Printf("%s --> %s\n", name, match[i])
				}
			}
		}
		// assert.Equal(t, testCase.Output, output)
	}

}
