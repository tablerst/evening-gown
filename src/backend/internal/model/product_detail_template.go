package model

import (
	"encoding/json"
	"errors"
	"strings"
)

var errDetailNotObject = errors.New("detail must be a JSON object")

const SettingKeyProductDetailTemplate = "product_detail_template"

// DefaultProductDetailTemplate returns the default template used to populate Product.DetailJSON.
// It includes the requested default fields: 颜色、尺码、件数、交付时间.
func DefaultProductDetailTemplate() json.RawMessage {
	// Keep keys compatible with existing frontend rendering:
	// - specs: [{k|label, v|value}]
	// - option_groups: [{name, options: []}]
	// NOTE: values are intentionally empty so merchandisers can fill them per product.
	v := map[string]any{
		"specs": []any{
			map[string]any{"k": "件数", "v": ""},
			map[string]any{"k": "交付时间", "v": ""},
		},
		"option_groups": []any{
			map[string]any{"name": "颜色", "options": []any{}},
			map[string]any{"name": "尺码", "options": []any{}},
		},
	}
	b, _ := json.Marshal(v)
	return b
}

// MergeProductDetailWithTemplate merges the given detail object with the template.
// Rules:
// - If detail is empty, return template.
// - If detail already contains a spec key (k/label/key/name), keep it.
// - If detail already contains an option group name (name/title/label), keep it.
// - Missing template items are appended.
func MergeProductDetailWithTemplate(template, detail json.RawMessage) (json.RawMessage, error) {
	if len(detail) == 0 {
		if len(template) == 0 {
			return DefaultProductDetailTemplate(), nil
		}
		return template, nil
	}

	tmplObj, err := asObject(template)
	if err != nil {
		// If template is broken, fallback to default.
		tmplObj, _ = asObject(DefaultProductDetailTemplate())
	}
	detailObj, err := asObject(detail)
	if err != nil {
		return nil, err
	}

	out := map[string]any{}
	for k, v := range detailObj {
		out[k] = v
	}

	// specs
	tmplSpecs := normalizeSpecs(tmplObj["specs"])
	userSpecs := normalizeSpecs(detailObj["specs"])
	if len(tmplSpecs) > 0 || len(userSpecs) > 0 {
		merged := mergeByKey(tmplSpecs, userSpecs)
		out["specs"] = merged
	}

	// option_groups
	tmplGroups := normalizeOptionGroups(tmplObj["option_groups"])
	userGroups := normalizeOptionGroups(detailObj["option_groups"])
	if len(tmplGroups) > 0 || len(userGroups) > 0 {
		merged := mergeByName(tmplGroups, userGroups)
		out["option_groups"] = merged
	}

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func asObject(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, errDetailNotObject
	}
	return obj, nil
}

func normalizeSpecs(raw any) []map[string]any {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		k := pickString(m, "k", "label", "key", "name")
		if k == "" {
			continue
		}
		out = append(out, m)
	}
	return out
}

func normalizeOptionGroups(raw any) []map[string]any {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		name := pickString(m, "name", "title", "label")
		if name == "" {
			continue
		}
		out = append(out, m)
	}
	return out
}

func pickString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s != "" {
			return s
		}
	}
	return ""
}

func mergeByKey(template, user []map[string]any) []any {
	seen := map[string]bool{}
	result := make([]any, 0, len(template)+len(user))

	// First keep user order.
	for _, it := range user {
		k := pickString(it, "k", "label", "key", "name")
		if k == "" {
			continue
		}
		seen[k] = true
		result = append(result, it)
	}
	// Append missing template items in template order.
	for _, it := range template {
		k := pickString(it, "k", "label", "key", "name")
		if k == "" {
			continue
		}
		if seen[k] {
			continue
		}
		result = append(result, it)
	}
	return result
}

func mergeByName(template, user []map[string]any) []any {
	seen := map[string]bool{}
	result := make([]any, 0, len(template)+len(user))

	for _, it := range user {
		name := pickString(it, "name", "title", "label")
		if name == "" {
			continue
		}
		seen[name] = true
		result = append(result, it)
	}
	for _, it := range template {
		name := pickString(it, "name", "title", "label")
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		result = append(result, it)
	}
	return result
}
