package config

import (
	"fmt"
	"strings"
	"time"
)

type Step struct {
	Name           string
	Description    string
	Variables      []string
	Dir            string
	Command        string
	Args           []string
	Depends        []string
	ContinueOn     ContinueOn
	RetryPolicy    *RetryPolicy
	MailOnError    bool
	Repeat         bool
	RepeatInterval time.Duration
	Preconditions  []*Condition
}

type RetryPolicy struct {
	Limit int
}

type ContinueOn struct {
	Failure bool
	Skipped bool
}

func (s *Step) String() string {
	vals := []string{}
	vals = append(vals, fmt.Sprintf("Name: %s", s.Name))
	vals = append(vals, fmt.Sprintf("Dir: %s", s.Dir))
	vals = append(vals, fmt.Sprintf("Command: %s", s.Command))
	vals = append(vals, fmt.Sprintf("Args: %s", s.Args))
	vals = append(vals, fmt.Sprintf("Depends: [%s]", strings.Join(s.Depends, ", ")))
	return strings.Join(vals, "\t")
}