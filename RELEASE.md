# Releasing a version

## Prepare

1. Create a `docs/changelog-v0.5.0` branch. Substitute v0.5.0 with the version you are releasing.
2. Add an entry to CHANGELOG.md for your version, with entries for merged PRs since the previous release. Make sure to write the doc with end-user in mind.
3. Add a ToC entry at the top of CHANGELOG.md (e.g. `- [v0.5.0](#v050)`) and a compare link at the bottom of CHANGELOG.md (e.g. `[v0.5.0]: https://github.com/hbagdi/deck/compare/v0.4.0...v0.5.0`).
4. Commit and push the branch to Github. Ensure that the Markdown is rendered correctly. Make changes as needed. The link on the version itself won't resolve correctly as the tag is not yet created.
5. Open a PR against the main branch and merge the PR.

## Release

1. Pull `main` and tag `HEAD` with your version. In our example, we will tag it `v0.5.0`: `git tag v0.5.0`
2. Push the tag to remote (Github): `git push origin --tags`

As of 1.9.0, the remaining steps are automated on tag pushes to Github. They are unnecessary unless the release job fails.

1. Ensure that your `go version` is not older than indicated in the header of `go.mod` and that you have goreleaser installed.
2. Run Goreleaser: `goreleaser release --rm-dist`. This will create a release in Github and upload all the artifacts.
3. Edit the release to remove all the commit messages as the content and instead add a link to the changelog. Refer to older releases for reference.

## Homebrew release

1. Clone the [Kong/homebrew-deck](https://github.com/Kong/homebrew-deck) repo.
2. Download and unpack dist.zip from the release job artifacts.
3. `cp <unpack directory>/deck.rb <homebrew-deck directory>/Formula/`. Make sure only version and checksum is changed and rest all is left as is.
4. Commit and push to master.

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
