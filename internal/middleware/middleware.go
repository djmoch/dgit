// See LICENSE file for copyright and license details

package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"djmo.ch/dgit/config"
	"djmo.ch/dgit/internal/projectlist"
	"djmo.ch/dgit/internal/repo"
	"djmo.ch/dgit/internal/request"
	"github.com/go-git/go-git/v5/plumbing"
)

func Get(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method not allowed")
			return
		}
		h(w, r)
	}
}

func ResolveHead(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctxRepo := r.Context().Value("repo")
		if ctxRepo == nil {
			h(w, r)
			return
		}
		repo := ctxRepo.(*repo.Repo)
		dReq := r.Context().Value("dReq").(*request.Request)
		head, err := repo.R.Head()
		if err != nil {
			head = plumbing.NewReferenceFromStrings("", "")
		}
		dReq.Revision = path.Base(string(head.Name()))
		if dReq.Revision == "." {
			dReq.Revision = ""
		}
		h(w, r)
	}
}

func Repos(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			c = r.Context().Value("cfg").(config.Config)

			repos []*repo.Repo
			err   error
		)
		if c.ProjectListPath == "" {
			if repos, err = getRepos(c); err != nil {
				log.Println("ERROR:", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Internal server error")
				return
			}
		} else {
			projects, err := projectlist.NewProjectList(c.ProjectListPath)
			if err != nil {
				log.Println("ERROR:", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Internal server error")
				return
			}
			if len(projects) == 0 {
				log.Println("WARNING: project list empty")
			}
			repos = getFilteredRepos(c, projects)
		}
		if len(repos) == 0 {
			log.Println("WARNING: no repositories found")
		}
		newReq := r.WithContext(context.WithValue(r.Context(), "repos", repos))
		h(w, newReq)
	}
}

func Repo(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			c   = r.Context().Value("cfg").(config.Config)
			req = r.Context().Value("dReq").(*request.Request)

			repo *repo.Repo
		)
		repo = tryToOpenRepo(filepath.Join(c.RepoBasePath, req.Repo), c)
		if repo != nil {
			newReq := r.WithContext(context.WithValue(r.Context(), "repo", repo))
			h(w, newReq)
			return
		}
		if c.RemoveSuffix {
			repo = tryToOpenRepo(filepath.Join(c.RepoBasePath, req.Repo)+".git", c)
			if repo != nil {
				newReq := r.WithContext(context.WithValue(r.Context(), "repo", repo))
				h(w, newReq)
				return
			}
			repo = tryToOpenRepo(filepath.Join(c.RepoBasePath, req.Repo, "/.git"), c)
			if repo != nil {
				newReq := r.WithContext(context.WithValue(r.Context(), "repo", repo))
				h(w, newReq)
				return
			}
		}
		h(w, r)
	}
}

func tryToOpenRepo(path string, c config.Config) *repo.Repo {
	if repo.IsRepo(path) {
		r, err := repo.NewRepo(path, c)
		if err != nil {
			log.Printf("failed to open repo at %s: %v", path, err)
			return nil
		}
		return r
	}
	return nil
}

func getRepos(cfg config.Config) ([]*repo.Repo, error) {
	var rl []*repo.Repo
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("getrepolist: error accessing %s: %v", path, err)
		}
		if info.IsDir() && repo.IsRepo(path) {
			re, err := repo.NewRepo(path, cfg)
			if err != nil {
				log.Printf("failed to open repo at %s: %v", path, err)
				return filepath.SkipDir
			}
			rl = append(rl, re)
			return filepath.SkipDir
		}
		return nil
	}
	filepath.Walk(cfg.RepoBasePath, walkFunc)
	return rl, nil
}

func getFilteredRepos(cfg config.Config, pl projectlist.ProjectList) []*repo.Repo {
	var rl []*repo.Repo
	for _, project := range pl {
		testpath := filepath.Join(cfg.RepoBasePath, project)
		if repo.IsRepo(testpath) {
			re, err := repo.NewRepo(testpath, cfg)
			if err != nil {
				log.Printf("failed to open repo at %s: %v", project, err)
				continue
			}
			rl = append(rl, re)
		}
	}
	return rl
}
