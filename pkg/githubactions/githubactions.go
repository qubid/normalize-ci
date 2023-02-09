package githubactions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cidverse/normalizeci/pkg/ncispec"
	"github.com/cidverse/normalizeci/pkg/projectdetails"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/gosimple/slug"

	"github.com/cidverse/normalizeci/pkg/common"
)

// Normalizer is the implementation of the normalizer
type Normalizer struct {
	version string
	name    string
	slug    string
}

// GetName returns the name of the normalizer
func (n Normalizer) GetName() string {
	return n.name
}

// GetSlug returns the slug of the normalizer
func (n Normalizer) GetSlug() string {
	return n.slug
}

// Check if this package can handle the current environment
func (n Normalizer) Check(env map[string]string) bool {
	if env["GITHUB_ACTIONS"] == "true" {
		return true
	}

	return false
}

// Normalize normalizes the environment variables into the common format
func (n Normalizer) Normalize(env map[string]string) map[string]string {
	nci := ncispec.OfMap(env)

	// common
	nci.NCI = "true"
	nci.NCI_VERSION = n.version
	nci.NCI_SERVICE_NAME = n.name
	nci.NCI_SERVICE_SLUG = n.slug

	// worker
	nci.NCI_WORKER_ID = env["RUNNER_TRACKING_ID"]
	nci.NCI_WORKER_NAME = env["RUNNER_TRACKING_ID"]
	nci.NCI_WORKER_VERSION = env["ImageVersion"]
	nci.NCI_WORKER_ARCH = runtime.GOOS + "/" + runtime.GOARCH

	// pipeline
	pipelineEvent := env["GITHUB_EVENT_NAME"]
	switch pipelineEvent {
	case "push":
		nci.NCI_PIPELINE_TRIGGER = string(ncispec.PipelineTriggerPush)
	case "pull_request":
		nci.NCI_PIPELINE_TRIGGER = string(ncispec.PipelineTriggerPullRequest)
	default:
		nci.NCI_PIPELINE_TRIGGER = string(ncispec.PipelineTriggerUnknown)
	}
	if nci.NCI_PIPELINE_TRIGGER == string(ncispec.PipelineTriggerPullRequest) {
		// PR
		splitRef := strings.Split(env["GITHUB_REF"], "/")
		nci.NCI_PIPELINE_PULL_REQUEST_ID = splitRef[2]
	}
	nci.NCI_PIPELINE_STAGE_NAME = env["GITHUB_WORKFLOW"]
	nci.NCI_PIPELINE_STAGE_SLUG = slug.Make(env["GITHUB_WORKFLOW"])
	nci.NCI_PIPELINE_JOB_NAME = env["GITHUB_ACTION"]
	nci.NCI_PIPELINE_JOB_SLUG = slug.Make(env["GITHUB_ACTION"])
	nci.NCI_PIPELINE_URL = fmt.Sprintf("%s/%s/actions/runs/%s", env["GITHUB_SERVER_URL"], env["GITHUB_REPOSITORY"], env["GITHUB_RUN_ID"])

	// repository
	projectDir := vcsrepository.FindRepositoryDirectory(common.GetWorkingDirectory())
	vcsData, addDataErr := vcsrepository.GetVCSRepositoryInformation(projectDir)
	if addDataErr != nil {
		panic(addDataErr)
	}
	nci.NCI_REPOSITORY_KIND = vcsData[ncispec.NCI_REPOSITORY_KIND]
	nci.NCI_REPOSITORY_REMOTE = vcsData[ncispec.NCI_REPOSITORY_REMOTE]
	nci.NCI_REPOSITORY_STATUS = vcsData[ncispec.NCI_REPOSITORY_STATUS]
	nci.NCI_COMMIT_REF_TYPE = vcsData[ncispec.NCI_COMMIT_REF_TYPE]
	nci.NCI_COMMIT_REF_NAME = vcsData[ncispec.NCI_COMMIT_REF_NAME]
	nci.NCI_COMMIT_REF_PATH = vcsData[ncispec.NCI_COMMIT_REF_PATH]
	nci.NCI_COMMIT_REF_SLUG = vcsData[ncispec.NCI_COMMIT_REF_SLUG]
	nci.NCI_COMMIT_REF_VCS = vcsData[ncispec.NCI_COMMIT_REF_VCS]
	nci.NCI_COMMIT_REF_RELEASE = vcsData[ncispec.NCI_COMMIT_REF_RELEASE]
	nci.NCI_COMMIT_SHA = vcsData[ncispec.NCI_COMMIT_SHA]
	nci.NCI_COMMIT_SHA_SHORT = vcsData[ncispec.NCI_COMMIT_SHA_SHORT]
	nci.NCI_COMMIT_TITLE = vcsData[ncispec.NCI_COMMIT_TITLE]
	nci.NCI_COMMIT_DESCRIPTION = vcsData[ncispec.NCI_COMMIT_DESCRIPTION]
	nci.NCI_COMMIT_AUTHOR_NAME = vcsData[ncispec.NCI_COMMIT_AUTHOR_NAME]
	nci.NCI_COMMIT_AUTHOR_EMAIL = vcsData[ncispec.NCI_COMMIT_AUTHOR_EMAIL]
	nci.NCI_COMMIT_COMMITTER_NAME = vcsData[ncispec.NCI_COMMIT_COMMITTER_NAME]
	nci.NCI_COMMIT_COMMITTER_EMAIL = vcsData[ncispec.NCI_COMMIT_COMMITTER_EMAIL]
	nci.NCI_COMMIT_COUNT = vcsData[ncispec.NCI_COMMIT_COUNT]
	nci.NCI_LASTRELEASE_REF_NAME = vcsData[ncispec.NCI_LASTRELEASE_REF_NAME]
	nci.NCI_LASTRELEASE_REF_SLUG = vcsData[ncispec.NCI_LASTRELEASE_REF_SLUG]
	nci.NCI_LASTRELEASE_REF_VCS = vcsData[ncispec.NCI_LASTRELEASE_REF_VCS]
	nci.NCI_LASTRELEASE_COMMIT_AFTER_COUNT = vcsData[ncispec.NCI_LASTRELEASE_COMMIT_AFTER_COUNT]

	// project details
	projectData := projectdetails.GetProjectDetails(nci.NCI_REPOSITORY_KIND, nci.NCI_REPOSITORY_REMOTE)
	if projectData != nil {
		nci.NCI_PROJECT_ID = projectData[ncispec.NCI_PROJECT_ID]
		nci.NCI_PROJECT_NAME = projectData[ncispec.NCI_PROJECT_NAME]
		nci.NCI_PROJECT_PATH = projectData[ncispec.NCI_PROJECT_PATH]
		nci.NCI_PROJECT_SLUG = projectData[ncispec.NCI_PROJECT_SLUG]
		nci.NCI_PROJECT_DESCRIPTION = projectData[ncispec.NCI_PROJECT_DESCRIPTION]
		nci.NCI_PROJECT_TOPICS = projectData[ncispec.NCI_PROJECT_TOPICS]
		nci.NCI_PROJECT_ISSUE_URL = projectData[ncispec.NCI_PROJECT_ISSUE_URL]
		nci.NCI_PROJECT_STARGAZERS = projectData[ncispec.NCI_PROJECT_STARGAZERS]
		nci.NCI_PROJECT_FORKS = projectData[ncispec.NCI_PROJECT_FORKS]
	}
	nci.NCI_PROJECT_DIR = projectDir

	// container registry
	nci.NCI_CONTAINERREGISTRY_HOST = ""
	nci.NCI_CONTAINERREGISTRY_REPOSITORY = slug.Make(common.GetDirectoryNameFromPath(filepath.Join(vcsrepository.FindRepositoryDirectory(common.GetWorkingDirectory())+string(os.PathSeparator), "file")))
	nci.NCI_CONTAINERREGISTRY_USERNAME = ""
	nci.NCI_CONTAINERREGISTRY_PASSWORD = ""
	nci.NCI_CONTAINERREGISTRY_TAG = nci.NCI_COMMIT_REF_RELEASE

	// control
	nci.NCI_DEPLOY_FREEZE = "false"

	return ncispec.ToMap(nci)
}

func (n Normalizer) Denormalize(env map[string]string) map[string]string {
	return make(map[string]string)
}

// NewNormalizer gets a instance of the normalizer
func NewNormalizer() Normalizer {
	entity := Normalizer{
		version: "0.3.0",
		name:    "GitHub Actions",
		slug:    "github-actions",
	}

	return entity
}
