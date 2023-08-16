/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package i18n

import (
	"context"
	"embed"
	"fmt"
	"io"
	"sync"

	log "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/utils/iconv"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/utils/ifile"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/utils/iregex"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/utils/iyaml"
)

// Manager for i18n contents, it is concurrent safe, supporting hot reload.
type Manager struct {
	mu      sync.RWMutex
	data    map[string]map[string]string // Translating map.
	pattern string                       // Pattern for regex parsing.
	options Options                      // configuration options.
}

// Options is used for i18n object configuration.
type Options struct {
	Path       []embed.FS // I18n files storage path.
	Language   string     // Default local language.
	Delimiters []string   // Delimiters for variable parsing.
}

var (
	// defaultDelimiters defines the default key variable delimiters.
	defaultDelimiters = []string{"{{.", "}}"}
	// language files folders.
	folders = "lang"
)

// New creates and returns a new i18n manager.
// The optional parameter `option` specifies the custom options for i18n manager.
// It uses a default one if it's not passed.
func New(options ...Options) *Manager {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	} else {
		opts = DefaultOptions()
	}
	if len(opts.Language) == 0 {
		opts.Language = defaultLanguage
	}
	if len(opts.Delimiters) == 0 {
		opts.Delimiters = defaultDelimiters
	}
	m := &Manager{
		options: opts,
		pattern: fmt.Sprintf(
			`%s(.+?)%s`,
			iregex.Quote(opts.Delimiters[0]),
			iregex.Quote(opts.Delimiters[1]),
		),
	}
	log.Infof(`New: %#v`, m)
	return m
}

// DefaultOptions creates and returns a default options for i18n manager.
func DefaultOptions() Options {
	var path []embed.FS

	return Options{
		Path:       path,
		Language:   defaultLanguage,
		Delimiters: defaultDelimiters,
	}
}

// SetPath sets the directory path storing i18n files.
func (m *Manager) SetPath(path []embed.FS) {
	m.options.Path = path
}

// SetLanguage sets the language for translator.
func (m *Manager) SetLanguage(language string) {
	m.options.Language = toLanguage(language)
	log.Infof(`SetLanguage: %s`, m.options.Language)
}

// SetDelimiters sets the delimiters for translator.
func (m *Manager) SetDelimiters(left, right string) {
	m.pattern = fmt.Sprintf(`%s(.+?)%s`, iregex.Quote(left), iregex.Quote(right))
	log.Infof(`SetDelimiters: %v`, m.pattern)
}

// T is alias of Translate for convenience.
func (m *Manager) T(ctx context.Context, content string) string {
	return m.Translate(ctx, content)
}

// Tf is alias of TranslateFormat for convenience.
func (m *Manager) Tf(ctx context.Context, format string, values ...interface{}) string {
	return m.TranslateFormat(ctx, format, values...)
}

// TranslateFormat translates, formats and returns the `format` with configured language
// and given `values`.
func (m *Manager) TranslateFormat(ctx context.Context, format string, values ...interface{}) string {
	return fmt.Sprintf(m.Translate(ctx, format), values...)
}

// Translate translates `content` with configured language.
func (m *Manager) Translate(ctx context.Context, content string) string {
	m.init(ctx)
	m.mu.RLock()
	defer m.mu.RUnlock()
	transLang := m.options.Language
	if lang := LanguageFromCtx(ctx); lang != "" {
		transLang = lang
	}

	// fallback to original content if translation doesn't exist in specified locale
	data := m.data[transLang]
	if data == nil {
		return content
	}
	// Parse content as name.
	if v, ok := data[content]; ok {
		return v
	}
	// Parse content as variables container.
	result, _ := iregex.ReplaceStringFuncMatch(
		m.pattern, content,
		func(match []string) string {
			if v, ok := data[match[1]]; ok {
				return v
			}
			// return match[1] will return the content between delimiters
			// return match[0] will return the original content
			return match[0]
		})
	log.Infof(`Translate for language: %s`, transLang)
	return result
}

// GetContent retrieves and returns the configured content for given key and specified language.
// It returns an empty string if not found.
func (m *Manager) GetContent(ctx context.Context, key string) string {
	m.init(ctx)
	m.mu.RLock()
	defer m.mu.RUnlock()
	transLang := m.options.Language
	if lang := LanguageFromCtx(ctx); lang != "" {
		transLang = lang
	}
	if data, ok := m.data[transLang]; ok {
		return data[key]
	}
	return ""
}

// init initializes the manager for lazy initialization design.
// The i18n manager is only initialized once.
func (m *Manager) init(ctx context.Context) {
	m.mu.RLock()
	// If the data is not nil, means it's already initialized.
	if m.data != nil {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]map[string]string)
	for _, item := range m.options.Path {
		// 读取语言文件夹所有文件
		readDir, err := item.ReadDir(folders)
		if err != nil {
			log.Error(ctx, "Failed to read all files in the %s directory %s", folders, err.Error())
			continue
		}
		for _, entry := range readDir {
			// 获取对应语言
			lang := ifile.Name(entry.Name())
			if m.data[lang] == nil {
				m.data[lang] = make(map[string]string)
			}
			if !entry.IsDir() {
				content, _ := m.getBytes(item, folders+"/"+entry.Name())
				if c, err := iyaml.Decode(content); err == nil {
					for k, v := range c {
						m.data[lang][k] = iconv.String(v)
					}
				} else {
					log.Error(ctx, "load i18n file '%s' failed: %+v", entry.Name(), err.Error())
				}
			}

		}
	}
}

func (m *Manager) getBytes(file embed.FS, filePath string) ([]byte, error) {
	f, err := file.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
	}
}
