package pages

import "fmt"

type PageDomain struct {
	Owner  string
	Repo   string
	Branch string
}

func NewPageDomain(owner string, repo string, branch string) *PageDomain {
	return &PageDomain{
		owner,
		repo,
		branch,
	}
}

func (p *PageDomain) key() string {
	return fmt.Sprintf("%s|%s|%s", p.Owner, p.Repo, p.Branch)
}
