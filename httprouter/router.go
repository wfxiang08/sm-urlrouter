package httprouter

import (
	"fmt"
	"github.com/wfxiang08/cyutils/utils/errors"
	"strings"
)

type Router struct {
	trees          map[string]*TreeNode
	action2Pattern map[string]*UrlPattern
	RouterHash     string
}

func (this *Router) CompileRoute() error {
	for _, pattern := range this.action2Pattern {
		err := pattern.Compile()
		if err != nil {
			return err
		}
	}

	for _, root := range this.trees {
		err := root.CompileTree()
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Router) ToRoute(action string, params map[string]string) (string, error) {
	if pattern, ok := this.action2Pattern[action]; ok {
		return pattern.GetInterpolatedUrl(params)
	} else {
		return "", errors.New(fmt.Sprintf("no action found for %s", action))
	}
}

//
// 添加Url Action
//
func (this *Router) AddAction(action *UrlPattern) error {
	if this.trees == nil {
		this.trees = make(map[string]*TreeNode)
		this.action2Pattern = make(map[string]*UrlPattern)
	}

	// 直接记录Action
	this.action2Pattern[action.Action] = action

	segments := SplitUrlSegments(action.Url)
	if len(segments) == 0 {
		return errors.New("invalid url segments")
	}

	// 创建Tree Node
	for _, method := range action.Methods {
		root := this.trees[method]
		if root == nil {
			// Root node
			root = &TreeNode{
				Segment:  "ROOT",
				Compiled: true,
			}
			this.trees[method] = root
		}
		err := root.AddRoute(segments, action)
		if err != nil {
			return err
		}
	}
	return nil
}

// 给定Path, Method返回解析之后的UrlPattern, 以及对应的Params
func (this *Router) ResolvePath(path string, method string) (*UrlPattern, Params) {

	// 去掉path后面的"?"
	index := strings.Index(path, "?")
	if index != -1 {
		path = path[0:index]
	}

	// 大小写标准化
	method = strings.ToUpper(method)

	var params Params
	if root, ok := this.trees[method]; ok {
		segments := SplitUrlSegments(path)
		return root.MatchTree(segments, params)
	} else {
		return nil, nil
	}
}
