// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package httprouter

import (
	"fmt"
	"github.com/wfxiang08/cyutils/utils"
	"github.com/wfxiang08/cyutils/utils/errors"
	log "github.com/wfxiang08/cyutils/utils/rolling_log"
	"regexp"
	"sort"
)

// Node的定义
// Path             Handle
//  \                *<1>
//  |:params with_pattern
//  ├s               nil
//  |├earch\         *<2>
//  |└upport\        *<3>
//  ├blog\           *<4>
//  |    └:post      nil
//  |         └\     *<5>
//  ├about-us\       *<6>
//  |        └team\  *<7>
//  └contact\        *<8>
// 在树桩结构中，优先考虑:params, 也就是:params一旦匹配，其他就不考虑？
//
// /users/<stage_name:.*>
// /users/stagename/<stage_name:.*>
//
type TreeNode struct {
	Segment         string // 例如: users <stage_name:.*> 等
	SegReg          *regexp.Regexp
	ParamName       string
	Action          *UrlPattern
	Children        []*TreeNode
	Compiled        bool
	CompiledSegment string
}

func (this *TreeNode) String() string {
	return fmt.Sprintf("Segment: %s, Compiled: %v", this.Segment, this.CompiledSegment)
}
func (this *TreeNode) CompileTree() error {
	// 编译正则
	if !this.Compiled {
		url, reg, err1 := PreprocessRegexPattern(this.Segment)

		//fmt.Printf("Compile tree: %s --> %s, %v. %v. children num: %d\n", this.Segment, url, reg, err1,
		//	len(this.Children))

		if err1 == nil {
			// 正常
			if reg {
				this.CompiledSegment = url
				urlRegex, err := regexp.Compile(url)
				if err != nil {
					return err
				}
				this.SegReg = urlRegex
				if len(this.SegReg.SubexpNames()) >= 2 {
					this.ParamName = this.SegReg.SubexpNames()[1]
				} else {
					return errors.New("Named Regex not exist")
				}
			}
		} else {
			// 异常，直接返回错误
			return err1
		}
		this.Compiled = true
	}

	// 递归处理children
	for i := 0; i < len(this.Children); i++ {
		err := this.Children[i].CompileTree()
		if err != nil {
			return err
		}
	}

	// 让正则表达式排在后面，优先处理字符串完全匹配
	if len(this.Children) > 1 {
		sort.Sort(treeNodeList(this.Children))
	}

	return nil
}

// segment如何和当前的节点做Match呢？
func (this *TreeNode) MatchSegment(segment string) (*Param, bool) {
	//if DebugMode {
	//	log.Debugf("Pattern Match: %s --> %s\n", segment, this.String())
	//}
	if this.SegReg != nil {
		match := this.SegReg.FindStringSubmatch(segment)
		if len(match) >= 2 {
			//if DebugMode {
			//	log.Debugf("Match Segment: %s --> %s\n", this.ParamName, match[1])
			//}
			param := &Param{
				Key:   this.ParamName,
				Value: match[1],
			}
			return param, true
		} else {
			//if DebugMode {
			//	log.Debugf("Invalid match for %s --> %s\n", segment, this.Segment)
			//}
			return nil, false
		}
	}
	return nil, this.Segment == segment
}

// 在Add Route阶段做匹配，只匹配原始的Segment
func (this *TreeNode) MatchPatternSegment(segment string) bool {
	// log.Debugf("Pattern Match: %s vs. %s\n", this.Segment, segment)
	return this.Segment == segment
}

//
// 只对自己的Children做匹配
//
func (this *TreeNode) AddRoute(segments []string, action *UrlPattern) error {
	// 证明n.Segment是已经Match完毕了
	// log.Debugf("Processing segments: %s\n", strings.Join(segments, "/"))
	if len(segments) > 0 {
		segment := segments[0]
		for i := 0; i < len(this.Children); i++ {
			if this.Children[i].MatchPatternSegment(segment) {
				return this.Children[i].AddRoute(segments[1:], action)
			}
		}
		// 没有匹配到合适的Pattern
		node := &TreeNode{
			Segment:  segment,
			Children: nil,
		}
		// log.Debugf("Add child %s/%s, @%s", utils.Green(this.Segment), utils.Magenta(segment), action.Url)
		this.Children = append(this.Children, node)

		if len(segments) == 1 {
			// 匹配完毕，直接给node添加action
			node.Action = action
			return nil
		} else {
			return node.AddRoute(segments[1:], action)
		}

	} else {
		if this.Action != nil {
			errorMsg := fmt.Sprintf("Duplicate patterns found: %s vs. %s", action.GetUrl(), this.Action.GetUrl())
			log.Printf("WARNING: " + utils.Red(errorMsg))
			if !this.Action.EndsWithSlash {
				this.Action.EndsWithSlash = action.EndsWithSlash
			}
		} else {
			this.Action = action
		}
		return nil
	}
}

//
// 1. 线程安全
// 2. MatchTree 不匹配当前的TreeNode, 只递归匹配自己的Chiren
// 3.
func (this *TreeNode) MatchTree(segments []string, params Params) (*UrlPattern, Params) {

	var action *UrlPattern
	currrentLen := len(params)

	// log.Debugf("Try: %s for: %s\n", this.Segment, strings.Join(segments, "/"))

	if len(segments) > 0 {
		segment := segments[0]
		for i := 0; i < len(this.Children); i++ {
			if param, ok := this.Children[i].MatchSegment(segment); ok {
				if param != nil {
					params = append(params, *param)
				}
				action, params = this.Children[i].MatchTree(segments[1:], params)
				// 如果匹配成功，则立马返回
				if action != nil {
					return action, params
				}

				// 准备下一个的匹配
				params = params[0:currrentLen]
			}
		}
		return nil, params // 匹配失败
	} else {
		return this.Action, params // 匹配当前的节点
	}
}

type treeNodeList []*TreeNode

func (p treeNodeList) Len() int {
	return len(p)
}
func (p treeNodeList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// 让非正则表达式排在前面
func (p treeNodeList) Less(i, j int) bool {
	return p[i].SegReg == nil && p[j].SegReg != nil
}
