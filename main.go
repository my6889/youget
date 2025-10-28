package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func download(url string) {
    var cmd *exec.Cmd
    if _, err := os.Stat("./cookies.sqlite"); err == nil {
        cmd = exec.Command("you-get", "-c", "cookies.sqlite", url)
    } else if _, err := os.Stat("./cookies.txt"); err == nil {
        cmd = exec.Command("you-get", "-c", "cookies.txt", url)
    } else {
        cmd = exec.Command("you-get", url)
    }
    output, err := cmd.CombinedOutput()
    fmt.Println(string(output))
    if err != nil {
        fmt.Printf("Error: %s\n", err)
    }
}

type youGetJSON struct {
	Title string `json:"title"`
}

func getTitle(url string) (string, error) {
    var cmd *exec.Cmd
    if _, err := os.Stat("./cookies.sqlite"); err == nil {
        cmd = exec.Command("you-get", "-c", "cookies.sqlite", "--json", url)
    } else if _, err := os.Stat("./cookies.txt"); err == nil {
        cmd = exec.Command("you-get", "-c", "cookies.txt", "--json", url)
    } else {
        cmd = exec.Command("you-get", "--json", url)
    }
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("you-get --json failed: %v, output: %s", err, string(output))
    }
    // Extract JSON object in case there are leading logs/noise
    raw := string(output)
    start := strings.Index(raw, "{")
    end := strings.LastIndex(raw, "}")
    var data []byte
    if start != -1 && end != -1 && end >= start {
        data = []byte(raw[start : end+1])
    } else {
        data = output
    }
    var meta youGetJSON
    if err := json.Unmarshal(data, &meta); err != nil {
        return "", fmt.Errorf("parse json failed: %v", err)
    }
    return meta.Title, nil
}

func findDownloadedVideoPath(title string) (string, error) {
	// Try common video extensions using the title as basename
	exts := []string{".mp4", ".mkv", ".flv", ".webm", ".avi", ".mov", ".ts", ".m4v"}
	for _, ext := range exts {
		candidate := title + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	// Fallback: scan current directory for files starting with title and ending with a known ext
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, title) {
			lower := strings.ToLower(name)
			for _, ext := range exts {
				if strings.HasSuffix(lower, ext) {
					return name, nil
				}
			}
		}
	}
	return "", fmt.Errorf("downloaded video not found for title: %s", title)
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
		download(string(a))
		title, err := getTitle(string(a))
		if err != nil {
			fmt.Printf("Get title failed: %v\n", err)
			continue
		}
		videoPath, err := findDownloadedVideoPath(title)
		if err != nil {
			fmt.Printf("Find video failed: %v\n", err)
			continue
		}
		if err := convertToMP3(videoPath); err != nil {
			fmt.Printf("Convert failed: %v\n", err)
		}
	}
}
