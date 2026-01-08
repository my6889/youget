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
	"time"
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
    
    // Record time before download to find the newest file
    beforeTime := time.Now()
    
    output, err := cmd.CombinedOutput()
    outputStr := string(output)
    fmt.Println(outputStr)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return "", err
    }
    
    // Try to extract filename from you-get output (e.g., "Skipping .\filename.mp4" or "Downloading filename.mp4")
    lines := strings.Split(outputStr, "\n")
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        // Look for "Skipping" or "Downloading" lines that contain .mp4
        if (strings.Contains(trimmed, "Skipping") || strings.Contains(trimmed, "Downloading")) && strings.Contains(trimmed, ".mp4") {
            // Extract filename - look for .mp4 and extract the filename
            re := regexp.MustCompile(`[^\s]+\.mp4`)
            matches := re.FindString(trimmed)
            if matches != "" {
                // Remove leading .\ or ./
                matches = strings.TrimPrefix(matches, ".\\")
                matches = strings.TrimPrefix(matches, "./")
                // Check if file exists
                if _, err := os.Stat(matches); err == nil {
                    return matches, nil
                }
            }
        }
    }
    
    // Try to find the actual downloaded file by looking for the newest .mp4 file
    entries, err := os.ReadDir(".")
    if err != nil {
        return "", fmt.Errorf("failed to read directory: %v", err)
    }
    
    var newestFile string
    var newestTime time.Time
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if strings.HasSuffix(strings.ToLower(name), ".mp4") {
            info, err := e.Info()
            if err != nil {
                continue
            }
            modTime := info.ModTime()
            // Only consider files modified after download started (or check all if file was skipped)
            if modTime.After(beforeTime) || modTime.Equal(beforeTime) {
                if newestFile == "" || modTime.After(newestTime) {
                    newestFile = name
                    newestTime = modTime
                }
            }
        }
    }
    
    if newestFile != "" {
        return newestFile, nil
    }
    
    // Fallback: parse title from output and sanitize it
    for _, line := range lines {
        if strings.HasPrefix(strings.TrimSpace(line), "title:") {
            parts := strings.SplitN(line, ":", 2)
            if len(parts) == 2 {
                title := strings.TrimSpace(parts[1])
                sanitized := sanitizeFileName(title)
                filename := sanitized + ".mp4"
                // Check if file exists with sanitized name
                if _, err := os.Stat(filename); err == nil {
                    return filename, nil
                }
                return filename, nil
            }
        }
    }
    
    return "", fmt.Errorf("title not found in you-get output and no .mp4 file detected")
}


func convertToMP3(inputPath string) error {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".mp3"
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
