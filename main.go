package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"h5pspeeder/h5p"
)

const (
	startID = 824962
	endID   = 825163
	step    = 1
)

func main() {
	configFile := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	fmt.Printf("\n========== 四史助手 v1.0 ==========\n\n")

	fmt.Printf("正在加载配置文件: %s\n", *configFile)
	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	courses := generateCourses()
	fmt.Printf("共生成 %d 个课程 (ID: %d - %d, 步长: %d)\n\n", len(courses), startID, endID, step)

	client := h5p.NewClient(config.Cookie, config.Sesskey)
	speeder := h5p.NewSpeeder(client, config.Settings)

	if err := speeder.RunAll(courses); err != nil {
		fmt.Printf("执行过程中出现错误\n")
	}

	fmt.Printf("\n完成！\n")
}

func generateCourses() []h5p.Course {
	var courses []h5p.Course
	for id := startID; id <= endID; id += step {
		courses = append(courses, h5p.Course{
			URL:  fmt.Sprintf("https://moodle.scnu.edu.cn/mod/h5pactivity/view.php?id=%d", id),
			Name: fmt.Sprintf("课程 %d", id),
		})
	}
	return courses
}

func loadConfig(filename string) (*h5p.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config h5p.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Cookie == "" {
		return nil, fmt.Errorf("cookie 不能为空")
	}
	if config.Sesskey == "" {
		return nil, fmt.Errorf("sesskey 不能为空")
	}

	if config.Settings.TargetProgress == 0 {
		config.Settings.TargetProgress = 95
	}
	if config.Settings.StepProgress == 0 {
		config.Settings.StepProgress = 3
	}
	if config.Settings.IntervalSeconds == 0 {
		config.Settings.IntervalSeconds = 0.5
	}
	if config.Settings.MaxConcurrent == 0 {
		config.Settings.MaxConcurrent = 10
	}

	return &config, nil
}
