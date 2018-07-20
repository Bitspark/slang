package daemon

import (
	"github.com/Bitspark/go-github/github"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SlangComponentLoader struct {
	github          *github.Client
	latestRelease   *github.RepositoryRelease
	owner           string
	repo            string
	path            string
	versionFilePath string
}

func NewComponentLoader(repo string, path string) *SlangComponentLoader {
	dl := &SlangComponentLoader{
		github.NewClient(nil),
		nil,
		"Bitspark",
		repo,
		path,
		filepath.Join(path, ".VERSION"),
	}
	dl.fetchLatestRelease()
	return dl
}

func (dl *SlangComponentLoader) NewerVersionExists() bool {
	return true
}

/*
 * Downloads & unpacks latest version of a component.
 */
func (dl *SlangComponentLoader) Load() error {
	release := dl.latestRelease

	if len(release.Assets) == 0 {
		return fmt.Errorf("release '%v' needs at least 1 asset which can be downloaded", release.Name)
	}

	asset := release.Assets[0]

	compDir, err := dl.downloadArchive(*asset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	if err = dl.replaceDirContentBy(compDir); err != nil {
		return err
	}
	if err = dl.updateLocalVersionFile(); err != nil {
		return err
	}

	return nil
}

func (dl *SlangComponentLoader) fetchLatestRelease() error {
	release, _, err := dl.github.Repositories.GetLatestRelease(context.Background(), dl.owner, dl.repo)
	dl.latestRelease = release
	return err
}

func (dl *SlangComponentLoader) GetLatestReleaseVersion() string {
	return *dl.latestRelease.TagName
}

func (dl *SlangComponentLoader) GetLocalReleaseVersion() string {
	_, err := os.Stat(dl.path)

	if os.IsNotExist(err) {
		return ""
	}

	versionFile, err := os.Open(dl.versionFilePath)
	defer versionFile.Close()

	if err != nil {
		return ""
	}

	currVersion := make([]byte, 10)
	versionFile.Read(currVersion)

	return string(currVersion)
}

func (dl *SlangComponentLoader) updateLocalVersionFile() error {
	version := dl.GetLatestReleaseVersion()
	err := ioutil.WriteFile(dl.versionFilePath, []byte(version), os.ModePerm)
	return err
}

func (dl *SlangComponentLoader) downloadArchive(url string) (string, error) {
	// Download archive file
	tmpArchiveFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpArchiveFile.Name())

	if err := download(url, tmpArchiveFile); err != nil {
		return "", err
	}
	tmpArchiveFile.Close()

	// Unpack archive into directory
	tmpDstDir, err := ioutil.TempDir("", dl.repo)
	if err != nil {
		return "", err
	}

	if _, err := unzip(tmpArchiveFile.Name(), tmpDstDir); err != nil {
		return "", err
	}

	return tmpDstDir, nil
}

func (dl *SlangComponentLoader) replaceDirContentBy(newDirPath string) error {
	if _, err := os.Stat(newDirPath); err != nil {
		return err
	}

	_, err := os.Stat(dl.path);
	if !(err == nil || os.IsNotExist(err)) {
		return err
	}

	if os.IsExist(err) {
		os.RemoveAll(dl.path)
	}

	if err = os.MkdirAll(dl.path, os.ModePerm); err != nil {
		return err
	}

	if err = moveAll(newDirPath, dl.path, true); err != nil {
		return err
	}

	return err
}
