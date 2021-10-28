package teamcity

import (
	"fmt"
	"time"
)

var daysOfWeek = map[string]time.Weekday{}

func init() {
	for d := time.Sunday; d <= time.Saturday; d++ {
		daysOfWeek[d.String()] = d
	}
}

func parseWeekday(v string) (time.Weekday, error) {
	if d, ok := daysOfWeek[v]; ok {
		return d, nil
	}

	return time.Sunday, fmt.Errorf("invalid weekday '%s'", v)
}

func expandStringSlice(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSet schema.NewSet
func flattenStringSlice(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

func getChangeExpandedStringList(oraw interface{}, nraw interface{}) (remove []string, add []string) {
	old := oraw.([]interface{})
	n := nraw.([]interface{})

	remove = make([]string, 0)
	add = make([]string, 0)

	for _, n := range n {
		if _, contains := sliceContainsString(old, n.(string)); !contains {
			add = append(add, n.(string))
		}
	}
	for _, o := range old {
		if _, contains := sliceContainsString(n, o.(string)); !contains {
			remove = append(remove, o.(string))
		}
	}

	return
}

func sliceContainsString(slice []interface{}, s string) (int, bool) {
	for idx, value := range slice {
		v := value.(string)
		if v == s {
			return idx, true
		}
	}
	return -1, false
}

func expandStringMapConditions(configured []interface{}) [][]string {
	vs := make([][]string, 0, len(configured))

	for _, i := range configured {
		vss := make([]string, 0, 3)
		e := i.(map[string]interface{})
		vss = append(vss, e["condition"].(string))
		vss = append(vss, e["name"].(string))
		vss = append(vss, e["value"].(string))
		vs = append(vs, vss)
	}
	return vs
}

func flattenExecuteConditions(conditions [][]string) []map[string]string {
	ecs := make([]map[string]string, 0)
	for _, v := range conditions {
		mp := make(map[string]string)
		mp["condition"] = v[0]
		mp["name"] = v[1]
		mp["value"] = v[2]
		ecs = append(ecs, mp)
	}
	return ecs
}
