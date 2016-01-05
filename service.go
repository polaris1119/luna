// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package luna

import (
	"errors"
	"sort"
	"time"

	"github.com/polaris1119/goutils"
	"github.com/twinj/uuid"
)

func CheckAuth(args map[string]interface{}) error {
	if DefaultService.CheckAuth != nil {
		return DefaultService.CheckAuth(args)
	}
	return DefaultService.checkAuth(args)
}

type Service struct {
	// 摘要算法用的盐
	Salt string

	// 权限校验函数
	CheckAuth func(map[string]interface{}) error
}

var DefaultService = new(Service)

func (s *Service) checkAuth(args map[string]interface{}) error {
	if sign, ok := args["sign"]; !ok {
		return errors.New("没有传递签名信息")
	} else {

		delete(args, "sign")

		// TODO：timestamp 校验
		if _, ok := args["timestamp"]; !ok {
			return errors.New("缺少timestamp")
		}

		// TODO：nonce 校验
		if _, ok := args["nonce"]; !ok {
			return errors.New("缺少nonce")
		}

		newSign := s.GenSign(args)
		if sign != newSign {
			return errors.New("签名不合法")
		}
	}

	return nil
}

func (s *Service) GenSign(args map[string]interface{}) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))

	buffer := goutils.NewBuffer()
	for _, k := range keys {
		buffer.Append(k).Append("=").Append(goutils.ConvertString(args[k]))
	}
	buffer.Append(s.Salt)

	return goutils.Md5(buffer.String())
}

func FillRequireArgs(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		args = make(map[string]interface{})
	}
	if _, ok := args["timestamp"]; !ok {
		args["timestamp"] = time.Now().Unix()
	}

	if _, ok := args["nonce"]; !ok {
		args["nonce"] = uuid.NewV4().String()
	}

	args["sign"] = DefaultService.GenSign(args)

	return args
}
