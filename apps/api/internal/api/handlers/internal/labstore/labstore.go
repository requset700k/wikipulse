// Package labstore holds the mock lab catalog used by handlers.
// Will be replaced by a DB-backed repository in a later sprint.
package labstore

import "github.com/kt-techuplabs/cledyu/backend/internal/domain/lab"

var catalog = []lab.Lab{
	{
		ID:          "lab-001",
		Title:       "Linux Basics",
		Description: "파일 시스템, 프로세스, 네트워크 기초 명령어를 실습합니다.",
		Difficulty:  lab.DifficultyBeginner,
		DurationMin: 60,
		Tags:        []string{"linux", "shell"},
		VMType:      "small",
		StepCount:   8,
	},
	{
		ID:          "lab-002",
		Title:       "Ansible Fundamentals",
		Description: "Playbook, Role, Inventory를 실제 환경에서 작성하고 실행합니다.",
		Difficulty:  lab.DifficultyIntermediate,
		DurationMin: 90,
		Tags:        []string{"ansible", "automation"},
		VMType:      "small",
		StepCount:   10,
	},
	{
		ID:          "lab-003",
		Title:       "Terraform Introduction",
		Description: "HCL로 인프라를 코드로 정의하고 AWS 리소스를 프로비저닝합니다.",
		Difficulty:  lab.DifficultyIntermediate,
		DurationMin: 120,
		Tags:        []string{"terraform", "iac", "aws"},
		VMType:      "medium",
		StepCount:   12,
	},
	{
		ID:          "lab-004",
		Title:       "Kubernetes Introduction",
		Description: "Pod, Deployment, Service를 직접 생성하고 클러스터를 운영합니다.",
		Difficulty:  lab.DifficultyAdvanced,
		DurationMin: 150,
		Tags:        []string{"kubernetes", "k8s", "container"},
		VMType:      "medium",
		StepCount:   15,
	},
}

func All() []lab.Lab { return catalog }

func Find(id string) (lab.Lab, bool) {
	for _, l := range catalog {
		if l.ID == id {
			return l, true
		}
	}
	return lab.Lab{}, false
}
