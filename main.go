package main

import (
	"crypto/tls"
	"fmt"
	"github.com/certifi/gocertifi"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	gitclient "gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	dir := "temp-repo"
	clone("https://github.com/lumen/lumen", "windows", "911267b21097ea70bf2ccdfd41152313525237fb", dir)
	fmt.Printf("Cloned into %s", dir)
}

func clone(clone_url string, branch string, change string, working_dir string) bool {
	cert_pool, err := gocertifi.CACerts()
	if err != nil {
		log.Fatalf("Failed to get CA certificates: %s!", err)
	}
	customClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: cert_pool},
		},
		Timeout: 300 * time.Second,
	}
	gitclient.InstallProtocol("https", githttp.NewClient(customClient))
	gitclient.InstallProtocol("http", githttp.NewClient(customClient))

	var repo *git.Repository

	cloneOptions := git.CloneOptions{
		URL: clone_url,
	}
	cloneOptions.Tags = git.NoTags
	cloneOptions.SingleBranch = true
	cloneOptions.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	log.Println(fmt.Sprintf("Cloning %s...\n", cloneOptions.ReferenceName))

	repo, err = git.PlainClone(working_dir, false, &cloneOptions)

	if err != nil && retriableCloneError(err) {
		log.Println("Timeout while cloning! Trying again...")
		_ = os.RemoveAll(working_dir)
		EnsureFolderExists(working_dir)
		repo, err = git.PlainClone(working_dir, false, &cloneOptions)
	}

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			log.Fatalf("Failed to clone because of a timeout from Git server!")
		} else {
			log.Fatalf(fmt.Sprintf("Failed to clone: %s!", err))
		}
	}

	ref, err := repo.Head()
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to get HEAD information!"))
		return false
	}

	if ref.Hash() != plumbing.NewHash(change) {
		log.Println(fmt.Sprintf("HEAD is at %s.", ref.Hash()))
		log.Println(fmt.Sprintf("Hard resetting to %s...", change))

		workTree, err := repo.Worktree()
		if err != nil {
			log.Fatalf(fmt.Sprintf("Failed to get work tree: %s!", err))
			return false
		}

		err = workTree.Reset(&git.ResetOptions{
			Commit: plumbing.NewHash(change),
			Mode:   git.HardReset,
		})
		if err != nil {
			log.Fatalf(fmt.Sprintf("Failed to force reset to %s: %s!", change, err))
			return false
		}
	}
	log.Println(fmt.Sprintf("Checked out %s on %s branch.", change, branch))
	log.Println("Successfully cloned!")
	return true
}

func EnsureFolderExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			log.Printf("Failed to mkdir %s: %s", path, err)
		}
	}
}

func retriableCloneError(err error) bool {
	if err == nil {
		return false
	}
	errorMessage := strings.ToLower(err.Error())
	if strings.Contains(errorMessage, "timeout") {
		return true
	}
	if strings.Contains(errorMessage, "tls") {
		return true
	}
	return false
}
