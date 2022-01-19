package provider

import (
	"context"
	"path/filepath"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"

	"github.com/google/go-github/v40/github"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const gitProviderName string = "Git"

// GitProvider holds the git provider config
type GitProvider struct {
	DefaultRemoteProvider `mapstructure:",squash"`
	Git                   config.Git `mapstructure:",squash"`
	client                *github.Client
}

// CreateGitConfigProvider creates GitProvider from the git config
func CreateGitConfigProvider(git *config.Git) (RemoteProvider, error) {
	var gitProviderObj GitProvider
	err := mapstructure.Decode(git, &gitProviderObj)
	if err != nil {
		return nil, err
	}
	logger.Logger().Debug("gitProviderObj", zap.Any("gitProviderObj", gitProviderObj))
	if err = validateGitProviderConfig(gitProviderObj); err != nil {
		return nil, err
	}
	gitProviderObj.Provider = gitProviderName
	gitProviderObj.configureClient()
	return &gitProviderObj, nil
}

// configureClient configures the github client with the GitProvider
func (gitProvider *GitProvider) configureClient() {
	if gitProvider.Git.AuthRequired {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: gitProvider.Git.AccessToken},
		)
		tc := oauth2.NewClient(context.TODO(), ts)
		gitProvider.client = github.NewClient(tc)
	} else {
		gitProvider.client = github.NewClient(nil)
	}
}

// Fetch downloads the config from the github repo
func (gitProvider *GitProvider) Fetch(ctx context.Context) (*fetcher.Response, error) {
	var gitResp fetcher.Response
	opts := &github.RepositoryContentGetOptions{
		Ref: gitProvider.Git.Ref,
	}
	fileContent, _, _, err := gitProvider.client.Repositories.GetContents(ctx, gitProvider.Git.Owner, gitProvider.Git.Repo, gitProvider.Git.FilePath, opts)
	if err != nil {
		return nil, err
	}
	data, err := fileContent.GetContent()
	if err != nil {
		return nil, err
	}
	gitResp.FileData = []byte(data)
	gitResp.FileName = filepath.Base(gitProvider.Git.FilePath)
	return &gitResp, nil
}
