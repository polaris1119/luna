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
	// 摘要算法用的盐（不区分不同原来）
	CommonSalt string

	// 区分不同来源的盐，只有在 CommonSalt 是空时有效
	// key 的值从 from 参数获取
	FromSalt map[string]string

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

		// 如果使用了 FromSalt，必须有 from 参数
		if s.CommonSalt == "" {
			if _, ok := args["from"]; !ok {
				return errors.New("缺少from参数")
			}
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

	if s.CommonSalt != "" {
		buffer.Append(s.CommonSalt)
	} else {
		if from, ok := args["from"]; ok {
			if salt, ok := s.FromSalt[goutils.ConvertString(from)]; ok {
				buffer.Append(salt)
			}
		}
	}

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
