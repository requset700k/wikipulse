// Package lab는 Lab 도메인 타입을 정의한다.
// 프론트의 lib/types.ts의 Lab, Difficulty와 1:1 대응.
package lab

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

type Lab struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Difficulty  Difficulty `json:"difficulty"`
	DurationMin int        `json:"duration_min"`
	Tags        []string   `json:"tags"`
	VMType      string     `json:"vm_type"`
	StepCount   int        `json:"step_count"`
}

type Step struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	HintLevels  []string `json:"hint_levels"`
}
