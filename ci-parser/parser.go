package ciparser

import (
	"fmt"
	"strings"
)

const (
	spliter    = "|"
	rowSpliter = "\n"
)

type CIParser interface {
	GetEachJobComment(string) ([]string, error)
	ParseJobStatus(string) (string, error)
}

func ParseCIComment(t CIParser, comment string) ([]string, error) {
	cs, err := t.GetEachJobComment(comment)
	if err != nil {
		return nil, err
	}

	r := make([]string, 0, len(cs))
	for _, c := range cs {
		if status, err := t.ParseJobStatus(c); err == nil {
			r = append(r, status)
		}
	}

	return r, nil
}

type CITable struct {
	// Title is the one of table for CI comment of PR. The format of comment may be like this.
	//
	//   | job name | result | detail |
	//   | --- | --- | --- |
	//   | test     | success | link   |
	//
	// The value of Title for ci comment above is
	// `| job name | result | detail |`
	Title string `json:"title" required:"true"`

	// ResultColumnNum is the column number of job result.
	ResultColumnNum int `json:"result_column_num" required:"true"`

	totleColumnNum int
}

func (t *CITable) Validate() error {
	n := numOfColumns(t.Title)
	if n <= 0 {
		return fmt.Errorf("title is not the one of CI table")
	}

	if t.ResultColumnNum > n {
		return fmt.Errorf("result_column_num must be <= %d", n)
	}

	if t.ResultColumnNum <= 0 {
		return fmt.Errorf("result_column_num must be bigger than 0")
	}

	t.totleColumnNum = n
	return nil
}

func (t CITable) IsCIComment(s string) bool {
	return strings.Count(s, t.Title+rowSpliter) == 1
}

func (t CITable) GetEachJobComment(c string) ([]string, error) {
	v := strings.Split(c, t.Title+rowSpliter)
	if len(v) != 2 {
		return nil, fmt.Errorf("invalid CI comment")
	}

	items := strings.Split(v[1], rowSpliter)
	n := len(items)
	if n < 2 {
		return nil, fmt.Errorf("invalid table")
	}

	for i := n - 1; i > 0; i-- {
		if _, err := t.parseJobResult(items[i]); err == nil {
			// The items[0] is like | --- | --- |, so ignore it.
			return items[1 : i+1], nil
		}
	}

	return nil, fmt.Errorf("empty table")
}

// parseJobResult return the single job result.
func (t CITable) parseJobResult(s string) (string, error) {
	if n := numOfColumns(s); n != t.totleColumnNum {
		return "", fmt.Errorf("invalid job comment")
	}

	return strings.Split(s, spliter)[t.ResultColumnNum], nil
}

func numOfColumns(t string) int {
	n := strings.Count(t, spliter)
	if n > 0 {
		return n - 1
	}
	return n
}

type JobStatusDesc struct {
	Desc     []string
	Status   string
	Priority int
}

func (j JobStatusDesc) isDescMatched(desc string) bool {
	for _, item := range j.Desc {
		if strings.Contains(desc, item) {
			return true
		}
	}
	return false
}

type CIParserImpl struct {
	CITable

	JobStatus []JobStatusDesc
}

func (p CIParserImpl) ParseJobStatus(c string) (string, error) {
	desc, err := p.parseJobResult(c)
	if err != nil {
		return "", err
	}

	for _, v := range p.JobStatus {
		if v.isDescMatched(desc) {
			return v.Status, nil
		}
	}
	return "", fmt.Errorf("unknown job description")
}

func (p CIParserImpl) InferFinalStatus(status []string) string {
	sn := make(map[string]bool)
	for _, item := range status {
		sn[item] = true
	}

	cp := -1
	s := ""
	for _, item := range p.JobStatus {
		if sn[item.Status] && (s == "" || item.Priority > cp) {
			cp = item.Priority
			s = item.Status
		}
	}
	return s
}
