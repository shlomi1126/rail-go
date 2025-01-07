package main

import (
	"fmt"
	"time"
)

func cheese() string {
	return fmt.Sprintf(" you eat 🧀 in %v, you can eat 🥩 in %v", time.Now().Format(time.TimeOnly), time.Now().Add(30*time.Minute).Format(time.TimeOnly))

}
