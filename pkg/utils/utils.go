package utils

import (
	"fmt"
	"math"
	"time"

	"github.com/kubeplay/gameserver/pkg/types"
	yaml "gopkg.in/yaml.v2"
)

func RoundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

func GetDeltaDuration(startTime, endTime string) string {
	start, _ := time.Parse(time.RFC3339, startTime)
	end, _ := time.Parse(time.RFC3339, endTime)
	delta := end.Sub(start)
	var d time.Duration
	if endTime != "" {
		d = RoundTime(delta, time.Second)
	} else {
		d = RoundTime(time.Since(start), time.Second)
	}
	switch {
	case d.Hours() >= 24: // day resolution
		return fmt.Sprintf("%.fd", math.Floor(d.Hours()/24))
	case d.Hours() >= 8760: // year resolution
		return fmt.Sprintf("%.fd", math.Floor(d.Hours()/8760))
	}
	return d.String()
}

func YamlToJson(input []byte) (types.Object, error) {
	var typeMeta types.TypeMeta
	if err := yaml.Unmarshal(input, &typeMeta); err != nil {
		return nil, err
	}
	var obj types.Object
	for _, reg := range types.RegisteredTypes {
		if reg.GetObjectKind() == typeMeta.Kind {
			obj = reg.New()
		}
	}
	if obj == nil {
		return nil, fmt.Errorf("type not found: %q", typeMeta.Kind)
	}
	if err := yaml.Unmarshal(input, obj); err != nil {
		return nil, err
	}
	return obj, nil
}
