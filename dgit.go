// See LICENSE file for copyright and license details

// Package dgit provides the DGit [http.Handler] and its helpers. DGit
// is designed to be a stripped down, template-driven, fast Git viewer
// written in pure Go.
//
// The following anti-features are not implemented:
//   - Syntax highlighting
//   - Online pull/merge requests
//   - Social features (e.g., stars, followers)
//   - Users are not a concept in DGit (although admins may choose to
//     namespace repositories according to their owner)
//
// Summarizing most of the above, we may call DGit a read-only
// repository browser, or a repository viewer.
package dgit

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"sort"

	"djmo.ch/dgit/config"
	"djmo.ch/dgit/data"
	"djmo.ch/dgit/internal/convert"
	"djmo.ch/dgit/internal/middleware"
	"djmo.ch/dgit/internal/repo"
	"djmo.ch/dgit/internal/request"
)

//go:embed templates/*.tmpl
var templates embed.FS

// DGit is an [http.Handler] and can therefore be dropped into an
// [http.ServeMux]. It serves read-only pages with Git repository
// information in the following manner, where / is the root of the
// DGit [http.Handler]:
//
//   - Navigating to / serves a list of Git repositories available for
//     viewing.
//   - Navigating to /{repo} serves the tree of the HEAD ref for the
//     of {repo}. If the repository contains a README file, it's raw
//     contents are displayed below the commit tree.
//   - Navigating to /{repo}/-/refs displays a list of branches and tags
//     for repository {repo}.
//   - Navigating to /{repo}/-/tree/{ref or commit}/{path} displays
//     the tree for {ref or commit} of {repo} at {path}. If not provided,
//     {path} defaults to the root of the repository.
//   - Navigating to /{repo}/-/blob/{ref or commit}/{path} displays
//     the blob contents for {ref or commit} of {repo} at {path}. If not
//     provided, {path} defaults to the root of the repository.
//   - Navigating to /{repo}/-/commit/{commit} displays the commit
//     message and diff for commit {commit} of repository {repo}.
//   - Navigating to /{repo}/-/log/{branch} displays summary information
//     for each commit in the history of branch {branch} in repository
//     {repo}. When navigating to /{repo}/-/log, callers are redirected
//     to /{repo}/log/{default branch}.
//   - Navigating to /{repo}/-/diff displays the diff of two commits,
//     specfied as get parameters id and id2. The diff is calculated
//     assuming that id2 comes earlier, as in the git command "git
//     diff id2..id."
//
// Where the variable {commit} is used above, it may refer to a commit
// hash or ref. If the ref is a branch, the commit is the branch's
// HEAD.
type DGit struct {
	Config config.Config
}

// ServeHTTP implements the [http.Handler] interface method.
func (d *DGit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dReq, err := request.Parse(r.URL)
	ctx := context.WithValue(r.Context(), "dReq", dReq)
	req := r.WithContext(ctx)
	if err != nil {
		switch {
		case errors.Is(err, request.ErrMalformed):
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad Request: %v", err)
		case errors.Is(err, request.ErrUnknownSection):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not Found: %v", err)
		default:
			log.Print("ERROR: unexpected error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Server Error")
		}
		return
	}
	switch dReq.Section {
	case "repo":
		h := middleware.Repos(d.rootHandler, d.Config)
		h(w, req)
	case "head":
		h := middleware.Repo(d.headHandler, d.Config, dReq)
		h(w, req)
	case "tree":
		h := middleware.Repo(d.treeHandler, d.Config, dReq)
		h(w, req)
	case "blob":
		h := middleware.Repo(d.blobHandler, d.Config, dReq)
		h(w, req)
	case "refs":
		h := middleware.Repo(d.refsHandler, d.Config, dReq)
		h(w, req)
	default:
		log.Println("ERROR: Request for unknown section:", dReq.Section)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad Request")
	}
}

func (d *DGit) treeHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*repo.Repo)
	dReq := r.Context().Value("dReq").(*request.Request)
	treeData, err := convert.RepoToTreeData(repo, dReq)
	if err != nil {
		if errors.Is(err, convert.ErrTreeNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not found")
			return
		}
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "tree.tmpl", treeData)
}

func (d *DGit) logHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (d *DGit) headHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*repo.Repo)
	dReq := r.Context().Value("dReq").(*request.Request)
	head, err := repo.R.Head()
	if err != nil {
		t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
		t.ExecuteTemplate(w, "tree.tmpl", data.TreeData{
			RequestData: data.RequestData{
				Repo: repo.Slug,
			},
		})
		return
	}
	dReq.RefOrCommit = path.Base(string(head.Name()))
	treeData, err := convert.RepoToTreeData(repo, dReq)
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "tree.tmpl", treeData)
}

func (d *DGit) rootHandler(w http.ResponseWriter, r *http.Request) {
	repos := r.Context().Value("repos").([]*repo.Repo)
	sort.Sort(sort.Reverse(repo.ByLastModified(repos)))
	indexData := convert.ReposToIndexData(repos)
	t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "index.tmpl", indexData)
}

func (d *DGit) commitHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (d *DGit) diffHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (d *DGit) blobHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*repo.Repo)
	dReq := r.Context().Value("dReq").(*request.Request)
	treeData, err := convert.RepoToBlobData(repo, dReq)
	if err != nil {
		if errors.Is(err, convert.ErrBlobNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not found")
			return
		}
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "blob.tmpl", treeData)
}

func (d *DGit) refsHandler(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*repo.Repo)
	refsData, err := convert.RepoToRefsData(repo)
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
	t := template.Must(template.ParseFS(templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "refs.tmpl", refsData)
}
