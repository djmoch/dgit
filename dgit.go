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
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"

	"djmo.ch/dgit/config"
	"djmo.ch/dgit/data"
	"djmo.ch/dgit/internal/convert"
	"djmo.ch/dgit/internal/middleware"
	"djmo.ch/dgit/internal/repo"
	"djmo.ch/dgit/internal/request"
	"github.com/dustin/go-humanize"
)

var funcMap = template.FuncMap{"Humanize": humanize.Time}

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
	if err != nil {
		switch {
		case errors.Is(err, request.ErrMalformed):
			log.Println("ERROR: bad request:", err)
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

	ctx := context.WithValue(r.Context(), "dReq", dReq)
	ctx = context.WithValue(ctx, "cfg", d.Config)
	req := r.WithContext(ctx)

	switch dReq.Section {
	case "repo":
		h := middleware.Get(middleware.Repos(d.rootHandler))
		h(w, req)
	case "head":
		h := middleware.Get(middleware.Repo(middleware.ResolveHead(d.treeHandler)))
		h(w, req)
	case "tree":
		h := middleware.Get(middleware.Repo(d.treeHandler))
		h(w, req)
	case "blob":
		h := middleware.Get(middleware.Repo(d.blobHandler))
		h(w, req)
	case "refs":
		h := middleware.Get(middleware.Repo(d.refsHandler))
		h(w, req)
	case "log":
		h := middleware.Get(middleware.Repo(d.logHandler))
		h(w, req)
	case "commit":
		h := middleware.Get(middleware.Repo(d.commitHandler))
		h(w, req)
	case "diff":
		h := middleware.Get(middleware.Repo(d.diffHandler))
		h(w, req)
	default:
		log.Println("ERROR: Request for unknown section:", dReq.Section)
		w.WriteHeader(http.StatusBadRequest)
		d.displayError(w, "Bad Request")
	}
}

func (d *DGit) treeHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	dReq := r.Context().Value("dReq").(*request.Request)
	if dReq.Revision == "" {
		t := template.Must(template.New("templates").Funcs(funcMap).
			ParseFS(d.Config.Templates, "templates/*.tmpl"))
		t.ExecuteTemplate(w, "tree.tmpl", data.TreeData{
			RequestData: data.RequestData{
				Repo: data.Repo{Slug: repo.Slug},
			},
		})
		return
	}
	treeData, err := convert.ToTreeData(repo, dReq)
	if err != nil {
		if errors.Is(err, convert.ErrDirectoryNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			d.displayError(w, "Not found")
			return
		}
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "tree.tmpl", treeData)
}

func (d *DGit) logHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	dReq := r.Context().Value("dReq").(*request.Request)
	logData, err := convert.ToLogData(repo, dReq)
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "log.tmpl", logData)
}

func (d *DGit) rootHandler(w http.ResponseWriter, r *http.Request) {
	repos := r.Context().Value("repos").([]*repo.Repo)
	sort.Sort(sort.Reverse(repo.ByLastModified(repos)))
	indexData := convert.ToIndexData(repos)
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "index.tmpl", indexData)
}

func (d *DGit) commitHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	dReq := r.Context().Value("dReq").(*request.Request)
	commitData, err := convert.ToCommitData(repo, dReq)
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "commit.tmpl", commitData)
}

func (d *DGit) diffHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	dReq := r.Context().Value("dReq").(*request.Request)
	diffData, err := convert.ToDiffData(repo, dReq)
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "diff.tmpl", diffData)
}

func (d *DGit) blobHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	dReq := r.Context().Value("dReq").(*request.Request)
	treeData, err := convert.ToBlobData(repo, dReq)
	if err != nil {
		if errors.Is(err, convert.ErrFileNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			d.displayError(w, "Not found")
			return
		}
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "blob.tmpl", treeData)
}

func (d *DGit) refsHandler(w http.ResponseWriter, r *http.Request) {
	repo := getRepo(r)
	if repo == nil {
		w.WriteHeader(http.StatusNotFound)
		d.displayError(w, "Repo not found")
		return
	}
	refsData, err := convert.ToRefsData(repo)
	sort.Sort(sort.Reverse(convert.ByAge(refsData.Branches)))
	sort.Sort(sort.Reverse(convert.ByAge(refsData.Tags)))
	if err != nil {
		log.Printf("ERROR: failed to extract template data from %s: %v", repo.Slug, err)
		w.WriteHeader(http.StatusInternalServerError)
		d.displayError(w, "Internal server error")
		return
	}
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "refs.tmpl", refsData)
}

func (d *DGit) displayError(w http.ResponseWriter, msg string) {
	t := template.Must(template.New("templates").Funcs(funcMap).
		ParseFS(d.Config.Templates, "templates/*.tmpl"))
	t.ExecuteTemplate(w, "error.tmpl", struct{ Message string }{Message: msg})
}

func getRepo(r *http.Request) *repo.Repo {
	ctxRepo := r.Context().Value("repo")
	if ctxRepo == nil {
		return nil
	}
	re := ctxRepo.(*repo.Repo)
	return re
}
