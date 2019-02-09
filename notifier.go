package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type cloudBuildResult struct {
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	LogURL           string           `json:"logUrl"`
	Source           source           `json:"source"`
	SourceProvenance sourceProvenance `json:"sourceProvenance"`
}

type source struct {
	RepoSource repoSource `json:"repoSource"`
}

type repoSource struct {
	RepoName   string `json:"repoName"`
	BranchName string `json:"branchName"`
}

type sourceProvenance struct {
	Repo resolvedRepoSource `json:"resolvedRepoSource"`
}

type resolvedRepoSource struct {
	CommitSHA string `json:"commitSha"`
}

type bitbucketStatusUpdate struct {
	State string `json:"state"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// CloudBuildNotifier consumes a Pub/Sub message.
func CloudBuildNotifier(ctx context.Context, m PubSubMessage) error {
	var buildResult *cloudBuildResult
	if err := json.Unmarshal(m.Data, &buildResult); err != nil {
		return err
	}

	payload, err := json.Marshal(&bitbucketStatusUpdate{
		State: formatState(buildResult.Status),
		Key:   "CLOUD-BUILD-NOTIFICATION",
		Name:  buildResult.Source.RepoSource.BranchName,
		URL:   buildResult.LogURL,
	})
	if err != nil {
		return err
	}

	splits := strings.Split(buildResult.Source.RepoSource.RepoName, "_")
	targetURL := formatURL(splits[1], splits[2], buildResult.SourceProvenance.Repo.CommitSHA)

	var username = os.Getenv("USERNAME")
	var passwd = os.Getenv("PASSWORD")
	client := &http.Client{}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payload))
	req.SetBasicAuth(username, passwd)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func formatState(state string) string {
	switch state {
	case "QUEUED":
		return "INPROGRESS"
	case "WORKING":
		return "INPROGRESS"
	case "SUCCESS":
		return "SUCCESSFUL"
	default:
		return "FAILED"
	}
}

func formatURL(ownerName string, repoName string, commitSHA string) string {
	return fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%v/%v/commit/%v/statuses/build",
		ownerName,
		repoName,
		commitSHA,
	)
}
