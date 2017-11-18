package httprouter

import (
	"bufio"
	"github.com/wfxiang08/cyutils/utils"
	log "github.com/wfxiang08/cyutils/utils/rolling_log"
	"io"
	"os"
	"strings"
)

const (
	kActionSeperator = "###"
)

var (
	kDefaultMethods = []string{"GET", "POST"}
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

func (this Param) String() string {
	return this.Key + "-->" + this.Value
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

func (this Params) String() string {
	results := make([]string, len(this))
	for i, p := range this {
		results[i] = p.String()
	}
	return strings.Join(results, ", ")
}

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

func ParseUrlConfig(urlConfig string) (*Router, error) {
	file, err := os.Open(urlConfig)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	router := &Router{}

	reader := bufio.NewReader(file)
	var line string
	var action *UrlPattern
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line := strings.Trim(line, "\"' \n")
		if strings.HasPrefix(line, "#") {
			continue
		}

		action, err = PreprocessPhpUrl(line)
		if err != nil {
			log.ErrorErrorf(err, "Preprocess url failed")
			continue
		}

		//fmt.Printf(utils.Green("----------------------\n"))
		err = router.AddAction(action)
		if err != nil {
			log.ErrorErrorf(err, "Add action failed")
			return nil, err
		}
	}

	if err != nil && err != io.EOF {
		log.ErrorErrorf(err, "Error found where parse url config")
		return nil, err
	} else {
		err = router.CompileRoute()
		if err != nil {
			log.ErrorErrorf(err, "Compile router failed")
			return nil, err
		}

		router.RouterHash = utils.GetFileMD5(urlConfig)

		return router, nil
	}
}
