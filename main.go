package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// sanitizeFileName removes invalid characters for Windows file names
func sanitizeFileName(name string) string {
	// Windows doesn't allow: < > : " / \ | ? *
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = invalidChars.ReplaceAllString(name, "-")
	// Remove leading/trailing spaces and dots
	name = strings.Trim(name, " .")
	return name
}

func download(url string) (string, error) {
	var cmd *exec.Cmd
	if _, err := os.Stat("./cookies.sqlite"); err == nil {
		cmd = exec.Command("you-get", "-c", "cookies.sqlite", url)
	} else if _, err := os.Stat("./cookies.txt"); err == nil {
		cmd = exec.Command("you-get", "-c", "cookies.txt", url)
	} else {
		cmd = exec.Command("you-get", url)
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	fmt.Println(outputStr)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
	}

	lines := strings.Split(outputStr, "\n")

	// 1) Prefer title -> sanitized filename
	var titleFilename string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "title:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				title := strings.TrimSpace(parts[1])
				sanitized := sanitizeFileName(title)
				filename := sanitized + ".mp4"
				if _, err := os.Stat(filename); err == nil {
					return filename, nil
				}
				titleFilename = filename
				break
			}
		}
	}

	// 2) Fallback: extract filename from output lines
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if (strings.Contains(trimmed, "Skipping") || strings.Contains(trimmed, "Downloading")) && strings.Contains(trimmed, ".mp4") {
			re := regexp.MustCompile(`[^\s]+\.mp4`)
			matches := re.FindString(trimmed)
			if matches != "" {
				matches = strings.TrimPrefix(matches, ".\\")
				matches = strings.TrimPrefix(matches, "./")
				// Try sanitized filename first (handles cases where the actual saved file has been sanitized)
				base := strings.TrimSuffix(matches, ".mp4")
				sanitized := sanitizeFileName(base) + ".mp4"
				if _, err := os.Stat(sanitized); err == nil {
					return sanitized, nil
				}
				// Fallback to original match if it exists
				if _, err := os.Stat(matches); err == nil {
					return matches, nil
				}
			}
		}
	}

	// 3) If title was found but文件未找到，仍返回基于 title 的文件名
	if titleFilename != "" {
		return titleFilename, nil
	}

	return "", fmt.Errorf("title not found in you-get output and no .mp4 file detected")
}

func convertToMP3(inputPath string) error {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".mp3"
	// If output exists, skip conversion instead of overwriting
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("MP3 already exists, skipping: %s\n", outputPath)
		return nil
	}
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-q:a", "0", "-map", "a", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Converting to mp3: %s -> %s\n", inputPath, outputPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Printf("MP3 created: %s\n", outputPath)
	return nil
}

func checkDependencies() error {
	if _, err := exec.LookPath("you-get"); err != nil {
		fmt.Println("you-get命令未安装或不可用，请确保you-get已被正确安装且位于PATH中。具体安装方式参考README.md。")
		return err
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("ffmpeg命令未安装或不可用，请确保ffmpeg已被正确安装且位于PATH中。具体安装方式可参考README.md。")
		return err
	}
	return nil
}

func main() {
	if err := checkDependencies(); err != nil {
		return
	}
	fi, err := os.Open("./download-list.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		fmt.Println("Downloading " + string(a))
		videoPath, err := download(string(a))
		if err != nil {
			fmt.Printf("Download failed: %v\n", err)
			continue
		}
		if err := convertToMP3(videoPath); err != nil {
			fmt.Printf("Convert failed: %v\n", err)
		}
	}
}
