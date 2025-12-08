package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/zstd"
)

// compressFileGzip 读取指定的输入文件，使用 GZIP 压缩，并将结果写入
// 以 .gz 结尾的新文件。compressionLevel 使用 gzip 的级别（-1 表示默认）。
func compressFileGzip(inputPath string, compressionLevel int) (outputPath string, err error) {
	// 1. 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file %s: %w", inputPath, err)
	}
	defer func() {
		inputFile.Close()
		os.Remove(inputPath)
	}()

	// 2. 确定输出文件名
	outputPath = inputPath + ".gz"
	if strings.HasSuffix(inputPath, ".gz") {
		// 防止重复添加 .gz 后缀
		outputPath = inputPath
	}

	// 检查输出文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		return "", fmt.Errorf("output file %s already exists. Aborting to prevent overwrite", outputPath)
	}

	// 3. 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	// 注意：文件句柄必须在 gzip Writer 关闭后才能关闭
	defer outputFile.Close()

	// 4. 创建 GZIP 压缩写入器
	level := compressionLevel
	if level == 0 {
		level = gzip.DefaultCompression
	}
	// gzip.NewWriterLevel 接受 -2..9（-2: HuffmanOnly? implementations may vary）但我们仅保证合理范围
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	gzWriter, err := gzip.NewWriterLevel(outputFile, level)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}

	// 确保所有缓冲区数据被写入
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil {
			if err == nil {
				// 仅在之前没有错误时记录关闭错误
				err = fmt.Errorf("error closing gzip writer: %w", closeErr)
			}
		}
	}()

	// 5. 将输入文件内容流式复制到 gzip 写入器
	_, err = io.Copy(gzWriter, inputFile)
	if err != nil {
		// 如果复制失败，尝试删除创建的输出文件
		os.Remove(outputPath)
		return "", fmt.Errorf("error copying data and compressing: %w", err)
	}

	return outputPath, nil
}

// compressFile 读取指定的输入文件，使用 Snappy 压缩，并将结果写入
// 以 .snappy 结尾的新文件。
func compressFileSnappy(inputPath string) (outputPath string, err error) {
	// 1. 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file %s: %w", inputPath, err)
	}
	defer func() {
		inputFile.Close()
		os.Remove(inputPath)
	}()

	// 2. 确定输出文件名
	outputPath = inputPath + ".snappy"
	if strings.HasSuffix(inputPath, ".snappy") {
		// 防止重复添加 .snappy 后缀
		outputPath = inputPath
	}

	// 检查输出文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		return "", fmt.Errorf("output file %s already exists. Aborting to prevent overwrite", outputPath)
	}

	// 3. 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer outputFile.Close()

	// 4. 创建 Snappy 压缩写入器
	// snappy.NewWriter 实现了 io.WriteCloser 接口，它将写入的数据进行 Snappy 压缩。
	snappyWriter := snappy.NewBufferedWriter(outputFile)
	// 确保所有缓冲区数据被写入，并且 Snappy 块格式正确关闭。
	defer func() {
		if closeErr := snappyWriter.Close(); closeErr != nil {
			if err == nil {
				// 仅在之前没有错误时记录关闭错误
				err = fmt.Errorf("error closing snappy writer: %w", closeErr)
			}
		}
	}()

	// 5. 将输入文件内容复制到 Snappy 写入器
	// io.Copy 会读取 inputFile 的所有内容，并写入到 snappyWriter 中
	_, err = io.Copy(snappyWriter, inputFile)
	if err != nil {
		// 如果复制失败，尝试删除创建的输出文件
		os.Remove(outputPath)
		return "", fmt.Errorf("error copying data and compressing: %w", err)
	}

	// fmt.Printf("Successfully compressed %s (%d bytes) to %s\n", inputPath, bytesWritten, outputPath)
	return outputPath, nil
}

// compressFileZstd 读取指定的输入文件，使用 Zstandard 压缩，并将结果写入
// 以 .zst 结尾的新文件。
// compressionLevel 参数用于控制压缩率和内存使用。
func compressFileZstd(inputPath string, compressionLevel zstd.EncoderLevel) (outputPath string, err error) {
	// 1. 打开输入文件
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file %s: %w", inputPath, err)
	}
	defer func() {
		inputFile.Close()
		os.Remove(inputPath)
	}()

	// 2. 确定输出文件名
	outputPath = inputPath + ".zst"
	// 避免重复后缀
	if strings.HasSuffix(inputPath, ".zst") {
		outputPath = inputPath
	}

	// 检查输出文件是否已存在
	if _, err := os.Stat(outputPath); err == nil {
		return "", fmt.Errorf("output file %s already exists. Aborting to prevent overwrite", outputPath)
	}

	// 3. 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	// 注意：文件句柄必须在 Zstd Writer 关闭后才能关闭
	defer outputFile.Close()

	// 4. 创建 Zstd 压缩写入器
	// zstd.NewWriter 实现了 io.WriteCloser 接口。
	// 关键：在这里设置压缩级别，控制内存使用。
	zstdWriter, err := zstd.NewWriter(outputFile, zstd.WithEncoderLevel(compressionLevel))
	if err != nil {
		return "", fmt.Errorf("failed to create Zstd writer: %w", err)
	}

	// 确保所有缓冲区数据被写入
	defer func() {
		if closeErr := zstdWriter.Close(); closeErr != nil {
			if err == nil {
				// 仅在之前没有错误时记录关闭错误
				err = fmt.Errorf("error closing Zstd writer: %w", closeErr)
			}
		}
	}()

	// 5. 将输入文件内容流式复制到 Zstd 写入器
	_, err = io.Copy(zstdWriter, inputFile)
	if err != nil {
		// 如果复制失败，尝试删除创建的输出文件
		os.Remove(outputPath)
		return "", fmt.Errorf("error copying data and compressing: %w", err)
	}

	// fmt.Printf("Successfully compressed %s (%d bytes) to %s with level %d\n",
	// 	inputPath, bytesWritten, outputPath, compressionLevel)
	return outputPath, nil
}
