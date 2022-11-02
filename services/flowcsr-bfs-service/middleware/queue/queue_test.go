package queue

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {

	InitQ()

	item := Item{
		EnvId:       1,
		TaskId:      1,
		RuleVersion: "ruleVersion",
		RepoList:    []CompileRepo{},
	}
	Q().Enqueue(item)
	item = Item{
		EnvId:       2,
		TaskId:      2,
		RuleVersion: "ruleVersion",
		RepoList:    []CompileRepo{},
	}
	Q().Enqueue(item)
	a := Q().Dequeue()
	fmt.Println(a.EnvId)
	fmt.Println(a.TaskId)
	a = Q().Dequeue()
	fmt.Println(a.EnvId)
	fmt.Println(a.TaskId)
}
