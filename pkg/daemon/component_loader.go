package daemon

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bitspark/go-github/github"
	"github.com/Bitspark/go-version"
)

type SlangComponentLoader struct {
	github          *github.Client
	latestRelease   *github.RepositoryRelease
	owner           string
	repo            string
	path            string
	latest          string
	versionFilePath string
}

func NewComponentLoaderLatestRelease(repo string, path string) *SlangComponentLoader {
	return newComponentLoader(repo, path, "release")

}

func NewComponentLoaderLatestMaster(repo string, path string) *SlangComponentLoader {
	return newComponentLoader(repo, path, "master")

}

func newComponentLoader(repo string, path string, latest string) *SlangComponentLoader {
	dl := &SlangComponentLoader{
		github.NewClient(nil),
		nil,
		"Bitspark",
		repo,
		path,
		latest,
		filepath.Join(path, ".VERSION"),
	}
	if latest == "release" {
		dl.fetchLatestRelease()
	}
	return dl
}

func (dl *SlangComponentLoader) NewerVersionExists() bool {
	localVersion := dl.GetLocalReleaseVersion()
	latestVersion := dl.GetLatestReleaseVersion()
	if latestVersion == nil {
		return false
	}
	return localVersion == nil || localVersion.LessThan(latestVersion)
}

/*
 * Downloads & unpacks latest version of a component.
 */
func (dl *SlangComponentLoader) Load() error {
	if dl.latest == "release" {
		release := dl.latestRelease

		if len(release.Assets) == 0 {
			return fmt.Errorf("release '%v' needs at least 1 asset which can be downloaded", release.Name)
		}

		asset := release.Assets[0]
		if err := dl.downloadArchiveAndUnpack(*asset.BrowserDownloadURL); err != nil {
			return err
		}
		if err := dl.updateLocalVersionFile(); err != nil {
			return err
		}
	} else {
		// Just download project as archive from master
		archiveURL := dl.getLatestArchiveURL()
		if err := dl.downloadArchiveAndUnpack(archiveURL); err != nil {
			return err
		}
	}
	return nil
}

func (dl *SlangComponentLoader) fetchLatestRelease() error {
	release, _, err := dl.github.Repositories.GetLatestRelease(context.Background(), dl.owner, dl.repo)
	dl.latestRelease = release
	return err
}

func (dl *SlangComponentLoader) getLatestArchiveURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%v/%v/zipball/master", dl.owner, dl.repo)
}

func (dl *SlangComponentLoader) updateLocalVersionFile() error {
	v := dl.GetLatestReleaseVersion()
	err := ioutil.WriteFile(dl.versionFilePath, []byte(v.String()), os.ModePerm)
	return err
}

func (dl *SlangComponentLoader) downloadArchiveAndUnpack(archiveURL string) error {
	archiveFilePath, err := dl.download(archiveURL)
	// Unpack archive into directory
	tmpDstDir, err := ioutil.TempDir("", dl.repo)
	if err != nil {
		return err
	}
	if _, err := unzip(archiveFilePath, tmpDstDir); err != nil {
		return err
	}
	defer os.Remove(archiveFilePath)
	defer os.RemoveAll(tmpDstDir)

	if err != nil {
		return err
	}
	if err = dl.replaceDirContentBy(tmpDstDir); err != nil {
		return err
	}
	return nil
}

func (dl *SlangComponentLoader) download(url string) (string, error) {
	// Download archive file
	tmpDstFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tmpDstFile.Close()

	if err := download(url, tmpDstFile); err != nil {
		return "", err
	}
	return tmpDstFile.Name(), nil
}

func (dl *SlangComponentLoader) replaceDirContentBy(newDirPath string) error {
	if _, err := os.Stat(newDirPath); err != nil {
		return err
	}

	_, err := os.Stat(dl.path)
	if !(err == nil || os.IsNotExist(err)) {
		return err
	}

	if os.IsExist(err) {
		os.RemoveAll(dl.path)
	}

	if err = copyAll(newDirPath, dl.path, true); err != nil {
		return err
	}

	return err
}

func (dl *SlangComponentLoader) GetLatestReleaseVersion() *version.Version {
	if dl.latestRelease == nil {
		return nil
	}
	return toVersion(*dl.latestRelease.TagName)
}

func (dl *SlangComponentLoader) GetLocalReleaseVersion() *version.Version {
	_, err := os.Stat(dl.path)

	if os.IsNotExist(err) {
		return nil
	}

	versionFile, err := os.Open(dl.versionFilePath)
	defer versionFile.Close()

	if err != nil {
		return nil
	}

	versionReader := bufio.NewReader(versionFile)
	currVersion, _, err := versionReader.ReadLine()

	if err != nil {
		return nil
	}

	return toVersion(strings.TrimSpace(string(currVersion)))
}

func IsNewestSlangVersion(myVerStr string) (bool, string, error) {
	myVer := toVersion(myVerStr)
	release, _, err := github.NewClient(nil).Repositories.GetLatestRelease(context.Background(), "Bitspark", "slang")
	if err != nil {
		return false, "", err
	}
	currVer := toVersion(*release.TagName)
	if myVer.LessThan(currVer) {
		return false, *release.TagName, nil
	}
	return true, *release.TagName, nil
}
