package skill

import (
	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

func RegisterBuiltinSkills(manager *SkillManager) {
	manager.MustRegister(yoloDetectionSkill())
	manager.MustRegister(imageClassificationSkill())
	manager.MustRegister(imageSegmentationSkill())
	manager.MustRegister(textClassificationSkill())
	manager.MustRegister(audioClassificationSkill())
}

func yoloDetectionSkill() *Skill {
	wf := &workflow.Workflow{
		ID:          uuid.New().String(),
		Name:        "YOLO Object Detection",
		Description: "Train a YOLO model for object detection",
		Version:     1,
		Target:      workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "data-loader", Type: workflow.NodeTypeDataLoader, Name: "Data Loader",
				Config: map[string]any{"format": "coco", "dataset": "${dataset_path}"}},
			{ID: "data-augmentation", Type: workflow.NodeTypeDataAugment, Name: "Data Augmentation",
				Config: map[string]any{"mosaic": 1.0, "hsv_h": 0.015, "hsv_s": 0.7, "hsv_v": 0.4}},
			{ID: "model-trainer", Type: workflow.NodeTypeModelTrainer, Name: "YOLO Trainer",
				Config: map[string]any{"model": "${model_name}", "epochs": "${epochs}", "batch_size": "${batch_size}",
					"imgsz": 640, "optimizer": "auto", "lr": 0.01}},
			{ID: "model-evaluator", Type: workflow.NodeTypeModelEvaluator, Name: "Model Evaluator",
				Config: map[string]any{"metrics": []string{"mAP50", "mAP50-95", "precision", "recall"}}},
			{ID: "model-exporter", Type: workflow.NodeTypeModelExporter, Name: "Model Exporter",
				Config: map[string]any{"format": "onnx", "half": false}},
		},
		Edges: []workflow.Edge{
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-loader", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "data-augmentation", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-augmentation", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-exporter", PortID: "input"}},
		},
	}
	return &Skill{
		ID:          "yolo-detection",
		Name:        "YOLO Object Detection",
		Description: "Train a YOLOv8/v5 model for object detection with support for COCO format datasets",
		Version:     "1.0.0",
		Author:      "AIStudio",
		Category:    CategoryObjectDetection,
		Tags:        []string{"yolo", "detection", "vision", "pytorch"},
		Workflow:    wf,
	}
}

func imageClassificationSkill() *Skill {
	wf := &workflow.Workflow{
		ID:          uuid.New().String(),
		Name:        "Image Classification",
		Description: "Train a classification model on image datasets",
		Version:     1,
		Target:      workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "data-loader", Type: workflow.NodeTypeDataLoader, Name: "Data Loader",
				Config: map[string]any{"format": "folder", "split": "train/val", "dataset": "${dataset_path}"}},
			{ID: "data-preprocessor", Type: workflow.NodeTypeDataPreprocess, Name: "Preprocessor",
				Config: map[string]any{"resize": 224, "normalize": true, "augmentation": "basic"}},
			{ID: "model-trainer", Type: workflow.NodeTypeModelTrainer, Name: "Classification Trainer",
				Config: map[string]any{"model": "${model_name}", "epochs": "${epochs}", "batch_size": "${batch_size}",
					"num_classes": "${num_classes}", "lr": 0.001, "weight_decay": 0.0001}},
			{ID: "model-evaluator", Type: workflow.NodeTypeModelEvaluator, Name: "Evaluator",
				Config: map[string]any{"metrics": []string{"accuracy", "precision", "recall", "f1"}}},
			{ID: "model-exporter", Type: workflow.NodeTypeModelExporter, Name: "Exporter",
				Config: map[string]any{"format": "onnx"}},
		},
		Edges: []workflow.Edge{
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-loader", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "data-preprocessor", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-preprocessor", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-exporter", PortID: "input"}},
		},
	}
	return &Skill{
		ID:          "image-classification",
		Name:        "Image Classification",
		Description: "Train an image classification model using ResNet, EfficientNet, or ViT backbones",
		Version:     "1.0.0",
		Author:      "AIStudio",
		Category:    CategoryClassification,
		Tags:        []string{"classification", "vision", "resnet", "efficientnet"},
		Workflow:    wf,
	}
}

func imageSegmentationSkill() *Skill {
	wf := &workflow.Workflow{
		ID:          uuid.New().String(),
		Name:        "Image Segmentation",
		Description: "Train a segmentation model for pixel-level classification",
		Version:     1,
		Target:      workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "data-loader", Type: workflow.NodeTypeDataLoader, Name: "Data Loader",
				Config: map[string]any{"format": "coco", "task": "segmentation", "dataset": "${dataset_path}"}},
			{ID: "data-augmentation", Type: workflow.NodeTypeDataAugment, Name: "Augmentation",
				Config: map[string]any{"flip": true, "rotate": 10, "scale": 0.5}},
			{ID: "feature-extractor", Type: workflow.NodeTypeFeatureExtract, Name: "Backbone",
				Config: map[string]any{"model": "${backbone}", "pretrained": true}},
			{ID: "model-trainer", Type: workflow.NodeTypeModelTrainer, Name: "Segmentation Trainer",
				Config: map[string]any{"model": "${model_name}", "epochs": "${epochs}", "batch_size": "${batch_size}",
					"num_classes": "${num_classes}", "lr": 0.0001}},
			{ID: "model-evaluator", Type: workflow.NodeTypeModelEvaluator, Name: "Evaluator",
				Config: map[string]any{"metrics": []string{"mIoU", "Dice", "PixelAccuracy"}}},
		},
		Edges: []workflow.Edge{
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-loader", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "data-augmentation", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-augmentation", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "feature-extractor", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "feature-extractor", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "input"}},
		},
	}
	return &Skill{
		ID:          "image-segmentation",
		Name:        "Image Segmentation",
		Description: "Train a semantic or instance segmentation model using UNet, DeepLab, or Mask R-CNN",
		Version:     "1.0.0",
		Author:      "AIStudio",
		Category:    CategorySegmentation,
		Tags:        []string{"segmentation", "vision", "unet", "deeplab"},
		Workflow:    wf,
	}
}

func textClassificationSkill() *Skill {
	wf := &workflow.Workflow{
		ID:          uuid.New().String(),
		Name:        "Text Classification",
		Description: "Train a text classification model with transformer-based embeddings",
		Version:     1,
		Target:      workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "data-loader", Type: workflow.NodeTypeDataLoader, Name: "Data Loader",
				Config: map[string]any{"format": "csv", "text_column": "text", "label_column": "label",
					"dataset": "${dataset_path}"}},
			{ID: "data-preprocessor", Type: workflow.NodeTypeDataPreprocess, Name: "Text Preprocessor",
				Config: map[string]any{"max_length": 512, "padding": true, "truncation": true}},
			{ID: "model-trainer", Type: workflow.NodeTypeModelTrainer, Name: "Text Classifier",
				Config: map[string]any{"model": "${model_name}", "epochs": "${epochs}", "batch_size": "${batch_size}",
					"num_classes": "${num_classes}", "lr": 2e-5, "weight_decay": 0.01}},
			{ID: "model-evaluator", Type: workflow.NodeTypeModelEvaluator, Name: "Evaluator",
				Config: map[string]any{"metrics": []string{"accuracy", "f1", "precision", "recall"}}},
		},
		Edges: []workflow.Edge{
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-loader", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "data-preprocessor", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-preprocessor", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "input"}},
		},
	}
	return &Skill{
		ID:          "text-classification",
		Name:        "Text Classification",
		Description: "Fine-tune a transformer model (BERT, RoBERTa, etc.) for text classification tasks",
		Version:     "1.0.0",
		Author:      "AIStudio",
		Category:    CategoryNLP,
		Tags:        []string{"nlp", "classification", "bert", "transformer"},
		Workflow:    wf,
	}
}

func audioClassificationSkill() *Skill {
	wf := &workflow.Workflow{
		ID:          uuid.New().String(),
		Name:        "Audio Classification",
		Description: "Train an audio classification model using spectrogram features",
		Version:     1,
		Target:      workflow.TargetPython,
		Nodes: []workflow.Node{
			{ID: "data-loader", Type: workflow.NodeTypeDataLoader, Name: "Audio Loader",
				Config: map[string]any{"format": "folder", "sample_rate": 16000, "dataset": "${dataset_path}"}},
			{ID: "feature-extractor", Type: workflow.NodeTypeFeatureExtract, Name: "Feature Extractor",
				Config: map[string]any{"features": "mel-spectrogram", "n_mels": 128, "n_fft": 1024, "hop_length": 512}},
			{ID: "model-trainer", Type: workflow.NodeTypeModelTrainer, Name: "Audio Classifier",
				Config: map[string]any{"model": "${model_name}", "epochs": "${epochs}", "batch_size": "${batch_size}",
					"num_classes": "${num_classes}", "lr": 0.001}},
			{ID: "model-evaluator", Type: workflow.NodeTypeModelEvaluator, Name: "Evaluator",
				Config: map[string]any{"metrics": []string{"accuracy", "f1"}}},
		},
		Edges: []workflow.Edge{
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "data-loader", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "feature-extractor", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "feature-extractor", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "input"}},
			{ID: uuid.New().String(), Source: workflow.EdgeEndpoint{NodeID: "model-trainer", PortID: "output"}, Target: workflow.EdgeEndpoint{NodeID: "model-evaluator", PortID: "input"}},
		},
	}
	return &Skill{
		ID:          "audio-classification",
		Name:        "Audio Classification",
		Description: "Train a model for audio classification using mel-spectrogram features and CNN backbones",
		Version:     "1.0.0",
		Author:      "AIStudio",
		Category:    CategoryAudio,
		Tags:        []string{"audio", "classification", "spectrogram"},
		Workflow:    wf,
	}
}