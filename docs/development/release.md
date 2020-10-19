# Releasing a version


## Prepare

- Ensure your local Go version is up to date. It should match the Go version
  in `Dockerfile`.
- Create a `docs/changelog-v0.5.0` branch. Substitute v0.5.0 with the version
  you are releasing.
- Write the changelog. Ensure the links are all created correctly and
  make sure to write the doc with end-user in mind.
- Commit and push the branch to Github. Ensure that the Markdown is rendered
  correctly. Make changes as needed. The link on the version itself won't
  resolve correctly as the tag is not yet created.
- Ensure you have Goreleaser and Docker installed locally.
- Once the changelog looks good, open a PR to `main` and wait for review.

## Release

- After the changelog is merged, `git checkout main; git pull`.
- Tag the `HEAD` with your version, e.g. `git tag v0.5.0`
- Push the tag to remote (Github), e.g. `git push --tags`
- Run Goreleaser: `goreleaser release --rm-dist`. This will create
  a release in Github and upload all the artifacts.
- Edit the release to remove all the commit messages as the content and
  instead add a link to the changelog. Refer to older releases for reference.
- Clone https://github.com/Kong/homebrew-deck. Copy deck.rb from your deck repo
  folder to homebrew-deck, e.g. `cp /path/to/deck/dist/deck.rb
  /path/to/homebrew-deck/Formula/deck.rb`. `git diff` in homebrew-deck to
  confirm that only the version and checksum have changed. Commit changes with
  a "release v0.5.0" message and push master to origin.

## Docker release

Assuming you are on the TAG commit, you need to perform the following:
```
export TAG=$(git describe --abbrev=0 --tags)
export COMMIT=$(git rev-parse --short $TAG)
docker build --build-arg TAG=$TAG --build-arg COMMIT=$COMMIT -t hbagdi/deck:$TAG .
docker push hbagdi/deck:$TAG

docker tag hbagdi/deck:$TAG kong/deck:$TAG
docker push kong/deck:$TAG


# if also the latest release (not for a back-ported patch release):
docker tag hbagdi/deck:$TAG hbagdi/deck:latest
docker push hbagdi/deck:latest

docker tag hbagdi/deck:latest kong/deck:latest
docker push kong/deck:latest
```
