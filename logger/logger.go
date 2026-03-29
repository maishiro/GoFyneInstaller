package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

// discardErrorWriter は書き込みエラーを無視するラッパーです。
// コンソールがない GUI モードでの os.Stdout への書き込み失敗を防ぎます。
type discardErrorWriter struct {
	w io.Writer
}

func (d discardErrorWriter) Write(p []byte) (int, error) {
	n, _ := d.w.Write(p)
	return n, nil // エラーを握りつぶして常に成功を返す
}

// InitLogger はログ出力を初期化し、標準出力とファイルの両方に出力するように設定します。
// ログファイルは Windows のテンポラリフォルダに作成されます。
func InitLogger(appName string) error {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("Install-%s-%s.log", appName, timestamp)
	tempDir := os.TempDir()
	logPath := filepath.Join(tempDir, fileName)

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	logFile = f

	// ファイルへの書き込みを優先し、標準出力のエラーは無視する
	multiWriter := io.MultiWriter(logFile, discardErrorWriter{os.Stdout})
	log.SetOutput(multiWriter)

	// ログの基本設定（日付、時刻、ソースファイル名と行番号）
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("--- Installer Session Started ---")
	log.Printf("Log file: %s", logPath)
	return nil
}

// Close はログファイルをクローズします。
func Close() {
	if logFile != nil {
		log.Printf("--- Installer Session Ended ---")
		logFile.Close()
	}
}
