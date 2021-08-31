/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"fmt"
	"github.com/go-logr/logr"
)

type Logger struct {
	log logr.Logger
}

func NewLogger(logger logr.Logger, keyandvalues ...interface{}) *Logger {
	if len(keyandvalues) > 0 {
		logger = logger.WithValues(keyandvalues...)
	}
	return &Logger{logger}
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	if l == nil {
		return
	}
	if len(args) == 0 {
		l.log.Info(msg)
	} else {
		l.log.Info(fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Errorf(err error, msg string, args ...interface{}) {
	if l == nil {
		return
	}
	if len(args) == 0 {
		l.log.Error(err, msg)
	} else {
		l.log.Error(err, fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) WithValues(keyandvalues ...interface{}) *Logger {
	return NewLogger(l.log, keyandvalues...)
}
