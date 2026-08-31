// internal/infrastructure/video/ffmpeg_processor.go

// Package video содержит реализации видеообработки через ffmpeg/ffprobe.
package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/aimv/media-gallery/internal/domain/service"
	"github.com/aimv/media-gallery/internal/pkg/apperror"
)

// FFmpegProcessor реализует интерфейс service.VideoProcessor с использованием ffmpeg/ffprobe.
type FFmpegProcessor struct {
	threads     int
	ffmpegPath  string // путь к исполняемому файлу ffmpeg (для подмены в тестах)
	ffprobePath string // путь к исполняемому файлу ffprobe (для подмены в тестах)
}

// NewFFmpegProcessor создаёт новый экземпляр FFmpegProcessor с заданным числом потоков.
func NewFFmpegProcessor(threads int) *FFmpegProcessor {
	return &FFmpegProcessor{
		threads:     threads,
		ffmpegPath:  "ffmpeg",
		ffprobePath: "ffprobe",
	}
}

// ProbeMetadata запускает ffprobe и извлекает технические метаданные видео.
func (p *FFmpegProcessor) ProbeMetadata(ctx context.Context, filePath string) (*service.VideoMetadata, error) {
	// #nosec G204 -- пути к исполняемым файлам не поступают из пользовательского ввода
	cmd := exec.CommandContext(ctx,
		p.ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration,codec_name",
		"-of", "json",
		filePath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		slog.Error("ffprobe failed",
			"file", filePath,
			"stderr", stderr.String(),
			"error", err,
		)
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	var probe struct {
		Streams []struct {
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			Duration  string `json:"duration"`
			CodecName string `json:"codec_name"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &probe); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	if len(probe.Streams) == 0 {
		return nil, fmt.Errorf("no video stream found")
	}

	stream := probe.Streams[0]
	if stream.Width == 0 || stream.Height == 0 || stream.CodecName == "" {
		return nil, fmt.Errorf("incomplete video metadata")
	}

	durationMS := int64(0)
	if stream.Duration != "" {
		durSec, err := strconv.ParseFloat(stream.Duration, 64)
		if err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}
		durationMS = int64(durSec * 1000)
	}

	return &service.VideoMetadata{
		Width:      stream.Width,
		Height:     stream.Height,
		DurationMS: durationMS,
		Codec:      stream.CodecName,
	}, nil
}

// Validate проверяет видеофайл, извлекая метаданные. В случае ошибки возвращает apperror.ErrInvalidInput.
func (p *FFmpegProcessor) Validate(ctx context.Context, filePath string) (*service.VideoMetadata, error) {
	meta, err := p.ProbeMetadata(ctx, filePath)
	if err != nil {
		return nil, apperror.NewAppError(
			apperror.ErrInvalidInput.Code,
			fmt.Sprintf("invalid video file: %v", err),
			apperror.ErrInvalidInput.HTTPStatus,
		)
	}
	return meta, nil
}

// ProcessToHLS конвертирует видео в HLS-поток, сохраняя сегменты и плейлист в outputDir.
func (p *FFmpegProcessor) ProcessToHLS(ctx context.Context, filePath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// #nosec G204 -- пути и параметры контролируются приложением, пользовательский ввод отсутствует
	cmd := exec.CommandContext(ctx,
		p.ffmpegPath,
		"-i", filePath,
		"-threads", strconv.Itoa(p.threads),
		"-preset", "veryfast",
		"-g", "48",
		"-sc_threshold", "0",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_segment_filename", outputDir+"/seg_%03d.ts",
		outputDir+"/master.m3u8",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		slog.Error("ffmpeg failed",
			"input", filePath,
			"output", outputDir,
			"stderr", stderr.String(),
			"error", err,
		)
		return fmt.Errorf("ffmpeg: %w", err)
	}

	return nil
}
