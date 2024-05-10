package pages

import "fmt"

type PageDomain struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

func NewPageDomain(owner string, repo string, branch string) *PageDomain {
	return &PageDomain{
		owner,
		repo,
		branch,
	}
}

func (p *PageDomain) Key() string {
	return fmt.Sprintf("%s|%s|%s", p.Owner, p.Repo, p.Branch)
}
