package pathtool

import (
	"fmt"
	"io"
	"os"
	"testing"
)

// 创建一个指定大小的临时测试文件
func createTestFile(size int64) (string, error) {
	f, err := os.CreateTemp("", "test_data_*.bin")
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 快速填充大文件
	if err := f.Truncate(size); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func BenchmarkCopyBuffer(b *testing.B) {
	// 测试文件大小：128MB
	fileSize := int64(128 * 1024 * 1024)
	srcFile, err := createTestFile(fileSize)
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(srcFile)

	// 定义不同的缓存大小进行对比
	bufferSizes := []int{
		4 * 1024,        // 4KB
		32 * 1024,       // 32KB (Go 默认值)
		64 * 1024,       // 64KB
		128 * 1024,      // 128KB
		512 * 1024,      // 512KB
		1024 * 1024,     // 1MB
		4 * 1024 * 1024, // 4MB
	}

	for _, size := range bufferSizes {
		b.Run(fmt.Sprintf("BufferSize-%dKB", size/1024), func(b *testing.B) {
			// 开启内存分配统计
			b.ReportAllocs()
			// 设置单次测试的数据吞吐量
			b.SetBytes(fileSize)

			for i := 0; i < b.N; i++ {
				b.StopTimer() // 停止计时，排除准备工作时间

				src, _ := os.Open(srcFile)
				dst, _ := os.Create(os.DevNull) // 写入黑洞，排除磁盘写入瓶颈，纯测拷贝逻辑
				buf := make([]byte, size)

				b.StartTimer() // 开始计时

				_, err := io.CopyBuffer(dst, src, buf)
				if err != nil {
					b.Error(err)
				}

				src.Close()
				dst.Close()
			}
		})
	}
}
