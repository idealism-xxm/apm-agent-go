// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// +build go1.11

package apmgoredisv8

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-redis/redis/v8"

	"go.elastic.co/apm"
)

// goRedisApmHook is an implementation of redis.Hook that reports cmds as spans to Elastic APM.
type goRedisApmHook struct{}

// NewGoRedisApmHook returns a redis.Hook that reports cmds as spans to Elastic APM.
func NewGoRedisApmHook() redis.Hook {
	return &goRedisApmHook{}
}

// BeforeProcess initiates the span for the redis cmd
func (r *goRedisApmHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, ctx := apm.StartSpan(ctx, getCmdName(cmd), "db.redis")
	return withSpan(ctx, span), nil
}

// AfterProcess ends the initiated span from BeforeProcess
func (r *goRedisApmHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if span := spanFromContext(ctx); span != nil {
		span.End()
	}
	return nil
}

// BeforeProcessPipeline initiates the span for the redis cmds
func (r *goRedisApmHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// Join all cmd names with ", ".
	var cmdNameBuf bytes.Buffer
	for i, cmd := range cmds {
		if i != 0 {
			cmdNameBuf.WriteString(", ")
		}
		cmdNameBuf.WriteString(getCmdName(cmd))
	}

	span, ctx := apm.StartSpan(ctx, cmdNameBuf.String(), "db.redis")
	return withSpan(ctx, span), nil
}

// AfterProcess ends the initiated span from BeforeProcessPipeline
func (r *goRedisApmHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if span := spanFromContext(ctx); span != nil {
		span.End()
	}
	return nil
}

func getCmdName(cmd redis.Cmder) string {
	cmdName := strings.ToUpper(cmd.Name())
	if cmdName == "" {
		cmdName = "(empty command)"
	}
	return cmdName
}

type goRedisApmSpanKey struct{}

var key = goRedisApmSpanKey{}

func withSpan(ctx context.Context, span *apm.Span) context.Context {
	return context.WithValue(ctx, key, span)
}

func spanFromContext(ctx context.Context) *apm.Span {
	span, _ := ctx.Value(key).(*apm.Span)
	return span
}
