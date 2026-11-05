# Releasing a version

## Prepare

1. Create a `docs/changelog-v0.5.0` branch. Substitute v0.5.0 with the version you are releasing.
1. In `README.md` update the installation instruction (Linux and Windows) to have the new version in the `curl` command.
1. Add an entry to CHANGELOG.md for your version, with entries for merged PRs since the previous release. Make sure to write the doc with end-user in mind.
1. Add a ToC entry at the top of CHANGELOG.md (e.g. `- [v0.5.0](#v050)`) and a compare link at the bottom of CHANGELOG.md (e.g. `[v0.5.0]: https://github.com/hbagdi/deck/compare/v0.4.0...v0.5.0`).
1. Commit and push the branch to Github. Ensure that the Markdown is rendered correctly. Make changes as needed. The link on the version itself won't resolve correctly as the tag is not yet created.
1. Open a PR against the main branch and merge the PR.

## Release

1. Pull `main` and tag `HEAD` with your version. In our example, we will tag it `v0.5.0`: `git tag v0.5.0`
1. Push the tag to remote (Github): `git push origin --tags`

As of 1.9.0, the remaining steps are automated on tag pushes to Github. They are unnecessary unless the release job fails.

1. Ensure that your `go version` is not older than indicated in the header of `go.mod` and that you have goreleaser installed.
1. Run Goreleaser: `goreleaser release --rm-dist`. This will create a release in Github and upload all the artifacts.
1. Edit the release to remove all the commit messages as the content and instead add a link to the changelog. Refer to older releases for reference.

## Homebrew release

1. Clone the [Kong/homebrew-deck](https://github.com/Kong/homebrew-deck) repo.
1. Download and unpack `dist.zip` from the [release job artifacts](https://github.com/Kong/deck/actions/workflows/release.yaml).
1. `cp <unpack directory>/deck.rb <homebrew-deck directory>/Formula/`. Make sure only version and checksum is changed and rest all is left as is.
1. Commit and push to master.

## winget release

GoReleaser generates the [winget](https://learn.microsoft.com/windows/package-manager/) manifests for `Kong.decK`.
The `winget` job of the release workflow submits them to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) with [wingetcreate](https://github.com/microsoft/winget-create).

- The job needs the `WINGET_GITHUB_TOKEN` secret: a classic personal access token with the `public_repo` scope. If the secret is not set, the job submits nothing and passes.
- wingetcreate forks microsoft/winget-pkgs into the account that owns the token, if that fork does not exist.
- If the job fails, for example because the token expired, the release is not affected. Replace the token and re-run the `winget` job.
- The job does not submit pre-release tags.
- A Microsoft moderator reviews each pull request before the version is available in winget.

## Docker release

As of 1.9.0, these steps are automated on tag pushes to Github. They are unnecessary unless the release job fails.

Assuming you are on the TAG commit, you need to perform the following:

```
export TAG=$(git describe --abbrev=0 --tags)
export COMMIT=$(git rev-parse --short $TAG)
docker build --build-arg TAG=$TAG --build-arg COMMIT=$COMMIT -t kong/deck:$TAG .
docker push kong/deck:$TAG

# if also the latest release (not for a back-ported patch release):
docker tag kong/deck:$TAG kong/deck:latest
docker push kong/deck:latest
```
