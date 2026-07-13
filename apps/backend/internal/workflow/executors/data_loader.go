package executors

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func DataLoaderExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		source, _ := config["source"].(string)
		if source == "" {
			source, _ = inputs["file_path"].(string)
		}
		if source == "" {
			source, _ = inputs["path"].(string)
		}
		if source == "" {
			return nil, fmt.Errorf("data source not specified")
		}

		format, _ := config["format"].(string)
		if format == "" {
			ext := filepath.Ext(source)
			switch ext {
			case ".csv":
				format = "csv"
			case ".json":
				format = "json"
			case ".jpg", ".jpeg", ".png", ".bmp", ".gif":
				format = "image"
			default:
				format = "csv"
			}
		}

		switch format {
		case "csv":
			return readCSV(source)
		case "json":
			return readJSON(source)
		case "image":
			return loadImage(source)
		default:
			return readCSV(source)
		}
	}
}

func readCSV(path string) (map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file %s: %w", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file %s: %w", path, err)
	}

	if len(records) == 0 {
		return map[string]interface{}{
			"data":     []interface{}{},
			"rowCount": 0,
			"columns":  []string{},
			"status":   "completed",
		}, nil
	}

	columns := records[0]
	var data []map[string]string
	for _, row := range records[1:] {
		record := make(map[string]string)
		for i, col := range columns {
			if i < len(row) {
				record[col] = row[i]
			}
		}
		data = append(data, record)
	}

	return map[string]interface{}{
		"data":     data,
		"rowCount": len(data),
		"columns":  columns,
		"status":   "completed",
	}, nil
}

func readJSON(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read json file %s: %w", path, err)
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse json file %s: %w", path, err)
	}

	return map[string]interface{}{
		"data":   parsed,
		"status": "completed",
	}, nil
}

func loadImage(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat image file %s: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file %s: %w", path, err)
	}

	return map[string]interface{}{
		"image": string(data),
		"metadata": map[string]interface{}{
			"path":     path,
			"size":     info.Size(),
			"filename": filepath.Base(path),
		},
		"status": "completed",
	}, nil
}
