package executors

import (
	"context"
	"fmt"
)

func SwitchExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		switchValue := inputs["switchValue"]

		cases, _ := config["cases"].([]interface{})

		selectedCase := "default"
		selectedPort := "default"

		for _, c := range cases {
			caseObj, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			val := caseObj["value"]
			port, _ := caseObj["outputPort"].(string)
			if port == "" {
				port, _ = caseObj["port"].(string)
			}

			if fmt.Sprintf("%v", val) == fmt.Sprintf("%v", switchValue) {
				selectedCase = fmt.Sprintf("%v", val)
				selectedPort = port
				break
			}
		}

		return map[string]interface{}{
			"selectedCase": selectedCase,
			"selectedPort": selectedPort,
			"value":        switchValue,
		}, nil
	}
}
