package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/aimv/media-gallery/internal/domain/media"
)

type FFmpegProcessor struct {
	ffmpegPath  string
	ffprobePath string
	root        string
}

func NewFFmpegProcessor(ffmpegPath, ffprobePath, root string) *FFmpegProcessor {
	return &FFmpegProcessor{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		root:        root,
	}
}

// Probe извлекает метаданные видео через ffprobe.
func (p *FFmpegProcessor) Probe(ctx context.Context, filePath string) (*media.VideoMetadata, error) {
	fullPath := filepath.Join(p.root, filePath)
	cmd := exec.CommandContext(ctx, p.ffprobePath,
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		fullPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probe struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	meta := &media.VideoMetadata{}
	if dur, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		meta.Duration = time.Duration(dur * float64(time.Second))
	}
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			meta.Width = stream.Width
			meta.Height = stream.Height
			meta.VideoCodec = stream.CodecName
		case "audio":
			meta.AudioCodec = stream.CodecName
		}
	}
	return meta, nil
}

// ConvertToHLS конвертирует исходное видео в HLS и атомарно перемещает
// результат в указанную выходную директорию (относительно root).
func (p *FFmpegProcessor) ConvertToHLS(ctx context.Context, inputPath, outputDir string) error {
	inputAbs := filepath.Join(p.root, inputPath)
	finalDir := filepath.Join(p.root, outputDir)

	// Создаём корневую временную директорию, если её ещё нет
	tempRoot := filepath.Join(p.root, "tmp")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return fmt.Errorf("create temp root dir: %w", err)
	}

	// Создаём уникальную временную директорию для генерации HLS
	tempDir, err := os.MkdirTemp(tempRoot, "hls-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // очистка при ошибке

	// Формируем аргументы ffmpeg
	masterPlaylist := filepath.Join(tempDir, "master.m3u8")
	segmentPattern := filepath.Join(tempDir, "seg_%05d.ts")

	args := []string{
		"-i", inputAbs,
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-c:a", "aac",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		masterPlaylist,
	}

	cmd := exec.CommandContext(ctx, p.ffmpegPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	// Атомарное перемещение: удаляем старую финальную директорию (если есть)
	if _, err := os.Stat(finalDir); err == nil {
		if err := os.RemoveAll(finalDir); err != nil {
			return fmt.Errorf("remove old final dir: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(finalDir), 0755); err != nil {
		return fmt.Errorf("create parent for final dir: %w", err)
	}
	if err := os.Rename(tempDir, finalDir); err != nil {
		return fmt.Errorf("rename temp to final: %w", err)
	}

	return nil
}
