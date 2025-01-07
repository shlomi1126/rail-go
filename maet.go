package main

import (
	"fmt"
	"time"
)

func meat() string {
	return fmt.Sprintf(" you eat 🥩 in %v, you can eat 🧀 in %v", time.Now().Format(time.TimeOnly), time.Now().Add(6*time.Hour).Format(time.TimeOnly))
}
